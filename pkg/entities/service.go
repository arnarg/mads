package entities

// This is mostly a re-creation of a subset of a consul agent service structs

type Service struct {
	Name    string         `yaml:"name"`
	Tags    []string       `yaml:"tags"`
	Port    int            `yaml:"port"`
	Connect ServiceConnect `yaml:"connect"`
}

type ServiceConnect struct {
	Native         bool                   `yaml:"native"`
	SidecarService *ServiceConnectSidecar `yaml:"sidecarService"`
}

type ServiceConnectSidecar struct {
	Proxy *ServiceConnectSidecarProxy `yaml:"proxy"`
}

type ServiceConnectSidecarProxy struct {
	Upstreams []ServiceConnectSidecarProxyUpstream `yaml:"upstreams"`
	Expose    ServiceConnectSidecarProxyExpose     `yaml:"expose"`
}

type ServiceConnectSidecarProxyUpstream struct {
}

type ServiceConnectSidecarProxyExpose struct {
}
