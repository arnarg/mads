package resource

type Container struct {
	Name        string           `hcl:"name,label"`
	Image       string           `hcl:"image,attr"`
	Entrypoint  []string         `hcl:"entrypoint,optional"`
	Args        []string         `hcl:"args,optional"`
	Metadata    *Metadata        `hcl:"metadata,block"`
	Networking  *NetworkConfig   `hcl:"networking,block"`
	Environment *HCLMap          `hcl:"env,block"`
	Files       []ContainerFile  `hcl:"file,block"`
	Mounts      []ContainerMount `hcl:"mount,block"`

	Pod string
}

type ContainerPortMapping struct {
	ContainerPort uint16 `hcl:"container,attr"`
	HostPort      uint16 `hcl:"host,attr"`
	HostIP        string `hcl:"host_ip,optional"`
	Protocol      string `hcl:"protocol,optional"`
}

type ContainerFile struct {
	Destination string `hcl:"destination,attr"`
	Content     string `hcl:"content,attr"`
	Mode        int64  `hcl:"mode,optional"`
}

type ContainerMount struct {
	Type        string `hcl:"type,attr"`
	Destination string `hcl:"destination,attr"`
	Source      string `hcl:"source,optional"`
}
