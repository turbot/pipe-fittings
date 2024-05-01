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
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
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

	mods map[string]*modconfig.ModVersionConstraint

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

	updateStrategy string
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
		updateStrategy: opts.UpdateStrategy,
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

func (i *ModInstaller) installModDependenciesRecursively(ctx context.Context, requiredModVersion *modconfig.ModVersionConstraint, dependencyMod *DependencyMod, parent *modconfig.Mod, commandTargettingParent bool) error {
	if error_helpers.IsContextCanceled(ctx) {
		// short circuit if the execution context has been cancelled
		return ctx.Err()
	}

	var errors []error
	// if dependencyMod we must install it
	if dependencyMod == nil {
		// find a gitref to install
		var err error
		dependencyMod, err = i.install(ctx, requiredModVersion, parent)
		if err != nil {
			return err
		}

	} else {
		// this mod is already installed - just update the install data
		i.installData.addExisting(dependencyMod, parent)
		slog.Debug(fmt.Sprintf("not installing %s with version constraint %s as version %s is already installed", requiredModVersion.Name, requiredModVersion.VersionString, dependencyMod.Mod.Version))
	}

	// if we are updating dependencyMod, we should update all its children
	// set the value of commandTargettingParent to pass down to child mods
	if !commandTargettingParent {
		commandTargettingParent = i.isCommandTargettingMod(dependencyMod.Mod.DependencyName)
	}

	// to get here we have the dependency mod - either we installed it or it was already installed
	// recursively install its dependencies
	for _, childDependency := range dependencyMod.Mod.Require.Mods {
		childDependencyMod, err := i.getModForRequirement(ctx, childDependency, commandTargettingParent)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if err := i.installModDependenciesRecursively(ctx, childDependency, childDependencyMod, dependencyMod.Mod, commandTargettingParent); err != nil {
			errors = append(errors, err)
			continue
		}
	}

	return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("%d child %s failed to install", len(errors), utils.Pluralize("dependency", len(errors))), errors...)
}

func (i *ModInstaller) install(ctx context.Context, requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (_ *DependencyMod, err error) {
	var modDef *modconfig.Mod
	var destPath string
	var resolvedRef *versionmap.ResolvedVersionConstraint

	defer func() {
		if err == nil {
			i.installData.onModInstalled(resolvedRef, modDef, parent)
		}
	}()

	switch {
	case requiredModVersion.VersionConstraint() != nil:
		// get available versions for this mod
		includePrerelease := requiredModVersion.IsPrerelease()
		availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name, includePrerelease)
		if err != nil {
			return nil, err
		}
		// get a resolved mod ref that satisfies the version constraints
		resolvedRef, err = i.getModRefSatisfyingVersionConstraint(requiredModVersion, availableVersions)
		if err != nil {
			return nil, err
		}
		// install the mod
		destPath, err = i.installFromTag(ctx, resolvedRef, parent)
		if err != nil {
			return nil, err
		}

	case requiredModVersion.Branch() != "":
		resolvedRef, destPath, err = i.installFromBranch(ctx, requiredModVersion, parent)
		if err != nil {
			return nil, err
		}

	case requiredModVersion.FilePath() != "":
	}

	// now load the installed mod and return it
	modDef, err = parse.LoadModfile(destPath)
	if err != nil {
		return nil, err
	}
	if modDef == nil {
		return nil, fmt.Errorf("'%s' has no mod definition file", resolvedRef.DependencyPath())
	}

	if !i.dryRun {
		// now the mod is installed in its final location, set mod dependency path
		if err := i.setModDependencyConfig(modDef, resolvedRef.DependencyPath()); err != nil {
			return nil, err
		}
	}

	// create installed version
	installedModVersion := &versionmap.InstalledModVersion{
		ResolvedVersionConstraint: resolvedRef,
		InstallPath:               destPath,
	}

	return &DependencyMod{
		Mod:              modDef,
		InstalledVersion: installedModVersion,
	}, nil

}

