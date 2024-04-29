package modinstaller

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	git "github.com/go-git/go-git/v5"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/versionmap"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

type ModInstaller struct {
	installData *InstallData

	// this will be updated as changes are made to dependencies
	workspaceMod *modconfig.Mod

	// since changes are made to workspaceMod, we need a copy of the Require as is on disk
	// to be able to calculate changes
	oldRequire *modconfig.Require

	mods versionmap.VersionConstraintMap

	// the final resting place of all dependency mods
	modsPath string
	// a shadow directory for installing mods
	// this is necessary to make mod installation transactional
	shadowDirPath string

	workspacePath string

	// what command is being run
	command string
	// are dependencies being added to the workspace
	dryRun bool
	// do we force install even if there are require errors
	force bool
	// optional map of installed plugin versions
	pluginVersions *modconfig.PluginVersionMap
}

func NewModInstaller(opts *InstallOpts) (*ModInstaller, error) {
	if opts.WorkspaceMod == nil {
		return nil, fmt.Errorf("no workspace mod passed to mod installer")
	}
	i := &ModInstaller{
		workspacePath:  opts.WorkspaceMod.ModPath,
		workspaceMod:   opts.WorkspaceMod,
		command:        opts.Command,
		dryRun:         opts.DryRun,
		force:          opts.Force,
		pluginVersions: opts.PluginVersions,
	}

	if opts.WorkspaceMod.Require != nil {
		i.oldRequire = opts.WorkspaceMod.Require.Clone()
	}

	if err := i.setModsPath(); err != nil {
		return nil, err
	}

	// load lock file
	workspaceLock, err := versionmap.LoadWorkspaceLock(i.workspacePath)
	if err != nil {
		return nil, err
	}

	// create install data
	i.installData = NewInstallData(workspaceLock, i.workspaceMod)

	// parse args to get the required mod versions
	requiredMods, err := i.GetRequiredModVersionsFromArgs(opts.ModArgs)
	if err != nil {
		return nil, err
	}
	i.mods = requiredMods

	return i, nil
}

func (i *ModInstaller) UninstallWorkspaceDependencies(ctx context.Context) error {
	workspaceMod := i.workspaceMod

	// remove required dependencies from the mod file
	if len(i.mods) == 0 {
		workspaceMod.RemoveAllModDependencies()

	} else {
		// verify all the mods specifed in the args exist in the modfile
		workspaceMod.RemoveModDependencies(i.mods)
	}

	// uninstall by calling Install
	if err := i.installMods(ctx, workspaceMod.Require.Mods, workspaceMod); err != nil {
		return err
	}

	if workspaceMod.Require.Empty() {
		workspaceMod.Require = nil
	}

	// if this is a dry run, return now
	if i.dryRun {
		slog.Debug("UninstallWorkspaceDependencies - dry-run=true, returning before saving mod file and cache\n")
		return nil
	}

	// write the lock file
	if err := i.installData.Lock.Save(); err != nil {
		return err
	}

	//  now safe to save the mod file
	if err := i.updateModFile(); err != nil {
		return err
	}

	// tidy unused mods
	if viper.GetBool(constants.ArgPrune) {
		if _, err := i.Prune(); err != nil {
			return err
		}
	}
	return nil
}

