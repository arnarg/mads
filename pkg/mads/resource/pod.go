package resource

type Pod struct {
	Name       string         `hcl:"name,label"`
	Metadata   *Metadata      `hcl:"metadata,block"`
	Networking *NetworkConfig `hcl:"networking,block"`
	Containers []*Container   `hcl:"container,block"`
}