// install a mod
func (i *ModInstaller) installFromTag(ctx context.Context, dependency *versionmap.ResolvedVersionConstraint, parent *modconfig.Mod) (string, error) {
	// get the temp location to install the mod to
	dependencyPath := dependency.DependencyPath()
	destPath := i.getDependencyShadowPath(dependencyPath)

	// does target dir exist>
	if _, err := os.Stat(destPath); err == nil {
		// this is unexpected - if the mod was successfully installed it would have been found in the lock file so we
		// would not be installing it
		// so just delete this folder and reinstall
		slog.Info("dependency mod folder already exists but not found in lock file - deleting and reinstalling", "mod", dependencyPath, "path", destPath)
		// delete the det path
		if err := os.RemoveAll(destPath); err != nil {
			return "", fmt.Errorf("could not remove existing mod path '%s': %w", destPath, err)
		}
	}

	slog.Debug("installing", "dependency", dependencyPath, "in", destPath)
	if _, err := i.installFromGit(dependency.Name, plumbing.ReferenceName(dependency.GitRefStr), destPath); err != nil {
		return "", err
	}
	return destPath, nil

}

func (i *ModInstaller) installFromBranch(_ context.Context, modVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (*versionmap.ResolvedVersionConstraint, string, error) {
	// build a DependencyVersion
	var dependencyVersion = &modconfig.DependencyVersion{
		Branch: modVersion.Branch(),
	}

	// get the temp location to install the mod to
	// just use the original constraint as the dependency path
	dependencyPath := modconfig.BuildModDependencyPath(modVersion.Name, dependencyVersion)
	destPath := i.getDependencyShadowPath(dependencyPath)

	// does target dir exist>
	if _, err := os.Stat(destPath); err == nil {
		// this is unexpected - if the mod was successfully installed it would have been found in the lock file so we
		// would not be installing it
		// so just delete this folder and reinstall
		slog.Info("dependency mod folder already exists but not found in lock file - deleting and reinstalling", "mod", dependencyPath, "path", destPath)
		// delete the det path
		if err := os.RemoveAll(destPath); err != nil {
			return nil, "", fmt.Errorf("could not remove existing mod path '%s': %w", destPath, err)
		}
	}

	slog.Debug("installing", "dependency", dependencyPath, "in", destPath)
	// build a git ref for the branch
	gitRef := plumbing.NewBranchReferenceName(dependencyVersion.Branch)
	repo, err := i.installFromGit(modVersion.Name, gitRef, destPath)
	if err != nil {
		return nil, "", err
	}
	// get the commit hash
	ref, err := repo.Reference(gitRef, true)

	// build a ResolvedVersionConstraint
	resolvedRef := versionmap.NewResolvedVersionConstraint(dependencyVersion, modVersion.Name, modVersion.Branch(), ref)
	return resolvedRef, destPath, nil
}

func (i *ModInstaller) installFromGit(repoName string, gitRefName plumbing.ReferenceName, installPath string) (*git.Repository, error) {
	// get the mod from git = first try https
	gitUrl := getGitUrl(repoName, GitUrlModeHTTPS)
	slog.Debug("installFromGit cloning the repo", gitUrl, gitRefName.String())

	gitHubToken := getGitToken()
	cloneOptions := git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: gitRefName,
		Depth:         1,
		SingleBranch:  true,
	}
	if gitHubToken != "" {
		// if we have a token, use it
		cloneOptions.Auth = getGitAuthForToken(gitHubToken)
	}

	repo, err := git.PlainClone(installPath,
		false, &cloneOptions)

	if err != nil {
		// if that failed, try ssh
		gitUrl := getGitUrl(repoName, GitUrlModeSSH)
		slog.Debug(">>> cloning", gitUrl, gitRefName.String())
		_, err = git.PlainClone(installPath,
			false,
			&git.CloneOptions{
				URL:           gitUrl,
				ReferenceName: gitRefName,
				Depth:         1,
				SingleBranch:  true,
			})
	}
	if err != nil {
		return nil, err
	}

	// verify the cloned repo contains a valid modfile
	if err := i.verifyModFile(repoName, installPath); err != nil {
		return nil, err
	}
	return repo, nil
}

