package versionmap

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/xlab/treeprint"
	"golang.org/x/exp/maps"
	"sort"
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
	// TACTICAL: update the struct version to the current - to enable migration
	dependency.StructVersion = WorkspaceLockStructVersion

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

// GetDependency returns the InstalledModVersion for the given path (with no constraints), and the full path (i.e. with constraints) to that dependency
func (m InstalledDependencyVersionsMap) GetDependency(path []string) (*InstalledModVersion, []string) {
	// build fully qualified path
	var fullPath []string
	if len(path) == 0 {
		return nil, nil
	}
	depName := path[0]
	key := depName
	fullPath = append(fullPath, key)
	depsForParent := m[key]
	var depVersion *InstalledModVersion
	var ok bool
	for i := 1; i < len(path); i++ {
		depName := path[i]
		depVersion, ok = depsForParent[depName]
		if !ok {
			return nil, nil
		}
		key := depVersion.DependencyPath()
		fullPath = append(fullPath, key)
		depsForParent = m[key]
	}
	return depVersion, fullPath
}

func (m InstalledDependencyVersionsMap) clone() InstalledDependencyVersionsMap {
	res := make(InstalledDependencyVersionsMap)
	for parent, deps := range m {
		res[parent] = make(map[string]*InstalledModVersion)
		for dep, v := range deps {
			res[parent][dep] = v
		}
	}
	return res
}
