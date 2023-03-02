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
	"github.com/arnarg/mads/pkg/envoy"
	"github.com/arnarg/mads/pkg/podman"
	"github.com/arnarg/mads/pkg/podman/containers"
	"github.com/arnarg/mads/pkg/podman/images"
	"github.com/arnarg/mads/pkg/podman/pods"
	"github.com/hashicorp/consul/api"
	"github.com/mitchellh/mapstructure"
)

const (
	lastAppliedLabel   = "mads/last-applied-configuration"
	serviceIDsLabel    = "mads/service-ids"
	managedServiceMeta = "mads_managed"
	servicePodNameMeta = "mads_pod_name"
)

type consulInfo struct {
	DebugConfig struct {
		GRPCAddrs    []string
		GRPCTLSAddrs []string
	}
}

type Config struct {
	PodmanSocketPath string
	EnvoyImage       string
}

type Orchestrator struct {
	pclient    *podman.Client
	cclient    *api.Client
	envoyImage string
	grpcAddr   string
	grpcPort   uint16
	grpcTLS    bool
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
		pclient:    pclient,
		cclient:    cclient,
		envoyImage: cfg.EnvoyImage,
		grpcAddr:   addr,
		grpcPort:   port,
		grpcTLS:    tls,
	}, nil
}

func (o *Orchestrator) Delete(ctx context.Context, nameOrID string) error {
	// Try to get pod info from podman
	pinfo, err := o.pclient.Pods().Inspect(ctx, nameOrID)
	if err != nil {
		return fmt.Errorf("could not get info on pod '%s': %s", nameOrID, err)
	}

	// Check if pod is mads managed.
	// Simply the presence of the last-applied-configuration label is enough.
	if _, ok := pinfo.Labels[lastAppliedLabel]; !ok {
		return fmt.Errorf("pod '%s' is not managed by mads", nameOrID)
	}

	// Get a list of services to clean up
	svcs := []string{}
	if svcList, ok := pinfo.Labels[serviceIDsLabel]; ok {
		svcs = strings.Split(svcList, ",")
	}

	// Deregister services
	for _, svc := range svcs {
		err := o.cclient.Agent().ServiceDeregister(svc)
		// If we get a 404 we might have already deregistered it in a previous run
		// but it doesn't matter and we'll just continue.
		if err != nil && !strings.Contains(err.Error(), "Unknown service ID") {
			return fmt.Errorf("could not deregister consul service '%s': %s", svc, err)
		}
	}

	// Delete podman pod.
	// We have confirmed that the pod has the last-applied-configuration label so we can just force delete it.
	err = o.pclient.Pods().Delete(ctx, nameOrID, true)
	if err != nil {
		return fmt.Errorf("could not delete pod '%s': %s", nameOrID, err)
	}

	return nil
}

func (o *Orchestrator) Apply(ctx context.Context, pod *entities.Pod) error {
	// Create services
	svcIDs := []string{}
	for _, svc := range pod.Services {
		id, ctr, err := o.createService(ctx, pod.Name, &svc)
		if err != nil {
			return err
		}

		// Add sidecar container to pod
		if ctr != nil {
			pod.Containers = append(pod.Containers, *ctr)
		}

		// Add to list of services
		svcIDs = append(svcIDs, id)
	}

	// Add service IDs to pod labels
	podLabels := map[string]string{
		serviceIDsLabel: strings.Join(svcIDs, ","),
	}
	for k, v := range pod.Labels {
		podLabels[k] = v
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
		// Save hash label
		podLabels[lastAppliedLabel] = currHash

		// Create pod creation request
		req := &pods.PodCreateRequest{
			Name:   pod.Name,
			Labels: podLabels,
		}

		// Apply hosts
		for host, ip := range pod.Hosts {
			req.HostAdd = append(req.HostAdd, fmt.Sprintf("%s:%s", host, ip))
		}

		// Take various config from containers that needs to be set on pod level
		for _, ctr := range pod.Containers {
			// Apply port mappings from all containers
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
				// Delete pod to cleanup (best effort)
				o.pclient.Pods().Delete(ctx, id, true)

				return fmt.Errorf("could not create container '%s' in pod '%s': %s", ctr.Name, pod.Name, err)
			}
		}
	}

	// Get pod info
	info, err := o.pclient.Pods().Inspect(ctx, pod.Name)
	if err != nil {
		return fmt.Errorf("could not get info for pod '%s': %s", pod.Name, err)
	}

	// Start pod
	if info.State != pods.PodStateRunning {
		err := o.pclient.Pods().Start(ctx, pod.Name)
		if err != nil && err != pods.ErrPodAlreadyStarted {
			return err
		}
	}

	return nil
}

