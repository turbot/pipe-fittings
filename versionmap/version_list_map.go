package versionmap

import (
	"sort"

	"github.com/turbot/pipe-fittings/v2/modconfig"
)

// DependencyVersionListMap is a map keyed by dependency name storing a list of versions for each dependency
type DependencyVersionListMap map[string]modconfig.DependencyVersionList

func (m DependencyVersionListMap) Add(name string, version *modconfig.DependencyVersion) {
	versions := m[name]
	versions = append(versions, version)
	// reverse sort the versions
	sort.Sort(sort.Reverse(versions))
	m[name] = versions

}

// FlatMap converts the DependencyVersionListMap map into a lookup keyed by qualified dependency name
func (m DependencyVersionListMap) FlatMap() map[string]struct{} {
	var res = make(map[string]struct{})
	for name, versions := range m {
		for _, version := range versions {
			key := modconfig.BuildModDependencyPath(name, version)
			res[key] = struct{}{}
		}
	}
	return res
}
