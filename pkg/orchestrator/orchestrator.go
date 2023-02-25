package orchestrator

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arnarg/mads/pkg/entities"
	"github.com/arnarg/mads/pkg/podman"
	"github.com/arnarg/mads/pkg/podman/containers"
	"github.com/arnarg/mads/pkg/podman/images"
	"github.com/arnarg/mads/pkg/podman/pods"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/connect/envoy"
	"github.com/mitchellh/mapstructure"
)

const (
	lastAppliedLabel = "mads/last-applied-configuration"
	envoyImage       = "envoyproxy/envoy:v1.22.8"
)

type consulInfo struct {
	DebugConfig struct {
		GRPCAddrs    []string
		GRPCTLSAddrs []string
	}
}

type Config struct {
	PodmanSocketPath string
}

type Orchestrator struct {
	pclient  *podman.Client
	cclient  *api.Client
	grpcAddr string
	grpcPort string
	grpcTLS  bool
}

func NewOrchestrator(cfg *Config) (*Orchestrator, error) {
	// Create a podman client
	pclient := podman.NewClient(&podman.Config{SocketPath: cfg.PodmanSocketPath})

	// Create a consul client
	cclient, err := api.NewClient(&api.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create consul client: %s", err)
	}

	// Get consul agent info
	cinfo, err := cclient.Agent().Self()
	if err != nil {
		return nil, fmt.Errorf("could not get agent info: %s", err)
	}

	// map to struct
	info := &consulInfo{}
	err = mapstructure.Decode(cinfo, info)
	if err != nil {
		return nil, err
	}

	// Find a valid address to use
	addr, port, tls, err := findGRPCAddrPort(info)
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		pclient:  pclient,
		cclient:  cclient,
		grpcAddr: addr,
		grpcPort: port,
		grpcTLS:  tls,
	}, nil
}

func (o *Orchestrator) Apply(ctx context.Context, pod *entities.Pod) error {
	// Create services
	for _, svc := range pod.Services {
		ctr, err := o.createService(ctx, &svc)
		if err != nil {
			return err
		}

		// Add sidecar container to pod
		if ctr != nil {
			pod.Containers = append(pod.Containers, *ctr)
		}
	}

	// Compute hash for current configuration
	currHash, err := pod.Hash()
	if err != nil {
		return fmt.Errorf("could not compute hash for pod '%s': %s", pod.Name, err)
	}

	// Check if a pod with same name exists
	exists, id, err := o.pclient.Pods().Exists(ctx, pod.Name)
	if err != nil {
		return fmt.Errorf("could not check if pod exists: %s", err)
	}

	// Check that applied hash is the same
	if exists {
		// Get pod info
		info, err := o.pclient.Pods().Inspect(ctx, id)
		if err != nil {
			return err
		}

		// Get last applied hash
		lastHash, ok := info.Labels[lastAppliedLabel]

		// No hash is present, we refuse to apply
		if !ok {
			return fmt.Errorf("pod '%s' has no mads label, will not apply", pod.Name)
		}

		// last applied hash is different from current configuration so we delete the pod
		if lastHash != currHash {
			err := o.pclient.Pods().Delete(ctx, id, true)
			if err != nil {
				return fmt.Errorf("could not delete old pod '%s': %s", id, err)
			}

			// Set new state
			exists = false
			id = ""
		}
	}

	// Create pod
	if !exists {
		// Create pod creation request
		req := &pods.PodCreateRequest{
			Name:   pod.Name,
			Labels: map[string]string{lastAppliedLabel: currHash},
		}

		// Apply port mappings from all containers
		for _, ctr := range pod.Containers {
			for _, mapping := range ctr.Ports {
				req.PortMappings = append(req.PortMappings, pods.PodPortMapping{
					HostIP:        mapping.HostIP,
					HostPort:      mapping.HostPort,
					ContainerPort: mapping.ContainerPort,
					Protocol:      mapping.Protocol,
				})
			}
		}

		// Create pod
		id, err = o.pclient.Pods().Create(ctx, req)
		if err != nil {
			return fmt.Errorf("could not create pod '%s': %s", pod.Name, err)
		}

		// Since we just created a new pod we need to create all of its containers
		for _, ctr := range pod.Containers {
			// Create container
			ctrName := fmt.Sprintf("%s-%s", pod.Name, ctr.Name)
			err := o.createContainer(ctx, ctrName, id, &ctr)
			if err != nil {
				// TODO: delete pod to cleanup
				return fmt.Errorf("could not create container '%s' in pod '%s': %s", ctr.Name, pod.Name, err)
			}
		}
	}

	// Get pod info
	info, err := o.pclient.Pods().Inspect(ctx, pod.Name)
	if err != nil {
		return fmt.Errorf("could not get info for pod '%s': %s", pod.Name, err)
	}

	if info.State != pods.PodStateRunning {
		err := o.pclient.Pods().Start(ctx, pod.Name)
		if err != nil && err != pods.ErrPodAlreadyStarted {
			return err
		}
	}

	return nil
}

