package modinstaller

import (
	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/versionhelpers"
)

func getVersionSatisfyingConstraint(constraint *versionhelpers.Constraints, availableVersions []*semver.Version) *semver.Version {
	// search the reverse sorted versions, finding the highest version which satisfies ALL constraints
	for _, version := range availableVersions {
		if constraint.Check(version) {
			return version
		}
	}
	return nil
}
