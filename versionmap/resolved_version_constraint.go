package versionmap

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

// ResolvedVersionConstraint is a struct to represent a version constraint which has been resolved to specific version
// (either a git tag, git commit (for a branch constraint) or a file location)
type ResolvedVersionConstraint struct {
	*DependencyVersion
	Name          string `json:"name,omitempty"`
	Alias         string `json:"alias,omitempty"`
	Constraint    string `json:"constraint,omitempty"`
	Commit        string `json:"commit,omitempty"`
	GitRef        string `json:"git_ref,omitempty"`
	StructVersion int    `json:"struct_version,omitempty"`
}

func NewResolvedVersionConstraint(version *DependencyVersion, name, alias string, constraintString string, gitRef, commit string) *ResolvedVersionConstraint {
	return &ResolvedVersionConstraint{
		DependencyVersion: version,
		Name:              name,
		Alias:             alias,
		Constraint:        constraintString,
		Commit:            commit,
		GitRef:            gitRef,
		StructVersion:     WorkspaceLockStructVersion,
	}
}

func (c ResolvedVersionConstraint) Equals(other *ResolvedVersionConstraint) bool {
	return c.Name == other.Name &&
		c.Version.Equal(other.Version) &&
		c.Constraint == other.Constraint &&
		c.Branch == other.Branch &&
		c.Commit == other.Commit &&
		c.GitRef == other.GitRef &&
		c.FilePath == other.FilePath
}

func (c ResolvedVersionConstraint) IsPrerelease() bool {
	return c.Version != nil && c.Version.Prerelease() != "" || c.Version.Metadata() != ""
}

func (c ResolvedVersionConstraint) DependencyPath() string {
	return modconfig.BuildModDependencyPath(c.Name, c.Version)
}
