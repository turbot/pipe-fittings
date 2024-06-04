package modinstaller

import (
	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionmap"
)

// interface for easy unit testing
type updateChecker interface {
	newerVersionAvailable(requiredVersion *modconfig.ModVersionConstraint, currentVersion *semver.Version) (bool, error)
	newCommitAvailable(version *versionmap.InstalledModVersion) (bool, error)
	getUpdateStrategy() string
}

func shouldUpdateMod(installedVersion *versionmap.InstalledModVersion,
	requiredModVersion *modconfig.ModVersionConstraint,
	commandTargettingParent bool,
	updateChecker updateChecker) (bool, error) {

	// so should we update?
	// if commandTargettingParent is set or if the required version constraint is different to the locked version constraint, update

	// if this command is not targetting this mod or it's parent, do not update under any circumstances
	if !commandTargettingParent {
		return false, nil
	}

	// does the current version satisfy the required version constraint - if not we always update
	if !installedVersion.SatisfiesConstraint(requiredModVersion) {
		return true, nil
	}

	// if there is a file path - always update as child dependency requirements may have changed
	if requiredModVersion.FilePath != "" {
		return true, nil
	}

	// if we are here, we need to determine which checks to perform, based on the constraint type and pull mode
	commitCheck, versionCheck := getUpdateOperations(requiredModVersion, updateChecker.getUpdateStrategy())
	// if no checked are needed, return false
	if !commitCheck && !versionCheck {
		return false, nil
	}

	// if the constraint is a version, check for available versions
	if versionCheck && requiredModVersion.VersionConstraint() != nil {
		updateAvailable, err := updateChecker.newerVersionAvailable(requiredModVersion, installedVersion.Version)
		if err != nil {
			return false, err
		}
		if updateAvailable {
			return true, nil
		}
	}

	// do we need to perform a a commit check?
	if commitCheck {
		return updateChecker.newCommitAvailable(installedVersion)
	}

	// do not update!
	return false, nil
}
