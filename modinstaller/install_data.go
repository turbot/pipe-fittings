package modinstaller

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/versionmap"
	"github.com/xlab/treeprint"
)

type InstallData struct {
	// record of the full dependency tree
	Lock    *versionmap.WorkspaceLock
	NewLock *versionmap.WorkspaceLock

	// ALL the available versions for each dependency mod(we populate this in a lazy fashion)
	allAvailable versionmap.DepdencyVersionListMap

	// list of dependencies installed by recent install operation
	Installed versionmap.DependencyVersionMap
	// list of dependencies which have been upgraded
	Upgraded versionmap.DependencyVersionMap
	// list of dependencies which have been downgraded
	Downgraded versionmap.DependencyVersionMap
	// list of dependencies which have been uninstalled
	Uninstalled  versionmap.DependencyVersionMap
	WorkspaceMod *modconfig.Mod
}

func NewInstallData(workspaceLock *versionmap.WorkspaceLock, workspaceMod *modconfig.Mod) *InstallData {
	return &InstallData{
		Lock:         workspaceLock,
		WorkspaceMod: workspaceMod,
		NewLock:      versionmap.EmptyWorkspaceLock(workspaceLock),
		allAvailable: make(versionmap.DepdencyVersionListMap),
		Installed:    make(versionmap.DependencyVersionMap),
		Upgraded:     make(versionmap.DependencyVersionMap),
		Downgraded:   make(versionmap.DependencyVersionMap),
		Uninstalled:  make(versionmap.DependencyVersionMap),
	}
}

// onModInstalled is called when a dependency is satisfied by installing a mod version
func (d *InstallData) onModInstalled(dependency *versionmap.ResolvedVersionConstraint, modDef *modconfig.Mod, parent *modconfig.Mod) {
	parentPath := parent.GetInstallCacheKey()
	// get the constraint from the parent (it must be there)
	//modVersionConstraint := parent.Require.GetModDependency(dependency.Name).ConstraintString

	// update lock
	d.NewLock.InstallCache.AddDependency(dependency, modDef.ShortName, modDef.Version, parentPath)
	//d.NewLock.InstallCache.AddDependency(dependency.Name, modDef.ShortName, modDef.Version, modVersionConstraint, parentPath, dependency.GitReference.Name().String(), dependency.GitReference.Hash().String())
}

// addExisting is called when a dependency is satisfied by a mod which is already installed
// (perhaps as a depdency of another mod)
func (d *InstallData) addExisting(dependencyName string, existingDep *DependencyMod, constraintString string, parent *modconfig.Mod) {
	// update lock
	parentPath := parent.GetInstallCacheKey()
	d.NewLock.InstallCache.AddDependency(existingDep.Constraint, existingDep.Mod.ShortName, existingDep.Mod.Version, parentPath)
}

// retrieve all available mod versions from our cache, or from Git if not yet cached
func (d *InstallData) getAvailableModVersions(modName string, includePrerelease bool) (versionmap.DependencyVersionList, error) {
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
	d.Installed = d.NewLock.InstallCache.GetMissingFromOther(d.Lock.InstallCache)
	d.Uninstalled = d.Lock.InstallCache.GetMissingFromOther(d.NewLock.InstallCache)
	d.Upgraded = d.Lock.InstallCache.GetUpgradedInOther(d.NewLock.InstallCache)
	d.Downgraded = d.Lock.InstallCache.GetDowngradedInOther(d.NewLock.InstallCache)
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
