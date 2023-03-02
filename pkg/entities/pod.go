package entities

import (
	"encoding/base64"
	"encoding/json"

	"github.com/creasty/defaults"
)

type Pod struct {
	Name       string            `yaml:"name" json:"name"`
	Hosts      map[string]string `yaml:"hosts"`
	Labels     map[string]string `yaml:"labels" json:"labels,omitempty"`
	Containers []Container       `yaml:"containers" json:"containers"`
	Services   []Service         `yaml:"services" json:"services,omitempty"`
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
	buf, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(buf), nil
}
