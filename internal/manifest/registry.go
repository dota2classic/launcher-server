package manifest

const RegistryFile = "registry.json"

type Package struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Folder   string `json:"folder"`
	Optional bool   `json:"optional"`
}

type Registry struct {
	Packages []Package `json:"packages"`
}
