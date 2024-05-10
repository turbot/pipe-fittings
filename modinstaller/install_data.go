package modinstaller

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/versionmap"
	"github.com/xlab/treeprint"
	"log/slog"
)

type InstallData struct {
	// record of the full dependency tree
	Lock    *versionmap.WorkspaceLock
	NewLock *versionmap.WorkspaceLock

	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	allAvailable versionmap.ResolvedVersionConstraintListMap

	// list of dependencies installed by recent install operation
	Installed versionmap.InstalledDependencyVersionsMap
	// list of dependencies which have been upgraded
	Upgraded versionmap.InstalledDependencyVersionsMap
	// list of dependencies which have been downgraded
	Downgraded versionmap.InstalledDependencyVersionsMap
	// list of dependencies which have been uninstalled
	Uninstalled  versionmap.InstalledDependencyVersionsMap
	WorkspaceMod *modconfig.Mod
}

func NewInstallData(workspaceLock *versionmap.WorkspaceLock, workspaceMod *modconfig.Mod) *InstallData {
	return &InstallData{
		Lock:         workspaceLock,
		WorkspaceMod: workspaceMod,
		NewLock:      versionmap.EmptyWorkspaceLock(workspaceLock),
		allAvailable: make(versionmap.ResolvedVersionConstraintListMap),
		Installed:    make(versionmap.InstalledDependencyVersionsMap),
		Upgraded:     make(versionmap.InstalledDependencyVersionsMap),
		Downgraded:   make(versionmap.InstalledDependencyVersionsMap),
		Uninstalled:  make(versionmap.InstalledDependencyVersionsMap),
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

// update the lock with the NewLock and dtermine if any mods have been uninstalled
func (d *InstallData) onInstallComplete() {
	installed := d.NewLock.InstallCache.GetMissingFromOther(d.Lock.InstallCache)
	uninstalled := d.Lock.InstallCache.GetMissingFromOther(d.NewLock.InstallCache)
	upgraded := d.Lock.InstallCache.GetUpgradedInOther(d.NewLock.InstallCache)
	downgraded := d.Lock.InstallCache.GetDowngradedInOther(d.NewLock.InstallCache)

	// for each installed parent in the installed map, check whethe rthe parent was actually upgraded
	// - if so the children will wrongly have been identified as installed,
	// (as their parent has a new version so GetMissingFromOther will not find the parent in the prev)
	// when actually they may have been upgraded or unchanged
	for installedParent, deps := range installed {
		for _, updgradedForParent := range upgraded {
			for _, u := range updgradedForParent {
				if u.DependencyPath() == installedParent {
					slog.Info("Checking for parent upgrade", "installedParent", installedParent, "u", u)
				}
				slog.Info("Checking for parent upgrade", "installedParent", installedParent, "u", u)
				slog.Info("Checking for parent upgrade", "deps", deps)

			}
			//if updgraded.Name != installedParent.Name {
			//	for name, dep := range deps {
			//		if _, ok := updgraded[name]; ok {
			//			delete(installed[parent], name)
			//		}
			//	}
			//}
		}
	}
	d.Installed = installed
	d.Uninstalled = uninstalled
	d.Upgraded = upgraded
	d.Downgraded = downgraded

	d.Lock = d.NewLock
}

func (d *InstallData) GetUpdatedTree() treeprint.Tree {
	return d.Upgraded.GetDependencyTree(d.WorkspaceMod.GetInstallCacheKey(), d.Lock)
}

func (d *InstallData) GetInstalledTree() treeprint.Tree {
	return d.Installed.GetDependencyTree(d.WorkspaceMod.GetInstallCacheKey(), d.Lock)
}

func (d *InstallData) GetUninstalledTree() treeprint.Tree {
	return d.Uninstalled.GetDependencyTree(d.WorkspaceMod.GetInstallCacheKey(), d.Lock)
}
