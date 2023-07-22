package mads

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/arnarg/mads/pkg/mads/resource"
	"github.com/arnarg/mads/pkg/podman"
	"github.com/arnarg/mads/pkg/podman/containers"
	"github.com/arnarg/mads/pkg/podman/images"
	"github.com/arnarg/mads/pkg/podman/pods"
)

type Config struct {
	SocketPath string
}

type Client struct {
	pm *podman.Client
}

func NewClient(cfg *Config) *Client {
	// Create a podman client
	pm := podman.NewClient(&podman.Config{SocketPath: cfg.SocketPath})

	return &Client{
		pm: pm,
	}
}

func (c *Client) ApplyModule(ctx context.Context, module *Module) error {
	// Apply all containers
	for _, container := range module.Containers {
		// Apply container
		if err := c.applyContainer(ctx, container); err != nil {
			return err
		}

		// Start container
		if err := c.pm.Containers().Start(ctx, container.Name); err != nil {
			return fmt.Errorf("could not start container '%s': %s", container.Name, err)
		}
	}

	// Apply all pods
	for _, pod := range module.Pods {
		// Apply pod
		if err := c.applyPod(ctx, pod); err != nil {
			return err
		}

		// Start pod
		if err := c.pm.Pods().Start(ctx, pod.Name); err != nil {
			return fmt.Errorf("could not start pod '%s': %s", pod.Name, err)
		}
	}

	return nil
}

func (c *Client) applyPod(ctx context.Context, pod *resource.Pod) error {
	netConf := pod.Networking

	// Some config from the containers need to be actually set on the pod
	for _, ctr := range pod.Containers {
		// Move exposed ports to pod
		if ctr.Networking != nil {
			for _, mapping := range ctr.Networking.PortMappings {
				// If netConf is nil we need to initialize it
				if netConf == nil {
					netConf = &resource.NetworkConfig{}
				}

				netConf.PortMappings = append(netConf.PortMappings, mapping)
			}
		}
	}

	// Create pod creation request
	req := &pods.PodCreateRequest{
		Name: pod.Name,
	}

	// If there's any network config we add it to the request
	if netConf != nil {
		// If hostname is set we set that in the request
		if netConf.Hostname != nil {
			req.Hostname = *netConf.Hostname
		}

		// Add all hosts to the pod request
		req.HostAdd = netConf.Hosts.ToHostAdd()

		// Add all port mappings
		for _, mapping := range netConf.PortMappings {
			req.PortMappings = append(req.PortMappings, pods.PodPortMapping{
				HostIP:        mapping.HostIP,
				HostPort:      mapping.HostPort,
				ContainerPort: mapping.ContainerPort,
				Protocol:      mapping.Protocol,
			})
		}
	}

	// Create the pod
	id, err := c.pm.Pods().Create(ctx, req)
	if err != nil {
		return err
	}

	// Add all containers to the pod
	for _, ctr := range pod.Containers {
		// Create a container
		ctrName := fmt.Sprintf("%s-%s", pod.Name, ctr.Name)

		// Create a new container definition so we can
		// change some settings
		cont := &resource.Container{
			Name:        ctrName,
			Pod:         id,
			Image:       ctr.Image,
			Entrypoint:  ctr.Entrypoint,
			Args:        ctr.Args,
			Metadata:    ctr.Metadata,
			Environment: ctr.Environment,
			Files:       ctr.Files,
			Mounts:      ctr.Mounts,
		}

		// Copy some networking config over too
		if ctr.Networking != nil {
			cont.Networking = &resource.NetworkConfig{
				Hostname: ctr.Networking.Hostname,
				Hosts:    ctr.Networking.Hosts,
			}
		}

		// Apply container
		if err := c.applyContainer(ctx, cont); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) applyContainer(ctx context.Context, container *resource.Container) error {
	// Get image
	imageID, err := realizeImage(ctx, c.pm, container.Image, "missing")
	if err != nil {
		return err
	}

	// Create container creation request
	req := &containers.ContainerCreateRequest{
		Name:    container.Name,
		Image:   imageID,
		Command: container.Args,
	}

	// If container should be part of a pod we set it here
	if container.Pod != "" {
		req.Pod = container.Pod
	}

	// If metadata is provided we want to add it the request
	if container.Metadata != nil {
		req.Annotations = container.Metadata.Annotations.ToMap()
		req.Labels = container.Metadata.Labels.ToMap()
	}

	// Add env to create request
	if container.Environment != nil {
		req.Env = container.Environment.ToMap()
	}

	// If network config is provided we want to add it to the request
	if container.Networking != nil {
		// If hostname is provided we set that in the request
		if container.Networking.Hostname != nil {
			req.Hostname = *container.Networking.Hostname
		}

		// Add all hosts to container
		req.HostAdd = container.Networking.Hosts.ToHostAdd()

		// Add all port mappings
		for _, mapping := range container.Networking.PortMappings {
			req.PortMappings = append(req.PortMappings, containers.ContainerPortMapping{
				HostIP:        mapping.HostIP,
				HostPort:      mapping.HostPort,
				ContainerPort: mapping.ContainerPort,
				Protocol:      mapping.Protocol,
			})
		}
	}

	// Add all mounts to request
	if len(container.Mounts) > 0 {
		for _, mount := range container.Mounts {
			req.Mounts = append(req.Mounts, containers.ContainerMount{
				Type:        mount.Type,
				Source:      mount.Source,
				Destination: mount.Destination,
			})
		}
	}

	// Create container
	err = c.pm.Containers().Create(ctx, req)
	if err != nil {
		return err
	}

	// Write files to tar archive
	if len(container.Files) > 0 {
		// Write tar archive into buffer
		buf, err := createTarArchive(ctx, container.Files)
		if err != nil {
			return err
		}

		// Copy tar archive buffer into container
		err = c.pm.Containers().Copy(ctx, container.Name, buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTarArchive(ctx context.Context, files []resource.ContainerFile) (io.Reader, error) {
	buf := &bytes.Buffer{}

	// Create a tar writer
	tw := tar.NewWriter(buf)
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
			return nil, err
		}

		// Write content to tar
		_, err = tw.Write(buf)
		if err != nil {
			return nil, err
		}
	}

	return buf, nil
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