func (o *Orchestrator) createService(ctx context.Context, svc *entities.Service) (*entities.Container, error) {
	// Create a service registration
	csvc := &api.AgentServiceRegistration{
		Name: svc.Name,
		Tags: svc.Tags,
		Port: svc.Port,
		Connect: &api.AgentServiceConnect{
			Native: svc.Connect.Native,
		},
	}

	// Add connect sidecar config if applicable
	if !csvc.Connect.Native && svc.Connect.SidecarService != nil {
		csvc.Connect.SidecarService = &api.AgentServiceRegistration{}

		// TODO: setup proxy config
	}

	// Register service
	err := o.cclient.Agent().ServiceRegister(csvc)
	if err != nil {
		return nil, err
	}

	// Get service metadata
	sidecarName := fmt.Sprintf("%s-sidecar-proxy", svc.Name)
	service, _, err := o.cclient.Agent().Service(sidecarName, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}

	// Check if service sidecar container needs to be created
	if service != nil {
		// Get prometheus bind addr for envoy
		adminAddr := "0.0.0.0"
		adminPort := 9100
		if promAddr, ok := service.Proxy.Config["envoy_prometheus_bind_addr"]; ok {
			split := strings.Split(promAddr.(string), ":")
			if len(split) > 1 {
				p, err := strconv.Atoi(split[1])
				if err == nil {
					adminPort = p
				}
			}
		}

		// Render envoy config for sidecar proxy
		// TODO: not depend on internal consul CLI util
		bcfg := &envoy.BootstrapConfig{}
		ecfg, err := bcfg.GenerateJSON(
			&envoy.BootstrapTplArgs{
				ProxyCluster:          svc.Name,
				ProxySourceService:    svc.Name,
				ProxyID:               sidecarName,
				AdminAccessLogPath:    "/dev/null",
				AdminBindAddress:      adminAddr,
				AdminBindPort:         strconv.FormatInt(int64(adminPort), 10),
				LocalAgentClusterName: "local_agent",
				Token:                 "",
				GRPC: envoy.GRPC{
					AgentAddress: o.grpcAddr,
					AgentPort:    o.grpcPort,
					AgentTLS:     o.grpcTLS,
				},
			},
			true,
		)
		if err != nil {
			return nil, err
		}

		// Return a container that should be added to pod
		return &entities.Container{
			Name:            sidecarName,
			Image:           envoyImage,
			ImagePullPolicy: images.PullPolicyMissing,
			RestartPolicy:   containers.RestartPolicyAlways,
			Args:            []string{"-c", "/etc/envoy/envoy.json"},
			Ports: []entities.ContainerPortMapping{
				{
					HostPort:      uint16(service.Port),
					ContainerPort: uint16(service.Port),
					Protocol:      "tcp",
				},
				{
					HostPort:      uint16(adminPort),
					ContainerPort: uint16(adminPort),
					Protocol:      "tcp",
				},
			},
			Files: []entities.ContainerFile{
				{
					Destination: "/etc/envoy/envoy.json",
					Content:     string(ecfg),
					Mode:        0644,
				},
			},
		}, nil
	}

	return nil, nil
}

