package containers

const (
	MountTypeBind   = "bind"
	MountTypeVolume = "volume"
	MountTypeTmpfs  = "tmpfs"

	RestartPolicyNo            = "no"
	RestartPolicyAlways        = "always"
	RestartPolicyOnFailure     = "on-failure"
	RestartPolicyUnlessStopped = "unless-stopped"
)

type ContainerCreateRequest struct {
	Name          string           `json:"name"`
	Image         string           `json:"image"`
	Namespace     string           `json:"namespace,omitempty"`
	Pod           string           `json:"pod,omitempty"`
	RestartPolicy string           `json:"restart_policy,omitempty"`
	Command       []string         `json:"command,omitempty"`
	Mounts        []ContainerMount `json:"mounts,omitempty"`
	HostAdd       []string         `json:"hostadd,omitempty"`
}

type ContainerMount struct {
	Destination string   `json:"destination"`
	Source      string   `json:"source,omitempty"`
	Type        string   `json:"type,omitempty"`
	Options     []string `json:"options,omitempty"`
}
