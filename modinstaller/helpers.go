package modinstaller

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionhelpers"
)

func getVersionSatisfyingConstraint(constraint *versionhelpers.Constraints, availableVersions modconfig.DependencyVersionList) *modconfig.DependencyVersion {
	// search the reverse sorted versions, finding the highest version which satisfies ALL constraints
	for _, version := range availableVersions {
		if constraint.Check(version.Version) {
			return version
		}
	}
	return nil
}
