package versionmap

import (
	"sort"
	"strings"

	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/xlab/treeprint"
	"golang.org/x/exp/maps"
)

// InstalledDependencyVersionsMap is a map of parent names to a map of dependencies for that parent, keyed by dependency name
type InstalledDependencyVersionsMap map[string]map[string]*InstalledModVersion

// AddDependency adds a dependency to the list of items installed for the given parent
func (m InstalledDependencyVersionsMap) AddDependency(parentName string, dependency *InstalledModVersion) {
	// get the map for this parent
	parentItems := m[parentName]
	// create if needed
	if parentItems == nil {
		parentItems = make(map[string]*InstalledModVersion)
	}
	// add the dependency
	parentItems[dependency.Name] = dependency
	// save
	m[parentName] = parentItems
}

// FlatMap converts the InstalledDependencyVersionsMap into a map[string]*InstalledModVersion, keyed by mod dependency path
func (m InstalledDependencyVersionsMap) FlatMap() map[string]*InstalledModVersion {
	res := make(map[string]*InstalledModVersion)
	for _, deps := range m {
		for _, dep := range deps {
			res[modconfig.BuildModDependencyPath(dep.Name, &dep.DependencyVersion)] = dep
		}
	}
	return res
}

func (m InstalledDependencyVersionsMap) GetDependencyTree(rootName string, lock *WorkspaceLock) treeprint.Tree {
	tree := treeprint.NewWithRoot(rootName)
	// TACTICAL: make sure there is a path from the root to the keys in the map
	// (this only happens 1 level deep transitive dependencies)
	if _, containsRoot := m[rootName]; !containsRoot {
		rootMap := make(map[string]*InstalledModVersion)
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

func (m InstalledDependencyVersionsMap) buildTree(name string, tree treeprint.Tree) {
	deps := m[name]
	depNames := maps.Keys(deps)
	sort.Strings(depNames)
	for _, name := range depNames {
		installedVersion := deps[name]
		fullName := modconfig.BuildModDependencyPath(name, &installedVersion.DependencyVersion)
		child := tree.AddBranch(fullName)
		// if there are children add them
		m.buildTree(fullName, child)
	}
}

// GetMissingFromOther returns a map of dependencies which exit in this map but not 'other'
func (m InstalledDependencyVersionsMap) GetMissingFromOther(other InstalledDependencyVersionsMap) InstalledDependencyVersionsMap {
	res := make(InstalledDependencyVersionsMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(map[string]*InstalledModVersion)
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

func (m InstalledDependencyVersionsMap) GetUpgradedInOther(other InstalledDependencyVersionsMap) InstalledDependencyVersionsMap {
	res := make(InstalledDependencyVersionsMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(map[string]*InstalledModVersion)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				switch {
				case dep.Version != nil && otherDep.Version != nil:
					if dep.Version.LessThan(otherDep.Version) {
						res.AddDependency(parent, otherDep)
					}
				case dep.Branch != "" && otherDep.Branch != "":
					if dep.Commit != otherDep.Commit {
						res.AddDependency(parent, otherDep)
					}
				case dep.FilePath != "" && otherDep.FilePath != "":
					// TODO - check file hash????

				}
			}
		}
	}
	return res
}

func (m InstalledDependencyVersionsMap) GetDowngradedInOther(other InstalledDependencyVersionsMap) InstalledDependencyVersionsMap {
	res := make(InstalledDependencyVersionsMap)
	for parent, deps := range m {
		otherDeps := other[parent]
		if otherDeps == nil {
			otherDeps = make(map[string]*InstalledModVersion)
		}
		for name, dep := range deps {
			if otherDep, ok := otherDeps[name]; ok {
				if otherDep.LessThan(&dep.DependencyVersion) {
					res.AddDependency(parent, otherDep)
				}
			}
		}
	}
	return res
}