// InstallWorkspaceDependencies installs all dependencies of the workspace mod
func (i *ModInstaller) InstallWorkspaceDependencies(ctx context.Context) (err error) {
	workspaceMod := i.workspaceMod
	defer func() {
		if err != nil && i.force {
			// suppress the error since this is a forced install
			slog.Debug("suppressing error in InstallWorkspaceDependencies because force is enabled", err)
			err = nil
		}
		// tidy unused mods
		// (put in defer so it still gets called in case of errors)
		if viper.GetBool(constants.ArgPrune) && !i.dryRun {
			// be sure not to overwrite an existing return error
			_, pruneErr := i.Prune()
			if pruneErr != nil && err == nil {
				err = pruneErr
			}
		}
	}()

	if validationErrors := workspaceMod.ValidateRequirements(i.pluginVersions); len(validationErrors) > 0 {
		if !i.force {
			// if this is not a force install, return errors in validation
			return error_helpers.CombineErrors(validationErrors...)
		}
		// ignore if this is a force install
		error_helpers.ShowWarning(fmt.Sprintf("--force is set, ignoring %d mod validation %s:\n\t%s",
			len(validationErrors),
			utils.Pluralize("error", len(validationErrors)),
			error_helpers.CombineErrors(validationErrors...).Error()))
	}

	// if mod args have been provided, add them to the workspace mod requires
	// (this will replace any existing dependencies of same name)
	if len(i.mods) > 0 {
		workspaceMod.AddModDependencies(i.mods)
	}

	if err := i.installMods(ctx, workspaceMod.Require.Mods, workspaceMod); err != nil {
		return err
	}

	// if this is a dry run, return now
	if i.dryRun {
		slog.Debug("InstallWorkspaceDependencies - dry-run=true, returning before saving mod file and cache\n")
		return nil
	}

	// write the lock file
	if err := i.installData.Lock.Save(); err != nil {
		return err
	}

	//  now safe to save the mod file
	if err := i.updateModFile(); err != nil {
		return err
	}

	if !workspaceMod.HasDependentMods() {
		// there are no dependencies - delete the cache
		err = i.installData.Lock.Delete()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *ModInstaller) GetModList() string {
	return i.installData.Lock.GetModList(i.workspaceMod.GetInstallCacheKey())
}

func (i *ModInstaller) removeOldShadowDirectories() error {
	removeErrors := []error{}
	// get the parent of the 'mods' directory - all shadow directories are siblings of this
	parent := filepath.Base(i.modsPath)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return err
	}
	for _, dir := range entries {
		if dir.IsDir() && filepaths.IsModInstallShadowPath(dir.Name()) {
			err := os.RemoveAll(filepath.Join(parent, dir.Name()))
			if err != nil {
				removeErrors = append(removeErrors, err)
			}
		}
	}
	return error_helpers.CombineErrors(removeErrors...)
}

func (i *ModInstaller) setModsPath() error {
	i.modsPath = filepaths.WorkspaceModPath(i.workspacePath)
	_ = i.removeOldShadowDirectories()
	i.shadowDirPath = filepaths.WorkspaceModShadowPath(i.workspacePath)
	return nil
}

