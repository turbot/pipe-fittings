package versionmap

import (
	"sort"
	"strings"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/xlab/treeprint"
	"golang.org/x/exp/maps"
)

// DependencyVersionMap is a map of parent names to a map of dependencies for that parent
type DependencyVersionMap map[string]InstalledVersionMap

// AddDependency adds a dependency to the list of items installed for the given parent
func (m DependencyVersionMap) AddDependency(parentName string, dependency *InstalledModVersion) {
	// get the map for this parent
	parentItems := m[parentName]
	// create if needed
	if parentItems == nil {
		parentItems = make(InstalledVersionMap)
	}
	// add the dependency
	parentItems.AddResolvedVersion(dependency)
	// save
	m[parentName] = parentItems
}

// FlatMap converts the DependencyVersionMap into a InstalledVersionMap, keyed by mod dependency path
func (m DependencyVersionMap) FlatMap() InstalledVersionMap {
	res := make(InstalledVersionMap)
	for _, deps := range m {
		for _, dep := range deps {
			res[modconfig.BuildModDependencyPath(dep.Name, dep.Version)] = dep
		}
	}
	return res
}

func (m DependencyVersionMap) GetDependencyTree(rootName string, lock *WorkspaceLock) treeprint.Tree {
	tree := treeprint.NewWithRoot(rootName)
	// TACTICAL: make sure there is a path from the root to the keys in the map
	// (this only happens 1 level deep transitive dependencies)
	if _, containsRoot := m[rootName]; !containsRoot {
		rootMap := make(InstalledVersionMap)
		rootDeps := lock.InstallCache[rootName]

		for dep := range m {
			depName := strings.Split(dep, "@")[0]
			if rootDep, ok := rootDeps[depName]; ok {
				rootMap[depName] = rootDep
			}
		}
		m[rootName] = rootMap
	}

	m.buildTree(rootName, tree)
	return tree
}

func (m DependencyVersionMap) buildTree(name string, tree treeprint.Tree) {
	deps := m[name]
	depNames := maps.Keys(deps)
	sort.Strings(depNames)
	for _, name := range depNames {
		version := deps[name]
		fullName := modconfig.BuildModDependencyPath(name, version.Version)
		child := tree.AddBranch(fullName)
		// if there are children add them
		m.buildTree(fullName, child)
	}
}

// GetMissingFromOther returns a map of dependencies which exit in this map but not 'other'
func (m DependencyVersionMap) GetMissingFromOther(other DependencyVersionMap) DependencyVersionMap {
	res := make(DependencyVersionMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(InstalledVersionMap)
		}
		for name, dep := range deps {
			if _, ok := otherDeps[name]; !ok {
				// TODO CHECK THIS STILL WORKS
				res.AddDependency(parent, dep)
			}
		}
	}
	return res
}

func (m DependencyVersionMap) GetUpgradedInOther(other DependencyVersionMap) DependencyVersionMap {
	res := make(DependencyVersionMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(InstalledVersionMap)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				if otherDep.Version.GreaterThan(dep.Version) {
					// TODO CHECK THIS STILL WORKS
					res.AddDependency(parent, otherDep)
				}
			}
		}
	}
	return res
}

func (m DependencyVersionMap) GetDowngradedInOther(other DependencyVersionMap) DependencyVersionMap {
	res := make(DependencyVersionMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(InstalledVersionMap)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				if otherDep.Version.LessThan(dep.Version) {
					res.AddDependency(parent, otherDep)
				}
			}
		}
	}
	return res
}