func (o *Orchestrator) createContainer(ctx context.Context, name, podID string, ctr *entities.Container) error {
	// Get image
	imageID, err := realizeImage(ctx, o.pclient, ctr.Image, ctr.ImagePullPolicy)
	if err != nil {
		return err
	}

	// Create container creation request
	req := &containers.ContainerCreateRequest{
		Name:    name,
		Image:   imageID,
		Pod:     podID,
		Command: ctr.Args,
	}

	// Apply mounts
	for _, mount := range ctr.Mounts {
		req.Mounts = append(req.Mounts, containers.ContainerMount{
			Type:        mount.Type,
			Destination: mount.Destination,
			Source:      mount.Source,
			Options:     mount.Options,
		})
	}

	// Create container
	err = o.pclient.Containers().Create(ctx, req)
	if err != nil {
		return err
	}

	// Write files to tar archive
	if len(ctr.Files) > 0 {
		buf := &bytes.Buffer{}

		// Write tar archive into buffer
		err := writeTarArchive(ctx, buf, ctr.Files)
		if err != nil {
			return err
		}

		// Copy tar archive buffer into container
		err = o.pclient.Containers().Copy(ctx, name, buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeTarArchive(ctx context.Context, w io.Writer, files []entities.ContainerFile) error {
	// Create a tar writer
	tw := tar.NewWriter(w)
	defer tw.Close()

	// Write each file to the tar archive
	for _, f := range files {
		buf := []byte(f.Content)

		// Set default mode
		mode := f.Mode
		if mode == 0 {
			mode = 0644
		}

		// Create header for file
		hdr := tar.Header{
			Format:  tar.FormatGNU,
			Name:    strings.TrimPrefix(f.Destination, "/"),
			Size:    int64(len(buf)),
			Mode:    mode,
			ModTime: time.Now(),
		}

		// Write header to tar
		err := tw.WriteHeader(&hdr)
		if err != nil {
			return err
		}

		// Write content to tar
		_, err = tw.Write(buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func realizeImage(ctx context.Context, client *podman.Client, rawImage, pullPolicy string) (string, error) {
	archivePrefixRegex := regexp.MustCompile(`^(?:docker|oci)-archive\:`)

	var info *images.ImageInfo

	// Check if it's a local archive image
	if archivePrefixRegex.MatchString(rawImage) {
		// Remove the archive prefix
		fpath := archivePrefixRegex.ReplaceAllString(rawImage, "")

		// Open file for reading
		imagef, err := os.Open(fpath)
		if err != nil {
			return "", fmt.Errorf("could not open archive image file for reading: %s", err)
		}
		defer imagef.Close()

		// Load image into podman
		info, err = client.Images().Load(ctx, imagef)
		if err != nil {
			return "", fmt.Errorf("could not load archive image: %s", err)
		}
	} else {
		// We try to pull the image instead
		iinfo, err := client.Images().Pull(ctx, rawImage, &images.PullOptions{Policy: pullPolicy})
		if err != nil {
			return "", fmt.Errorf("could not pull image '%s': %s", rawImage, err)
		}
		info = iinfo
	}

	return info.Id, nil
}

func findGRPCAddrPort(info *consulInfo) (string, string, bool, error) {
	// First look in TLS addrs
	for _, addr := range info.DebugConfig.GRPCTLSAddrs {
		if strings.HasPrefix(addr, "tcp://") {
			noPrefix := strings.TrimPrefix(addr, "tcp://")
			split := strings.Split(noPrefix, ":")

			if len(split) < 2 {
				continue
			}

			address := split[0]
			port := split[1]

			if a := net.ParseIP(address); a != nil && !a.IsLoopback() {
				return address, port, true, nil
			}
		}
	}

	// Then look in non TLS addres
	for _, addr := range info.DebugConfig.GRPCAddrs {
		if strings.HasPrefix(addr, "tcp://") {
			noPrefix := strings.TrimPrefix(addr, "tcp://")
			split := strings.Split(noPrefix, ":")

			if len(split) < 2 {
				continue
			}

			address := split[0]
			port := split[1]

			if a := net.ParseIP(address); a != nil && !a.IsLoopback() {
				return address, port, false, nil
			}
		}
	}

	return "", "", false, fmt.Errorf("no valid grpc address found")
}