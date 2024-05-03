package modinstaller

import (
	"github.com/turbot/pipe-fittings/v2/versionhelpers"
	"github.com/turbot/pipe-fittings/v2/versionmap"
)

func getVersionSatisfyingConstraint(constraint *versionhelpers.Constraints, availableVersions versionmap.ResolvedVersionConstraintList) *versionmap.ResolvedVersionConstraint {
	// search the reverse sorted versions, finding the highest version which satisfies ALL constraints
	for _, version := range availableVersions {
		if constraint.Check(version.Version) {
			return version
		}
	}
	return nil
}
