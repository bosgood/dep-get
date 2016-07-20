package nodejs

import (
	"fmt"
)

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

// NodeDependency declares a node dependency and a way to download it
type NodeDependency struct {
	Name       string
	Version    string
	PackageURL string
}

func collectDependencies(
	memo []NodeDependency,
	deps map[string]*NPMShrinkwrapDependency,
) []NodeDependency {
	for k, v := range deps {
		memo = append(memo, NodeDependency{
			Name:       k,
			Version:    v.Version,
			PackageURL: v.Resolved,
		})
		memo = collectDependencies(memo, v.Dependencies)
	}
	return memo
}

// CollectDependencies flattens all given node dependencies into one list
func CollectDependencies(npmShrinkwrap NPMShrinkwrap) []NodeDependency {
	var deps []NodeDependency
	deps = collectDependencies(deps, npmShrinkwrap.Dependencies)

	var dedupedDeps []NodeDependency
	var depSet = make(map[string]bool)
	for _, dep := range deps {
		depStr := fmt.Sprintf("%s@%s", dep.Name, dep.Version)
		if dupeDep := depSet[depStr]; !dupeDep {
			dedupedDeps = append(dedupedDeps, dep)
			depSet[depStr] = true
		}
	}

	return dedupedDeps
}
