package modinstaller

import (
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/versionmap"
)

type InstallData struct {
	// record of the full dependency tree
	Lock    *versionmap.WorkspaceLock
	NewLock *versionmap.WorkspaceLock

	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	allAvailable versionmap.ResolvedVersionConstraintListMap

	Installed   [][]string
	Uninstalled [][]string
	Upgraded    [][]string
	Downgraded  [][]string

	WorkspaceMod *modconfig.Mod
}

func NewInstallData(workspaceLock *versionmap.WorkspaceLock, workspaceMod *modconfig.Mod) *InstallData {
	return &InstallData{
		Lock:         workspaceLock,
		WorkspaceMod: workspaceMod,
		NewLock:      versionmap.EmptyWorkspaceLock(workspaceLock),
		allAvailable: make(versionmap.ResolvedVersionConstraintListMap),
	}
}

// onModInstalled is called when a dependency is satisfied by installing a mod version
func (d *InstallData) onModInstalled(installedMod *DependencyMod, parent *modconfig.Mod) {
	parentPath := parent.GetInstallCacheKey()
	// update lock
	d.NewLock.InstallCache.AddDependency(parentPath, installedMod.InstalledVersion)
}

// addExisting is called when a dependency is satisfied by a mod which is already installed
// (perhaps as a depdency of another mod)
func (d *InstallData) addExisting(existingDep *DependencyMod, parent *modconfig.Mod) {
	// update lock
	parentPath := parent.GetInstallCacheKey()
	d.NewLock.InstallCache.AddDependency(parentPath, existingDep.InstalledVersion)
}

// retrieve all available mod versions from our cache, or from Git if not yet cached
func (d *InstallData) getAvailableModVersions(modName string, includePrerelease bool) (versionmap.ResolvedVersionConstraintList, error) {
	// have we already loaded the versions for this mod
	availableVersions, ok := d.allAvailable[modName]
	if ok {
		return availableVersions, nil
	}

	// so we have not cached this yet - retrieve from Git
	var err error
	availableVersions, err = getTagVersionsFromGit(modName, includePrerelease)
	if err != nil {
		return nil, perr.BadRequestWithMessage("could not retrieve version data from Git URL " + modName + " - " + err.Error())
	}
	// update our cache
	d.allAvailable[modName] = availableVersions

	return availableVersions, nil
}

// update the lock with the NewLock and determine if mod installs, upgrades, downgrades or uninstalls have occurred
func (d *InstallData) onInstallComplete() error {
	root := d.WorkspaceMod.GetInstallCacheKey()

	getUpgradedDowngradedUninstalled := func(depPath []string, oldDep *versionmap.InstalledModVersion) error {
		// find this path in new lock
		newDep, fullPath := d.NewLock.InstallCache.GetDependency(depPath)

		if newDep == nil {
			// get the full path from the old lock
			_, oldFullPath := d.Lock.InstallCache.GetDependency(depPath)
			d.Uninstalled = append(d.Uninstalled, oldFullPath)
			return nil
		}

		switch {

		// if they are both version constraints, compare the versions
		case oldDep.DependencyVersion.Version != nil && newDep.DependencyVersion.Version != nil:
			switch {
			case oldDep.DependencyVersion.Version.GreaterThan(newDep.DependencyVersion.Version):
				d.Upgraded = append(d.Upgraded, fullPath)
			case newDep.DependencyVersion.Version.LessThan(oldDep.DependencyVersion.Version):
				d.Downgraded = append(d.Downgraded, fullPath)
			// otherwise check the commit hash
			case oldDep.Commit != newDep.Commit:
				d.Upgraded = append(d.Upgraded, fullPath)
			}

		// if they are both the same branch, compare the commit hash
		case oldDep.Branch != "" && oldDep.Branch == newDep.Branch:
			if oldDep.Commit != newDep.Commit {
				d.Upgraded = append(d.Upgraded, fullPath)
			}
			// if they are bothe the same tag, compare the commit hash
		case oldDep.Tag != "" && oldDep.Tag == newDep.Tag:
			if oldDep.Commit != newDep.Commit {
				d.Upgraded = append(d.Upgraded, fullPath)
			}
			// both the same filepath - nothing to do
		case oldDep.FilePath != "" && newDep.FilePath == oldDep.FilePath:
			// nothing to do here
		default:
			// to get here, then they have must have different constraint types or different filepaths
			//- mark as uninstalled and installed
			_, oldFullPath := d.Lock.InstallCache.GetDependency(depPath)
			d.Installed = append(d.Installed, fullPath)
			d.Uninstalled = append(d.Uninstalled, oldFullPath)
		}
		return nil
	}

	if err := d.Lock.WalkCache(root, getUpgradedDowngradedUninstalled); err != nil {
		return err
	}

	// we also need to check for any new dependencies which have been installed
	// do this by walking the new lock and checking if the path exists in the old lock
	getInstalled := func(depPath []string, newDep *versionmap.InstalledModVersion) error {
		// find this path in new lock
		oldDep, _ := d.Lock.InstallCache.GetDependency(depPath)

		if oldDep == nil {
			// get the full path from the new lock
			_, fullPath := d.NewLock.InstallCache.GetDependency(depPath)
			if fullPath == nil {
				return fmt.Errorf("could not find path %v in new lock", depPath)
			}
			d.Installed = append(d.Installed, fullPath)
			return nil
		}
		return nil
	}
	if err := d.NewLock.WalkCache(root, getInstalled); err != nil {
		return err
	}

	d.Lock = d.NewLock
	return nil
}
