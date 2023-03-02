package entities

import "github.com/creasty/defaults"

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
	LocalBindAddress string `yaml:"localBindAddress"`
	LocalBindPort    uint16 `yaml:"localBindPort"`
	DestinationName  string `yaml:"destinationName"`
}

type ServiceConnectSidecarProxyExpose struct {
	Paths []ServiceConnectSidecarProxyExposePath `yaml:"paths"`
}

type ServiceConnectSidecarProxyExposePath struct {
	Path          string `yaml:"path"`
	LocalPathPort uint16 `yaml:"localPathPort"`
	ListenerPort  uint16 `yaml:"listenerPort"`
	Protocol      string `default:"http" yaml:"protocol"`
}

func (p *ServiceConnectSidecarProxyExposePath) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(p)

	type plain ServiceConnectSidecarProxyExposePath
	if err := unmarshal((*plain)(p)); err != nil {
		return err
	}

	return nil
}
