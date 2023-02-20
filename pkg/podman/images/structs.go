package images

const (
	PullPolicyAlways  = "always"
	PullPolicyMissing = "missing"
	PullPolicyNewer   = "newer"
	PullPolicyNever   = "never"
)

type ImageInfo struct {
	Id           string
	Annotations  map[string]string
	Architecture string
	Author       string
	Comment      string
	Config       ImageConfig
	Created      string
	Digest       string
	Labels       map[string]string
	ManifestType string
	NamesHistory []string
	Os           string
	Parent       string
	RepoDigests  []string
	RepoTags     []string
	Size         uint64
	User         string
	Version      string
	VirtualSize  uint64
}

type ImageLoadResponse struct {
	Names []string
}

type PullOptions struct {
	Reference string `json:"reference"`
	Policy    string `json:"policy"`
}

type ImagePullResponse struct {
	Id     string   `json:"id"`
	Images []string `json:"images"`
	Stream string   `json:"stream"`
	Error  string   `json:"error"`
}

type ImageConfig struct {
	Cmd        []string
	Entrypoint []string
	Env        []string
	Labels     map[string]string
	StopSignal string
	User       string
	WorkingDir string
}

type ImageRootFS struct {
	Type   string
	Layers []string
}