// commitShadow recursively copies over the contents of the shadow directory
// to the mods directory, replacing conflicts as it goes
// (uses `os.Create(dest)` under the hood - which truncates the target)
func (i *ModInstaller) commitShadow(ctx context.Context) error {
	if error_helpers.IsContextCanceled(ctx) {
		return ctx.Err()
	}
	if _, err := os.Stat(i.shadowDirPath); os.IsNotExist(err) {
		// nothing to do here
		// there's no shadow directory to commit
		// this is not an error and may happen when install does not make any changes
		return nil
	}
	entries, err := os.ReadDir(i.shadowDirPath)
	if err != nil {
		return fmt.Errorf("could not read shadow directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		source := filepath.Join(i.shadowDirPath, entry.Name())
		destination := filepath.Join(i.modsPath, entry.Name())
		slog.Debug("copying", source, destination)
		if err := copy.Copy(source, destination); err != nil {
			// return sperr.WrapWithRootMessage(err, "could not commit shadow directory '%s'", entry.Name())
			return fmt.Errorf("could not commit shadow directory '%s': %w", entry.Name(), err)
		}
	}
	return nil
}

func (i *ModInstaller) shouldCommitShadow(ctx context.Context, installError error) bool {
	// no commit if this is a dry run
	if i.dryRun {
		return false
	}
	// commit if this is forced - even if there's errors
	return installError == nil || i.force
}

func (i *ModInstaller) installMods(ctx context.Context, mods []*modconfig.ModVersionConstraint, parent *modconfig.Mod) (err error) {
	defer func() {
		var commitErr error
		if i.shouldCommitShadow(ctx, err) {
			commitErr = i.commitShadow(ctx)
		}

		// if this was forced, we need to suppress the install error
		// otherwise the calling code will fail
		if i.force {
			err = nil
		}

		// ensure we return any commit error
		if commitErr != nil {
			err = commitErr
		}

		// force remove the shadow directory - we can ignore any error here, since
		// these directories get cleaned up before any install session
		_ = os.RemoveAll(i.shadowDirPath)
	}()

	var errors []error
	// forceupdate is used for child depdency mods - if the parent is being updated
	const forceUpdate = false
	for _, requiredModVersion := range mods {
		// do we have this mod installed for another parent?
		currentMod, err := i.getModForRequirement(ctx, requiredModVersion, forceUpdate)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		if err := i.installModDependenciesRecursively(ctx, requiredModVersion, currentMod, parent, forceUpdate); err != nil {
			errors = append(errors, err)
		}
	}

	// update the lock to be the new lock, and record any uninstalled mods
	i.installData.onInstallComplete()

	return i.buildInstallError(errors)
}

func (i *ModInstaller) buildInstallError(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	verb := "install"
	if i.updating() {
		verb = "update"
	}
	prefix := fmt.Sprintf("%d %s failed to %s", len(errors), utils.Pluralize("dependency", len(errors)), verb)
	err := error_helpers.CombineErrorsWithPrefix(prefix, errors...)
	return err
}

func (i *ModInstaller) installModDependenciesRecursively(ctx context.Context, requiredModVersion *modconfig.ModVersionConstraint, dependencyMod *DependencyMod, parent *modconfig.Mod, forceUpdate bool) error {
	if error_helpers.IsContextCanceled(ctx) {
		// short circuit if the execution context has been cancelled
		return ctx.Err()
	}

	var errors []error
	// if dependencyMod we must install it
	if dependencyMod == nil {
		// get available versions for this mod
		includePrerelease := requiredModVersion.IsPrerelease()
		availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name, includePrerelease)
		if err != nil {
			return err
		}

		// get a resolved mod ref that satisfies the version constraints
		resolvedRef, err := i.getModRefSatisfyingConstraints(requiredModVersion, availableVersions)
		if err != nil {
			return err
		}

		// install the mod
		dependencyMod, err = i.install(ctx, resolvedRef, parent)
		if err != nil {
			return err
		}

	} else {
		// this mod is already installed - just update the install data
		i.installData.addExisting(dependencyMod, parent)
		slog.Debug(fmt.Sprintf("not installing %s with version constraint %s as version %s is already installed", requiredModVersion.Name, requiredModVersion.VersionString, dependencyMod.Mod.Version))
	}

	// if we are updating dependencyMod, we should update all its children
	if !forceUpdate {
		forceUpdate = i.isUpdateCommandTargettingMod(dependencyMod.Mod.DependencyName)
	}

	// to get here we have the dependency mod - either we installed it or it was already installed
	// recursively install its dependencies
	for _, childDependency := range dependencyMod.Mod.Require.Mods {
		childDependencyMod, err := i.getModForRequirement(ctx, childDependency, forceUpdate)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if err := i.installModDependenciesRecursively(ctx, childDependency, childDependencyMod, dependencyMod.Mod, forceUpdate); err != nil {
			errors = append(errors, err)
			continue
		}
	}

	return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("%d child %s failed to install", len(errors), utils.Pluralize("dependency", len(errors))), errors...)
}

// is this an update command and this mod was included in the args - or there were no args
func (i *ModInstaller) isUpdateCommandTargettingMod(modName string) bool {
	if !i.updating() {
		return false
	}
	// this is an update command - if there are no args, sdo we are updating all mods
	if len(i.mods) == 0 {
		return true
	}
	// if this mod in the list of updates?
	_, updateMod := i.mods[modName]
	return updateMod
}

func (i *ModInstaller) getModForRequirement(ctx context.Context, requiredModVersion *modconfig.ModVersionConstraint, forceUpdate bool) (*DependencyMod, error) {
	// do we have an installed version of this mod matching the required mod constraint
	installedVersion, err := i.installData.Lock.FindLockedModVersion(requiredModVersion)
	if err != nil {
		return nil, err
	}
	if installedVersion == nil {
		// try new lock - we may have just installed it
		installedVersion, err = i.installData.NewLock.FindLockedModVersion(requiredModVersion)
		if err != nil {
			return nil, err
		}
		if installedVersion == nil {
			return nil, nil
		}
	}

	// can we update this
	shouldUpdate, err := i.shouldUpdateMod(installedVersion.ResolvedVersionConstraint, requiredModVersion, forceUpdate)
	if err != nil {
		return nil, err
	}

	if shouldUpdate {
		// return nil mod to indicate we should update
		return nil, nil
	}
	// load the existing mod and return
	mod, err := i.loadDependencyMod(ctx, installedVersion.ResolvedVersionConstraint)
	if err != nil {
		return nil, err
	}
	return &DependencyMod{
		Mod:              mod,
		InstalledVersion: installedVersion,
	}, nil
}

