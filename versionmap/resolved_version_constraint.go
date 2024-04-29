package versionmap

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/turbot/pipe-fittings/modconfig"
)

// ResolvedVersionConstraint is a struct to represent a version constraint which has been resolved to specific version
// (either a git tag, git commit (for a branch constraint) or a file location)
type InstalledModVersion struct {
	*ResolvedVersionConstraint
	Alias string `json:"alias,omitempty"`
}

// ResolvedVersionConstraint is a struct to represent a version constraint which has been resolved to specific version
// (either a git tag, git commit (for a branch constraint) or a file location)
type ResolvedVersionConstraint struct {
	DependencyVersion
	Name          string `json:"name,omitempty"`
	Constraint    string `json:"constraint,omitempty"`
	Commit        string `json:"commit,omitempty"`
	GitRefStr     string `json:"git_ref,omitempty"`
	StructVersion int    `json:"struct_version,omitempty"`
}

func NewResolvedVersionConstraint(version *DependencyVersion, name, constraintString string, gitRef *plumbing.Reference) *ResolvedVersionConstraint {
	return &ResolvedVersionConstraint{
		DependencyVersion: *version,
		Name:              name,
		Constraint:        constraintString,
		Commit:            gitRef.Hash().String(),
		GitRefStr:         gitRef.Name().String(),
		StructVersion:     WorkspaceLockStructVersion,
	}
}

func (c ResolvedVersionConstraint) Equals(other *ResolvedVersionConstraint) bool {
	return c.Name == other.Name &&
		c.Version.Equal(other.Version) &&
		c.Constraint == other.Constraint &&
		c.Branch == other.Branch &&
		c.Commit == other.Commit &&
		c.GitRefStr == other.GitRefStr &&
		c.FilePath == other.FilePath
}

func (c ResolvedVersionConstraint) IsPrerelease() bool {
	return c.Version != nil && c.Version.Prerelease() != "" || c.Version.Metadata() != ""
}

func (c ResolvedVersionConstraint) DependencyPath() string {
	return modconfig.BuildModDependencyPath(c.Name, c.Version)
}
