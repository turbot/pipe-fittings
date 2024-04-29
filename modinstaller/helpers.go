package modinstaller

import (
	"github.com/turbot/pipe-fittings/versionhelpers"
	"github.com/turbot/pipe-fittings/versionmap"
)

func getVersionSatisfyingConstraint(constraint *versionhelpers.Constraints, availableVersions versionmap.DependencyVersionList) *versionmap.DependencyVersion {
	// search the reverse sorted versions, finding the highest version which satisfies ALL constraints
	for _, version := range availableVersions {
		if constraint.Check(version.Version) {
			return version
		}
	}
	return nil
}
