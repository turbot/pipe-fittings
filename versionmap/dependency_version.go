package versionmap

import (
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5/plumbing"
)

type DependencyVersion struct {
	*semver.Version
	GitRef *plumbing.Reference
}

// make a collection type for this
type DependencyVersionList []*DependencyVersion

// Len returns the length of a collection. The number of Version instances
// on the slice.
func (c DependencyVersionList) Len() int {
	return len(c)
}

// Less is needed for the sort interface to compare two Version objects on the
// slice. If checks if one is less than the other.
func (c DependencyVersionList) Less(i, j int) bool {
	return c[i].Version.LessThan(c[j].Version)
}

// Swap is needed for the sort interface to replace the Version objects
// at two different positions in the slice.
func (c DependencyVersionList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
