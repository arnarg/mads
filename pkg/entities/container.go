package entities

import "github.com/creasty/defaults"

type Container struct {
	Name            string                 `yaml:"name" json:"name"`
	Image           string                 `yaml:"image" json:"image"`
	ImagePullPolicy string                 `default:"always" yaml:"imagePullPolicy" json:"imagePullPolicy,omitempty"`
	RestartPolicy   string                 `default:"always" yaml:"restartPolicy" json:"restartPolicy,omitempty"`
	Args            []string               `yaml:"args" json:"args,omitempty"`
	Ports           []ContainerPortMapping `yaml:"ports" json:"ports,omitempty"`
	Files           []ContainerFile        `yaml:"files" json:"files,omitempty"`
	Mounts          []ContainerMount       `yaml:"mounts" mounts:"mounts,omitempty"`
}

func (c *Container) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(c)

	type plain Container
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return nil
}

type ContainerPortMapping struct {
	HostIP        string `yaml:"hostIP" json:"hostIP,omitempty"`
	HostPort      uint16 `yaml:"hostPort" json:"hostPort,omitempty"`
	ContainerPort uint16 `yaml:"containerPort" json:"containerPort,omitempty"`
	Protocol      string `yaml:"protocol" json:"protocol,omitempty"`
}

type ContainerFile struct {
	Destination string `yaml:"destination" json:"destination"`
	Content     string `yaml:"content" json:"content"`
	Mode        int64  `default:"0644" yaml:"mode" json:"mode,omitempty"`
}

func (f *ContainerFile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(f)

	type plain ContainerFile
	if err := unmarshal((*plain)(f)); err != nil {
		return err
	}

	return nil
}

type ContainerMount struct {
	Type        string   `default:"bind" yaml:"type" json:"type"`
	Source      string   `yaml:"source" json:"source"`
	Destination string   `yaml:"destination" json:"destination"`
	Options     []string `yaml:"options" json:"options,omitempty"`
}

func (m *ContainerMount) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(m)

	type plain ContainerMount
	if err := unmarshal((*plain)(m)); err != nil {
		return err
	}

	return nil
}
