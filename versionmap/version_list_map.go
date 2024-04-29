package versionmap

import (
	"github.com/Masterminds/semver/v3"
	"sort"

	"github.com/turbot/pipe-fittings/modconfig"
)

// VersionListMap is a map keyed by dependency name storing a list of versions for each dependency
type VersionListMap map[string]semver.Collection

func (m VersionListMap) Add(name string, version *semver.Version) {
	versions := append(m[name], version) //nolint:gocritic // TODO: potential bug here?
	// reverse sort the versions
	sort.Sort(sort.Reverse(versions))
	m[name] = versions

}

// FlatMap converts the DepdencyVersionListMap map into a bool map keyed by qualified dependency name
func (m VersionListMap) FlatMap() map[string]bool {
	var res = make(map[string]bool)
	for name, versions := range m {
		for _, version := range versions {
			key := modconfig.BuildModDependencyPath(name, version)
			res[key] = true
		}
	}
	return res
}

// DepdencyVersionListMap is a map keyed by dependency name storing a list of versions for each dependency
type DepdencyVersionListMap map[string]DependencyVersionList

func (m DepdencyVersionListMap) Add(name string, version *DependencyVersion) {
	versions := append(m[name], version) //nolint:gocritic // TODO: potential bug here?
	// reverse sort the versions
	sort.Sort(sort.Reverse(versions))
	m[name] = versions

}

// FlatMap converts the DepdencyVersionListMap map into a bool map keyed by qualified dependency name
func (m DepdencyVersionListMap) FlatMap() map[string]bool {
	var res = make(map[string]bool)
	for name, versions := range m {
		for _, version := range versions {
			key := modconfig.BuildModDependencyPath(name, version.Version)
			res[key] = true
		}
	}
	return res
}
