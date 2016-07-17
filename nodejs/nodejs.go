package nodejs

// DependenciesFileName is the name of the nodejs dependencies lock file
const DependenciesFileName = "npm-shrinkwrap.json"

// PackageJSON represents a nodejs package.json file
type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// NPMShrinkwrap represents a npm-shrinkwrap.json file
type NPMShrinkwrap struct {
	Name         string                              `json:"name"`
	Version      string                              `json:"version"`
	Dependencies map[string]*NPMShrinkwrapDependency `json:"dependencies"`
}

// NPMShrinkwrapDependency represents a dependencies
// block from an npm-shrinkwrap.json file
type NPMShrinkwrapDependency struct {
	Version      string                              `json:"version"`
	From         string                              `json:"from"`
	Resolved     string                              `json:"resolved"`
	Dependencies map[string]*NPMShrinkwrapDependency `json:"dependencies"`
}