func (o *Orchestrator) createService(ctx context.Context, podName string, svc *entities.Service) (string, *entities.Container, error) {
	// Create a service registration
	csvc := &api.AgentServiceRegistration{
		ID:   fmt.Sprintf("mads-pod-%s-%s", podName, svc.Name),
		Name: svc.Name,
		Tags: svc.Tags,
		Port: svc.Port,
		Meta: map[string]string{
			managedServiceMeta: "true",
			servicePodNameMeta: podName,
		},
		Connect: &api.AgentServiceConnect{
			Native: svc.Connect.Native,
		},
	}

	// Add connect sidecar config if applicable
	if !csvc.Connect.Native && svc.Connect.SidecarService != nil {
		csvc.Connect.SidecarService = &api.AgentServiceRegistration{}

		if svc.Connect.SidecarService.Proxy != nil {
		// Setup proxy config
		var proxyCfg *api.AgentServiceConnectProxyConfig

		// Add upstreams to service registration
		if len(svc.Connect.SidecarService.Proxy.Upstreams) > 0 {
			proxyCfg = &api.AgentServiceConnectProxyConfig{Mode: api.ProxyModeTransparent}

			for _, upstream := range svc.Connect.SidecarService.Proxy.Upstreams {
				proxyCfg.Upstreams = append(proxyCfg.Upstreams, api.Upstream{
					LocalBindAddress: upstream.LocalBindAddress,
					LocalBindPort:    int(upstream.LocalBindPort),
					DestinationName:  upstream.DestinationName,
				})
			}
		}

		// Add expose paths to proxy config
		if len(svc.Connect.SidecarService.Proxy.Expose.Paths) > 0 {
			if proxyCfg == nil {
				proxyCfg = &api.AgentServiceConnectProxyConfig{Mode: api.ProxyModeTransparent}
			}

			for _, expose := range svc.Connect.SidecarService.Proxy.Expose.Paths {
				proxyCfg.Expose.Paths = append(proxyCfg.Expose.Paths, api.ExposePath{
					Path:          expose.Path,
					LocalPathPort: int(expose.LocalPathPort),
					ListenerPort:  int(expose.ListenerPort),
					Protocol:      expose.Protocol,
				})
			}
		}

		// Save proxy config in service registration
		if proxyCfg != nil {
			csvc.Connect.SidecarService.Proxy = proxyCfg
		}
	}
	}

	// Register service
	err := o.cclient.Agent().ServiceRegister(csvc)
	if err != nil {
		return "", nil, err
	}

	// Get service metadata
	sidecarID := fmt.Sprintf("%s-sidecar-proxy", csvc.ID)
	service, _, err := o.cclient.Agent().Service(sidecarID, &api.QueryOptions{})
	if err != nil {
		return "", nil, err
	}

	// Check if service sidecar container needs to be created
	if service != nil {
		// Render envoy config for sidecar proxy
		ecfg, err := envoy.TemplateConfig(&envoy.TemplateParams{
			AdminAddress: "0.0.0.0",
			AdminPort:    9100,
			ServiceName:  svc.Name,
			ServiceID:    service.ID,
			AgentAddress: o.grpcAddr,
			AgentPort:    o.grpcPort,
			AgentTLS:     o.grpcTLS,
		})
		if err != nil {
			return "", nil, err
		}

		// Create port mappings for sidecar container
		ports := []entities.ContainerPortMapping{
			{
				HostPort:      uint16(service.Port),
				ContainerPort: uint16(service.Port),
				Protocol:      "tcp",
			},
		}

		// Add any expose ports
		if service.Proxy != nil && len(service.Proxy.Expose.Paths) > 0 {
			for _, expose := range service.Proxy.Expose.Paths {
				ports = append(ports, entities.ContainerPortMapping{
					HostPort:      uint16(expose.ListenerPort),
					ContainerPort: uint16(expose.ListenerPort),
					// TODO: handle this more gracefully
					Protocol: "tcp",
				})
			}
		}

		// Return a container that should be added to pod
		return csvc.ID, &entities.Container{
			Name:            fmt.Sprintf("%s-sidecar-proxy", svc.Name),
			Image:           o.envoyImage,
			ImagePullPolicy: images.PullPolicyMissing,
			RestartPolicy:   containers.RestartPolicyAlways,
			Args:            []string{"-c", "/etc/envoy/envoy.yml"},
			Ports:           ports,
			// Write the envoy bootstrap config file in the container
			Files: []entities.ContainerFile{
				{
					Destination: "/etc/envoy/envoy.yml",
					Content:     string(ecfg),
					Mode:        0644,
				},
			},
		}, nil
	}

	return csvc.ID, nil, nil
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

func findGRPCAddrPort(info *consulInfo) (string, uint16, bool, error) {
	// First look in TLS addrs
	for _, addr := range info.DebugConfig.GRPCTLSAddrs {
		if strings.HasPrefix(addr, "tcp://") {
			noPrefix := strings.TrimPrefix(addr, "tcp://")
			split := strings.Split(noPrefix, ":")

			if len(split) < 2 {
				continue
			}

			address := split[0]
			portStr := split[1]

			if a := net.ParseIP(address); a != nil && !a.IsLoopback() {
				port64, err := strconv.ParseUint(portStr, 10, 16)
				if err != nil {
					return "", 0, false, err
				}

				return address, uint16(port64), true, nil
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
			portStr := split[1]

			if a := net.ParseIP(address); a != nil && !a.IsLoopback() {
				port64, err := strconv.ParseUint(portStr, 10, 16)
				if err != nil {
					return "", 0, false, err
				}

				return address, uint16(port64), false, nil
			}
		}
	}

	return "", 0, false, fmt.Errorf("no valid grpc address found")
}
