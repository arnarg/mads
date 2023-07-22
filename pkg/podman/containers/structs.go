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

type ContainerInfo struct {
	Id        string
	Name      string
	Namespace string
	Image     string
	Pod       string
	State     ContainerState
}

type ContainerState struct {
	Dead       bool
	Error      string
	ExitCode   int
	OOMKilled  bool
	Paused     bool
	Pid        int
	Restarting bool
	Running    bool
	Status     string
}

type ContainerCreateRequest struct {
	Name          string                 `json:"name"`
	Image         string                 `json:"image"`
	Namespace     string                 `json:"namespace,omitempty"`
	Hostname      string                 `json:"hostname,omitempty"`
	Pod           string                 `json:"pod,omitempty"`
	RestartPolicy string                 `json:"restart_policy,omitempty"`
	Command       []string               `json:"command,omitempty"`
	Annotations   map[string]string      `json:"annotations,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	PortMappings  []ContainerPortMapping `json:"portmappings,omitempty"`
	Mounts        []ContainerMount       `json:"mounts,omitempty"`
	HostAdd       []string               `json:"hostadd,omitempty"`
	Env           map[string]string      `json:"env,omitempty"`
}

type ContainerMount struct {
	Destination string   `json:"destination"`
	Source      string   `json:"source,omitempty"`
	Type        string   `json:"type,omitempty"`
	Options     []string `json:"options,omitempty"`
}

type ContainerPortMapping struct {
	HostIP        string `json:"host_ip"`
	HostPort      uint16 `json:"host_port"`
	ContainerPort uint16 `json:"container_port"`
	Protocol      string `json:"protocol"`
	Range         uint16 `json:"range"`
}
