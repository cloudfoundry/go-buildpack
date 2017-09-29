package godep

type Godep struct {
	ImportPath      string   `json:"ImportPath"`
	GoVersion       string   `json:"GoVersion"`
	Packages        []string `json:"Packages"`
	WorkspaceExists bool     `json:"WorkspaceExists"`
}
