package entities

import (
	"encoding/base64"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

// //////
// Pod //
// //////
type Pod struct {
	Name       string      `yaml:"name"`
	Containers []Container `yaml:"containers"`
	Services   []Service   `yaml:"services"`
}

func (p *Pod) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(p)

	type plain Pod
	if err := unmarshal((*plain)(p)); err != nil {
		return err
	}

	return nil
}

func (p *Pod) Hash() (string, error) {
	buf, err := yaml.Marshal(p)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// ////////////
// Container //
// ////////////
type Container struct {
	Name            string                 `yaml:"name"`
	Image           string                 `yaml:"image"`
	ImagePullPolicy string                 `default:"always" yaml:"imagePullPolicy"`
	RestartPolicy   string                 `default:"always" yaml:"restartPolicy"`
	Args            []string               `yaml:"args"`
	Ports           []ContainerPortMapping `yaml:"ports"`
	Files           []ContainerFile        `yaml:"files"`
	Mounts          []ContainerMount       `yaml:"mounts"`
}

type ContainerPortMapping struct {
	HostIP        string `yaml:"hostIP"`
	HostPort      uint16 `yaml:"hostPort"`
	ContainerPort uint16 `yaml:"containerPort"`
	Protocol      string `yaml:"protocol"`
}

type ContainerFile struct {
	Destination string `yaml:"destination"`
	Content     string `yaml:"content"`
	Mode        int64  `default:"0644" yaml:"mode"`
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
	Type        string   `default:"bind" yaml:"type"`
	Source      string   `yaml:"source"`
	Destination string   `yaml:"destination"`
	Options     []string `yaml:"options"`
}

func (m *ContainerMount) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(m)

	type plain ContainerMount
	if err := unmarshal((*plain)(m)); err != nil {
		return err
	}

	return nil
}
