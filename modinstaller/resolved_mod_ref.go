package modinstaller

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionhelpers"
	"github.com/turbot/pipe-fittings/versionmap"
)

// ResolvedModRef is a struct to represent a resolved mod git reference
type ResolvedModRef struct {
	// the FQN of the mod - also the Git URL of the mod repo
	Name string
	// the mod version
	Version *versionmap.DependencyVersion
	// the version constraint
	Constraint *versionhelpers.Constraints
	// the file path for local mods
	FilePath     string
	GitReference *plumbing.Reference
}

func NewResolvedModRef(requiredModVersion *modconfig.ModVersionConstraint, version *versionmap.DependencyVersion) (*ResolvedModRef, error) {
	res := &ResolvedModRef{
		Name:       requiredModVersion.Name,
		Version:    version,
		Constraint: requiredModVersion.Constraint,
		// this may be empty strings
		FilePath: requiredModVersion.FilePath,
	}
	if res.FilePath == "" {
		res.GitReference = version.GitRef
	}

	return res, nil
}

// DependencyPath returns name in the format <dependency name>@v<dependencyVersion>
func (r *ResolvedModRef) DependencyPath() string {
	return modconfig.BuildModDependencyPath(r.Name, r.Version.Version)
}