// loadDependencyMod tries to load the mod definition from the shadow directory
// and falls back to the 'mods' directory of the root mod
func (i *ModInstaller) loadDependencyMod(ctx context.Context, modVersion *versionmap.ResolvedVersionConstraint) (*modconfig.Mod, error) {
	// construct the dependency path - this is the relative path of the dependency we are installing
	dependencyPath := modVersion.DependencyPath()

	// first try loading from the shadow dir
	modDefinition, err := i.loadDependencyModFromRoot(ctx, i.shadowDirPath, dependencyPath)
	if err != nil {
		return nil, err
	}

	// failed to load from shadow dir, try mods dir
	if modDefinition == nil {
		modDefinition, err = i.loadDependencyModFromRoot(ctx, i.modsPath, dependencyPath)
		if err != nil {
			return nil, err
		}
	}

	// if we still failed, give up
	if modDefinition == nil {
		return nil, fmt.Errorf("could not find dependency mod '%s'", dependencyPath)
	}

	// set the DependencyName, DependencyPath and AppVersion properties on the mod
	if err := i.setModDependencyConfig(modDefinition, dependencyPath); err != nil {
		return nil, err
	}

	return modDefinition, nil
}

func (i *ModInstaller) loadDependencyModFromRoot(ctx context.Context, modInstallRoot string, dependencyPath string) (*modconfig.Mod, error) {
	slog.Debug("loadDependencyModFromRoot", "dependencyPath", dependencyPath, "modInstallRoot", modInstallRoot)

	modPath := path.Join(modInstallRoot, dependencyPath)
	modDefinition, err := parse.LoadModfile(modPath)
	if err != nil {
		// return nil, sperr.WrapWithMessage(err, "failed to load mod definition for %s from %s", dependencyPath, modInstallRoot)
		return nil, fmt.Errorf("failed to load mod definition for %s from %s: %w", dependencyPath, modInstallRoot, err)
	}
	return modDefinition, nil
}

// determine if we should update this mod, and if so whether there is an update available
func (i *ModInstaller) shouldUpdateMod(installedVersion *versionmap.ResolvedVersionConstraint, requiredModVersion *modconfig.ModVersionConstraint, forceUpdate bool) (bool, error) {
	// so should we update?
	// if forceUpdate is set or if the required version constraint is different to the locked version constraint, update
	// TODO update to not assume a constraint
	isSatisfied, errs := requiredModVersion.Constraint().Validate(installedVersion.Version)
	if len(errs) > 0 {
		return false, error_helpers.CombineErrors(errs...)
	}

	// is this mod being updated (i.e. is this an update command and this mod was included in the args - or there were no args))
	updating := forceUpdate || i.isUpdateCommandTargettingMod(installedVersion.Name)

	// if this is an update command, or if the current version does not satisfy the required version constraint,
	// check for update
	if updating || !isSatisfied {
		// get available versions for this mod
		includePrerelease := requiredModVersion.IsPrerelease()
		availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name, includePrerelease)
		if err != nil {
			return false, err
		}

		return i.updateAvailable(requiredModVersion, installedVersion.Version, availableVersions)
	}
	return false, nil

}

// determine whether there is a newer mod version avoilable which satisfies the dependency version constraint
func (i *ModInstaller) updateAvailable(requiredVersion *modconfig.ModVersionConstraint, currentVersion *semver.Version, availableVersions versionmap.DependencyVersionList) (bool, error) {
	latestVersion, err := i.getModRefSatisfyingConstraints(requiredVersion, availableVersions)
	if err != nil {
		return false, err
	}
	if latestVersion.Version.GreaterThan(currentVersion) {
		return true, nil
	}
	return false, nil
}

// get the most recent available mod version which satisfies the version constraint
func (i *ModInstaller) getModRefSatisfyingConstraints(modVersion *modconfig.ModVersionConstraint, availableVersions versionmap.DependencyVersionList) (*versionmap.ResolvedVersionConstraint, error) {
	// find a version which satisfies the version constraint
	var dependencyVersion = getVersionSatisfyingConstraint(modVersion.Constraint(), availableVersions)
	if dependencyVersion == nil {
		return nil, fmt.Errorf("no version of %s found satisfying version constraint: %s", modVersion.Name, modVersion.VersionString)
	}

	//name, alias string, version *semver.Version, constraintString string, gitRef, commit string

	return versionmap.NewResolvedVersionConstraint(dependencyVersion, modVersion.Name, modVersion.VersionString, dependencyVersion.GitRef), nil
}

