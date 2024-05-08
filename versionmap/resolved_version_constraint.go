package versionmap

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

// InstalledModVersion is a struct to represent a version of a mod which has been installed
type InstalledModVersion struct {
	*ResolvedVersionConstraint
	Alias string `json:"alias"`
}

func (v InstalledModVersion) SatisfiesConstraint(requiredVersion *modconfig.ModVersionConstraint) bool {
	if c := requiredVersion.VersionConstraint(); c != nil {
		if v.Version == nil {
			return false
		}
		return c.Check(v.Version)
	}
	if b := requiredVersion.BranchName; b != "" {
		return v.Branch == b
	}
	if f := requiredVersion.FilePath; f != "" {
		return v.FilePath == f
	}
	if t := requiredVersion.Tag; t != "" {
		return v.Tag == t
	}
	// unexpected
	return false
}

// ResolvedVersionConstraint is a struct to represent a version constraint which has been resolved to specific version
// (either a git tag, git commit (for a branch constraint) or a file location)
type ResolvedVersionConstraint struct {
	modconfig.DependencyVersion
	Name string `json:"name,omitempty"`

	Commit        string `json:"commit,omitempty"`
	GitRefStr     string `json:"git_ref,omitempty"`
	StructVersion int    `json:"struct_version,omitempty"`
}

func NewResolvedVersionConstraint(version *modconfig.DependencyVersion, name string, gitRef *plumbing.Reference) *ResolvedVersionConstraint {
	res := &ResolvedVersionConstraint{
		DependencyVersion: *version,
		Name:              name,
		StructVersion:     WorkspaceLockStructVersion,
	}
	if gitRef != nil {
		res.GitRefStr = gitRef.Name().String()
		res.Commit = gitRef.Hash().String()
	}
	return res

}

func (c ResolvedVersionConstraint) Equals(other *ResolvedVersionConstraint) bool {
	return c.Name == other.Name &&
		c.Version.Equal(other.Version) &&
		c.Branch == other.Branch &&
		c.Commit == other.Commit &&
		c.GitRefStr == other.GitRefStr &&
		c.FilePath == other.FilePath &&
		c.Tag == other.Tag
}

func (c ResolvedVersionConstraint) IsPrerelease() bool {
	return c.Version != nil && c.Version.Prerelease() != "" || c.Version.Metadata() != ""
}

func (c ResolvedVersionConstraint) DependencyPath() string {
	return modconfig.BuildModDependencyPath(c.Name, &c.DependencyVersion)
}

type ResolvedVersionConstraintList []*ResolvedVersionConstraint

// Len returns the length of a collection. The number of Version instances
// on the slice.
func (c ResolvedVersionConstraintList) Len() int {
	return len(c)
}

// Less is needed for the sort interface to compare two Version objects on the
// slice. If checks if one is less than the other.
func (c ResolvedVersionConstraintList) Less(i, j int) bool {
	// if both i and j have versions,
	return c[i].Version.LessThan(c[j].Version)
}

// Swap is needed for the sort interface to replace the Version objects
// at two different positions in the slice.
func (c ResolvedVersionConstraintList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type ResolvedVersionConstraintListMap map[string]ResolvedVersionConstraintList
