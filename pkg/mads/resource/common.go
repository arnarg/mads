package resource

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Annotation struct {
	Name  string    `hcl:"name,label"`
	Value cty.Value `hcl:"value,attr"`
}

type Annotations []Annotation

func (a Annotations) ToMap() map[string]string {
	annotations := map[string]string{}

	for _, annotation := range a {
		if annotation.Value.IsNull() {
			continue
		}

		switch annotation.Value.Type() {
		case cty.String:
			annotations[annotation.Name] = annotation.Value.AsString()
		case cty.Bool:
			b := annotation.Value.True()
			annotations[annotation.Name] = strconv.FormatBool(b)
		case cty.Number:
			num := annotation.Value.AsBigFloat()

			if num.IsInt() {
				i, _ := num.Int64()
				annotations[annotation.Name] = strconv.FormatInt(i, 10)
			} else {
				annotations[annotation.Name] = fmt.Sprintf("%f", num)
			}
		}
	}

	return annotations
}

type Label struct {
	Name  string    `hcl:"name,label"`
	Value cty.Value `hcl:"value,attr"`
}

type Labels []Label

func (a Labels) ToMap() map[string]string {
	labels := map[string]string{}

	for _, label := range a {
		if label.Value.IsNull() {
			continue
		}

		switch label.Value.Type() {
		case cty.String:
			labels[label.Name] = label.Value.AsString()
		case cty.Bool:
			b := label.Value.True()
			labels[label.Name] = strconv.FormatBool(b)
		case cty.Number:
			num := label.Value.AsBigFloat()

			if num.IsInt() {
				i, _ := num.Int64()
				labels[label.Name] = strconv.FormatInt(i, 10)
			} else {
				labels[label.Name] = fmt.Sprintf("%f", num)
			}
		}
	}

	return labels
}

type Metadata struct {
	Annotations Annotations `hcl:"annotation,block"`
	Labels      Labels      `hcl:"label,block"`
}

type Host struct {
	Host    string `hcl:"host,label"`
	Address string `hcl:"address,attr"`
}

type Hosts []Host

func (h Hosts) ToHostAdd() []string {
	hosts := []string{}

	for _, host := range h {
		hosts = append(hosts, fmt.Sprintf("%s:%s", host.Host, host.Address))
	}

	return hosts
}

type PortMapping struct {
	ContainerPort uint16 `hcl:"container,attr"`
	HostPort      uint16 `hcl:"host,attr"`
	HostIP        string `hcl:"host_ip,optional"`
	Protocol      string `hcl:"protocol,optional"`
}

type NetworkConfig struct {
	Hostname     *string       `hcl:"hostname,optional"`
	Hosts        Hosts         `hcl:"host,block"`
	PortMappings []PortMapping `hcl:"expose,block"`
}

type HCLMap struct {
	Body hcl.Body `hcl:",remain"`
}

func (m *HCLMap) ToMap() map[string]string {
	vals := map[string]string{}

	attrs, diag := m.Body.JustAttributes()
	if diag.HasErrors() {
		return vals
	}

	for _, attr := range attrs {
		value, diag := attr.Expr.Value(nil)
		if diag.HasErrors() {
			continue
		}

		if value.IsNull() {
			continue
		}

		switch value.Type() {
		case cty.String:
			vals[attr.Name] = value.AsString()
		case cty.Bool:
			b := value.True()
			vals[attr.Name] = strconv.FormatBool(b)
		case cty.Number:
			num := value.AsBigFloat()

			if num.IsInt() {
				i, _ := num.Int64()
				vals[attr.Name] = strconv.FormatInt(i, 10)
			} else {
				vals[attr.Name] = fmt.Sprintf("%f", num)
			}
		}
	}

	return vals
}
