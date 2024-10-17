package modconfig

import (
	"github.com/Masterminds/semver/v3"
)

// DependencyVersion is a struct that encapsulates the version of a mod dependency
// the version may be specified as a filepath, a branch or a semver version
type DependencyVersion struct {
	Version  *semver.Version `json:"version,omitempty"`
	Branch   string          `json:"branch,omitempty"`
	FilePath string          `json:"path,omitempty"`
	Tag      string          `json:"tag,omitempty"`
}

func (v DependencyVersion) String() string {
	stringForm := ""
	if v.Version != nil {
		stringForm = v.Version.String()
	}
	if v.Branch != "" {
		stringForm += " " + v.Branch
	}
	if v.FilePath != "" {
		stringForm += " " + v.FilePath
	}
	if v.Tag != "" {
		stringForm += " " + v.Tag
	}
	return stringForm
}

func (v DependencyVersion) Equal(other *DependencyVersion) bool {
	if other == nil {
		return false

	}
	// if both have Version, check that
	if v.Version != nil && other.Version != nil {
		return v.Version.Equal(other.Version) && v.Version.Metadata() == other.Version.Metadata()
	}
	// if both have Branch, check that
	if v.Branch != "" && other.Branch != "" {
		return v.Branch == other.Branch
	}
	// if both have FilePath, check that
	if v.FilePath != "" && other.FilePath != "" {
		return v.FilePath == other.FilePath
	}

	// if both have Tag, check that
	if v.Tag != "" && other.Tag != "" {
		return v.Tag == other.Tag
	}
	return false
}

func (v DependencyVersion) LessThan(other *DependencyVersion) bool {
	// if bother have versions, check
	if v.Version != nil && other.Version != nil {
		return v.Version.LessThan(other.Version)
	}
	return false
}

func (v DependencyVersion) GreaterThan(other *DependencyVersion) bool {
	// if bother have versions, check
	if v.Version != nil && other.Version != nil {
		return v.Version.GreaterThan(other.Version)
	}
	return false
}

type DependencyVersionList []*DependencyVersion

// Len returns the length of a collection. The number of Version instances
// on the slice.
func (c DependencyVersionList) Len() int {
	return len(c)
}

// Less is needed for the sort interface to compare two Version objects on the
// slice. If checks if one is less than the other.
func (c DependencyVersionList) Less(i, j int) bool {
	return c[i].LessThan(c[j])
}

// Swap is needed for the sort interface to replace the Version objects
// at two different positions in the slice.
func (c DependencyVersionList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
