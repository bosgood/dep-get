package nodejs

import (
	"testing"
)

func TestCollectDependenciesEmpty(t *testing.T) {
	emptyDepsFile := NPMShrinkwrap{
		Name: "my-app",
		Version: "0.0.1",
	}

	deps := CollectDependencies(emptyDepsFile)

	if len(deps) > 0 {
		t.Errorf("Expected empty Dependencies")
	}
}

func TestCollectDependenciesOneLevel(t *testing.T) {
	oneLevelDepsFile := NPMShrinkwrap{
		Name: "my-app",
		Version: "0.0.2",
		Dependencies: map[string]*NPMShrinkwrapDependency{
			"bluebird": &NPMShrinkwrapDependency{
				Version: "3.3.4",
				From: "https://registry.npmjs.org/bluebird/-/bluebird-3.3.4.tgz",
				Resolved: "https://registry.npmjs.org/bluebird/-/bluebird-3.3.4.tgz",
			},
		},
	}

	deps := CollectDependencies(oneLevelDepsFile)
	if len(deps) != 1 {
		t.Errorf("Expected to find one dependency")
	}

	val := deps[0]
	if val.Name != "bluebird" || val.Version != "3.3.4" {
		t.Errorf("Expected bluebird dep to exist")
	}
}

func TestCollectDependenciesTwoLevels(t *testing.T) {
	twoLevelDepsFile := NPMShrinkwrap{
		Name: "my-app",
		Version: "0.0.3",
		Dependencies: map[string]*NPMShrinkwrapDependency{
			"on-finished": &NPMShrinkwrapDependency{
				Version: "2.3.0",
				From: "https://registry.npmjs.org/on-finished/-/on-finished-2.3.0.tgz",
				Resolved: "https://registry.npmjs.org/on-finished/-/on-finished-2.3.0.tgz",
				Dependencies: map[string]*NPMShrinkwrapDependency{
					"ee-first": &NPMShrinkwrapDependency{
						Version: "1.1.1",
						From: "https://registry.npmjs.org/ee-first/-/ee-first-1.1.1.tgz",
						Resolved: "https://registry.npmjs.org/ee-first/-/ee-first-1.1.1.tgz",
					},
				},
			},
		},
	}

	deps := CollectDependencies(twoLevelDepsFile)
	if len(deps) != 2 {
		t.Errorf("Expected to find two dependencies")
	}

	val := deps[0]
	if val.Name != "on-finished" || val.Version != "2.3.0" {
		t.Errorf("Expected on-finished dep to exist")
	}

	val = deps[1]
	if val.Name != "ee-first" || val.Version != "1.1.1" {
		t.Errorf("Expected ee-first dep to exist")
	}
}