// install a mod
func (i *ModInstaller) install(ctx context.Context, dependency *versionmap.ResolvedVersionConstraint, parent *modconfig.Mod) (_ *DependencyMod, err error) {
	var modDef *modconfig.Mod
	// get the temp location to install the mod to
	dependencyPath := dependency.DependencyPath()
	destPath := i.getDependencyShadowPath(dependencyPath)

	defer func() {
		if err == nil {
			i.installData.onModInstalled(dependency, modDef, parent)
		}
	}()

	// does target dir exist>
	if _, err := os.Stat(destPath); err == nil {
		// this is unexpected - if the mod was successfully installed it would have been found in the lock file so we
		// would not be installing it
		// so just delete this folder and reinstall
		slog.Info("dependency mod folder already exists but not found in lock file - deleting and reinstalling", "mod", dependencyPath, "path", destPath)
		// delete the det path
		if err := os.RemoveAll(destPath); err != nil {
			return nil, fmt.Errorf("could not remove existing mod path '%s': %w", destPath, err)
		}
	}

	slog.Debug("installing", "dependency", dependencyPath, "in", destPath)
	if err = i.installFromGit(dependency, destPath); err != nil {
		return nil, err
	}
	// now load the installed mod and return it
	modDef, err = parse.LoadModfile(destPath)
	if err != nil {
		return nil, err
	}
	if modDef == nil {
		return nil, fmt.Errorf("'%s' has no mod definition file", dependencyPath)
	}

	if !i.dryRun {
		// now the mod is installed in its final location, set mod dependency path
		if err := i.setModDependencyConfig(modDef, dependencyPath); err != nil {
			return nil, err
		}
	}

	// create installed version
	installedModVersion := &versionmap.InstalledModVersion{
		ResolvedVersionConstraint: dependency,
		Alias:                     modDef.ShortName,
	}

	return &DependencyMod{
		Mod:              modDef,
		InstalledVersion: installedModVersion,
	}, nil
}

func (i *ModInstaller) installFromGit(dependency *versionmap.ResolvedVersionConstraint, installPath string) error {
	// get the mod from git = first try https
	gitUrl := getGitUrl(dependency.Name, GitUrlModeHTTPS)
	slog.Debug("installFromGit cloning the repo", gitUrl, dependency.GitRefStr)

	gitHubToken := getGitToken()

	// otherwise use go-got to clone
	cloneOptions := git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: dependency.GitRef.Name(),
		Depth:         1,
		SingleBranch:  true,
	}

	if gitHubToken != "" {
		// if we have a token, use it
		cloneOptions.Auth = getGitAuthForToken(gitHubToken)
	}

	_, err := git.PlainClone(installPath,
		false, &cloneOptions)

	if err != nil {
		// if that failed, try ssh
		gitUrl := getGitUrl(dependency.Name, GitUrlModeSSH)
		slog.Debug(">>> cloning", gitUrl, dependency.GitRefStr)
		_, err = git.PlainClone(installPath,
			false,
			&git.CloneOptions{
				URL:           gitUrl,
				ReferenceName: dependency.GitRef.Name(),
				Depth:         1,
				SingleBranch:  true,
			})
	}
	if err != nil {
		return err
	}

	// If your function returns an error, make sure to handle it
	return nil

	// verify the cloned repo contains a valid modfile
	if err := i.verifyModFile(dependency, installPath); err != nil {
		return err

	}
	return nil
}

// build the path of the temp location to copy this depednency to
func (i *ModInstaller) getDependencyDestPath(dependencyFullName string) string {
	return filepath.Join(i.modsPath, dependencyFullName)
}

// build the path of the temp location to copy this dependency to
func (i *ModInstaller) getDependencyShadowPath(dependencyFullName string) string {
	return filepath.Join(i.shadowDirPath, dependencyFullName)
}

// set the mod dependency path
func (i *ModInstaller) setModDependencyConfig(mod *modconfig.Mod, dependencyPath string) error {
	return mod.SetDependencyConfig(dependencyPath)
}

func (i *ModInstaller) updating() bool {
	return i.command == "update"
}

func (i *ModInstaller) uninstalling() bool {
	return i.command == "uninstall"
}

func (i *ModInstaller) verifyModFile(dependency *versionmap.ResolvedVersionConstraint, installPath string) error {
	for _, modFilePath := range app_specific.ModFilePaths(installPath) {
		_, err := os.Stat(modFilePath)
		if err == nil {
			// found the modfile
			return nil
		}
	}
	return sperr.New("mod '%s' does not contain a valid mod file", dependency.Name)
}
