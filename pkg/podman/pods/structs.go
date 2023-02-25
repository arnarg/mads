package pods

const (
	PodStateCreated = "Created"
	PodStateRunning = "Running"
)

type PodInfo struct {
	Id               string
	Name             string
	Namespace        string
	Hostname         string
	Labels           map[string]string
	Containers       []PodContainer
	State            string
	ExitPolicy       string
	Created          string
	CreateCommand    []string
	CreateCgroup     bool
	CgroupParent     string
	CgroupPath       string
	CreateInfra      bool
	InfraContainerID string
	InfraConfig      PodInfraConfig
	SharedNamespaces []string
	NumContainers    int
}

type PodInfraConfig struct {
	PortBindings        map[string][]PortBinding
	HostNetwork         bool
	StaticIP            string
	StaticMAC           string
	NoManagedResolvConf bool
	PidNS               string `json:"pid_ns"`
	UserNS              string `json:"userns"`
	UtsNS               string `json:"uts_ns"`
}

type PortBinding struct {
	HostIp   string
	HostPort string
}

type PodContainer struct {
	Id    string
	Name  string
	State string
}

type PodCreateRequest struct {
	Name         string            `json:"name"`
	Hostname     string            `json:"hostname"`
	Labels       map[string]string `json:"labels"`
	PortMappings []PodPortMapping  `json:"portmappings"`
}

type PodPortMapping struct {
	HostIP        string `json:"host_ip"`
	HostPort      uint16 `json:"host_port"`
	ContainerPort uint16 `json:"container_port"`
	Protocol      string `json:"protocol"`
	Range         uint16 `json:"range"`
}