// is this command targetting this mod - i.e. mod was included in the args - or there were no args
func (i *ModInstaller) isCommandTargettingMod(modName string) bool {
	// if there are no args, all mods are targetted
	if len(i.mods) == 0 {
		return true
	}
	// if this mod in the list of updates?
	_, updateMod := i.mods[modName]
	return updateMod
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

func (i *ModInstaller) getModForRequirement(ctx context.Context, requiredModVersion *modconfig.ModVersionConstraint, commandTargettingParent bool) (*DependencyMod, error) {
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
	shouldUpdate, err := i.shouldUpdateMod(installedVersion, requiredModVersion, commandTargettingParent)
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
func (i *ModInstaller) shouldUpdateMod(installedVersion *versionmap.InstalledModVersion, requiredModVersion *modconfig.ModVersionConstraint, commandTargettingParent bool) (bool, error) {
	// so should we update?
	// if commandTargettingParent is set or if the required version constraint is different to the locked version constraint, update

	// does the current version satisfy the required version constraint
	isSatisfied := installedVersion.SatisfiesConstraint(requiredModVersion)

	// is this mod being updated (i.e. is this an update command and this mod was included in the args - or there were no args))
	// commandTargettingParent is set when this is a dependency mod and the parent mod is being updated
	commandTargettingThisMod := i.isCommandTargettingMod(installedVersion.Name) || commandTargettingParent
	if !commandTargettingThisMod {
		// if this command is not targetting this mod, do not update under any circumstances
		return false, nil
	}

	commitCheck := false
	// if this command IS targetting this mod, a[pply the update strategy
	switch i.updateStrategy {
	case constants.ModUpdateFull:
		// 'ModUpdateFull' - check everything for both latest and accuracy
		commitCheck = true

	case constants.ModUpdateLatest:
		// 'ModUpdateLatest' update everything to latest, but only branches - not tags - are commit checked (which is the same as latest)
		// if there is a branch constraint, do a commit check
		if requiredModVersion.Branch() != "" {
			commitCheck = true
		}
	case constants.ModUpdateDevelopment:
		// 'ModUpdateDevelopment' updates branches, file system and broken constraints to latest, leave satisfied constraints unchanged

		// if constraint is not satisfied, update
		if !isSatisfied {
			return true, nil
		}
		// if there is a (satisfied) version constraint, do not update
		if requiredModVersion.VersionConstraint() != nil {
			return false, nil
		}
		// if there is a branch constraint, do a commit check
		if requiredModVersion.Branch() != "" {
			commitCheck = true
		}
		// TODO handle branch/file system update
	case constants.ModUpdateMinimal:
		// 'ModUpdateDevelopment' only updates broken constraints, do not check branches for new commits
		return !isSatisfied, nil
	case constants.ModUpdateNone:
		// no dependency updates
		return false, nil

	}

	// ok to get here, we will see if an update is available.

	// handle file separately??
	if requiredModVersion.FilePath() != "" {
		// TODO how to handle file? hash? modified date
		// just return true for now
		return true, nil
	}

	// if the constraint is a version, check for available versions
	if requiredModVersion.VersionConstraint() != nil {
		// get available versions for this mod
		includePrerelease := requiredModVersion.IsPrerelease()
		availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name, includePrerelease)
		if err != nil {
			return false, err
		}

		updateAvailable, err := i.newerVersionAvailable(requiredModVersion, installedVersion.Version, availableVersions)
		if err != nil {
			return false, err
		}
		if updateAvailable {
			return true, nil
		}
	}

	// do we need to performa a commit check?
	if !commitCheck {
		return false, nil
	}

	return i.newCommitAvailable(installedVersion)
	// if commitCheck flag is set, perform
	return false, nil

}

// determine whether there is a newer mod version avoilable which satisfies the dependency version constraint
func (i *ModInstaller) newerVersionAvailable(requiredVersion *modconfig.ModVersionConstraint, currentVersion *semver.Version, availableVersions versionmap.ResolvedVersionConstraintList) (bool, error) {
	latestVersion, err := i.getModRefSatisfyingVersionConstraint(requiredVersion, availableVersions)
	if err != nil {
		return false, err
	}
	if latestVersion.Version.GreaterThan(currentVersion) {
		return true, nil
	}
	return false, nil
}

func (i *ModInstaller) newCommitAvailable(version *versionmap.InstalledModVersion) (bool, error) {
	var latestCommit string
	var err error

	switch {
	case version.Branch != "":
		latestCommit, err = i.getLatestCommitForBranch(version)
	case version.Version != nil:

		latestCommit, err = i.getLatestCommitForTag(version)
	case version.FilePath != "":
		// TODO CHECK FILE VERSIONS? OR EXPECT TO NEVER GET HERE
		return false, nil
	default:
		err = fmt.Errorf("no version, branch or file path set for installed mod")
	}

	if err != nil {
		return false, err
	}
	// if the latest commit is different to the installed commit, return true
	return latestCommit != version.Commit, nil
}

// get the most recent available mod version which satisfies the version constraint
func (i *ModInstaller) getModRefSatisfyingVersionConstraint(modVersion *modconfig.ModVersionConstraint, availableVersions versionmap.ResolvedVersionConstraintList) (*versionmap.ResolvedVersionConstraint, error) {
	// mod version MUST have a version constrait to be here
	if modVersion.VersionConstraint() == nil {
		return nil, fmt.Errorf("getModRefSatisfyingVersionConstraint should not be called if mod version has no version constraint")
	}

	// find a version which satisfies the version constraint
	var dependencyVersion = getVersionSatisfyingConstraint(modVersion.VersionConstraint(), availableVersions)
	if dependencyVersion == nil {
		return nil, fmt.Errorf("no version of %s found satisfying version constraint: %s", modVersion.Name, modVersion.VersionString)
	}

	return dependencyVersion, nil
}
func (i *ModInstaller) getLatestCommitForBranch(installedVersion *versionmap.InstalledModVersion) (string, error) {
	branch := installedVersion.Branch
	if branch == "" {
		return "", fmt.Errorf("getLatestCommitForBranch called but installed version has no branch")
	}
	// Open the local repository
	modPath := path.Join(i.modsPath, installedVersion.InstallPath)
	repo, err := git.PlainOpen(modPath)
	if err != nil {
		return "", err
	}

	// Fetch the latest changes from the remote for the specific branch
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("+refs/heads/" + branch + ":refs/remotes/origin/" + branch)},
		Force:    true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("error fetching branch '%s' from remote: %w", branch, err)
	}

	// Get the latest commit on the fetched remote branch
	ref, err := repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+branch), true)
	if err != nil {
		return "", fmt.Errorf("error finding reference for branch '%s': %w", branch, err)
	}

	// Return the hash of the latest commit
	return ref.Hash().String(), nil
}

func (i *ModInstaller) getLatestCommitForTag(installedVersion *versionmap.InstalledModVersion) (string, error) {
	// a version must be set to call this function
	if installedVersion.Version == nil {
		return "", fmt.Errorf("getLatestCommitForTag called but installed version has no version")
	}
	// Open the local repository
	modPath := path.Join(i.modsPath, installedVersion.InstallPath)
	repo, err := git.PlainOpen(modPath)
	if err != nil {
		return "", err
	}

	// Fetch only tag information from the remote
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"+refs/tags/*:refs/tags/*"},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("error fetching tags from remote: %w", err)
	}

	// Specify the tag you want to check
	remoteRef, err := repo.Reference(plumbing.ReferenceName(installedVersion.GitRefStr), true)
	if err != nil {
		return "", fmt.Errorf("error finding remote tag: %w", err)
	}

	return remoteRef.Hash().String(), nil
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

func (i *ModInstaller) verifyModFile(name, installPath string) error {
	for _, modFilePath := range app_specific.ModFilePaths(installPath) {
		_, err := os.Stat(modFilePath)
		if err == nil {
			// found the modfile
			return nil
		}
	}
	return sperr.New("mod '%s' does not contain a valid mod file", name)
}
