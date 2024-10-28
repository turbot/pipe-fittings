package modinstaller

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/plugin"
	"github.com/turbot/pipe-fittings/sperr"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/versionmap"
)

type ModInstaller struct {
	installData *InstallData

	// this will be updated as changes are made to dependencies
	workspaceMod *modconfig.Mod

	// since changes are made to workspaceMod, we need a copy of the Require as is on disk
	// to be able to calculate changes
	oldRequire *modconfig.Require

	targetMods map[string]*modconfig.ModVersionConstraint

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
	// TODO why does powerpipe care about plugins???
	// optional map of installed plugin versions
	pluginVersions *plugin.PluginVersionMap

	updateStrategy string
}

func NewModInstaller(opts *InstallOpts) (*ModInstaller, error) {
	if opts.WorkspaceMod == nil {
		return nil, fmt.Errorf("no workspace mod passed to mod installer")
	}
	// NOTE: ensure worksapace path is absolute
	workspacePath, err := filepath.Abs(opts.WorkspaceMod.GetModPath())
	if err != nil {
		return nil, err
	}
	i := &ModInstaller{
		workspacePath: workspacePath,
		workspaceMod:  opts.WorkspaceMod,
		command:       opts.Command,
		dryRun:        opts.DryRun,
		force:         opts.Force,
		// TODO why does powerpipe care about plugins???
		pluginVersions: opts.PluginVersions,
		updateStrategy: opts.UpdateStrategy,
	}

	if require := opts.WorkspaceMod.GetRequire(); require != nil {
		i.oldRequire = require.Clone()
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
	i.targetMods = requiredMods

	return i, nil
}

func (i *ModInstaller) UninstallWorkspaceDependencies(ctx context.Context) error {
	workspaceMod := i.workspaceMod

	// remove required dependencies from the mod file
	if len(i.targetMods) == 0 {
		workspaceMod.RemoveAllModDependencies()
	} else {
		workspaceMod.RemoveModDependencies(i.targetMods)
	}

	// uninstall by calling Install
	if err := i.installMods(ctx, workspaceMod); err != nil {
		return err
	}

	if workspaceMod.GetRequire().Empty() {
		workspaceMod.SetRequire(nil)
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
			slog.Debug("suppressing error in InstallWorkspaceDependencies because force is enabled", "error", err)
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
	if len(i.targetMods) > 0 {
		workspaceMod.AddModDependencies(i.targetMods)
	}

	if err := i.installMods(ctx, workspaceMod); err != nil {
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
	var removeErrors []error
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

func (i *ModInstaller) installMods(ctx context.Context, parent *modconfig.Mod) (err error) {
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

	for _, requiredModVersion := range i.workspaceMod.GetRequire().Mods {
		// is this mod targeted by the command (i.e. was the mod name passed as an arg
		// - or else were no args passed, targeting all mods)
		commandTargetingMod := i.isCommandTargetingMod(requiredModVersion)

		// do we have this mod installed for any parent?
		currentMod, err := i.getModForRequirement(ctx, requiredModVersion, commandTargetingMod)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		if err := i.installModDependenciesRecursively(ctx, requiredModVersion, currentMod, parent, commandTargetingMod); err != nil {
			errors = append(errors, err)
		}
	}

	// update the lock to be the new lock, and record any uninstalled mods
	if len(errors) == 0 {
		err = i.installData.onInstallComplete()
		if err != nil {
			errors = append(errors, err)
		}
	}
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

	// if dependencyMod is nil, this means it is not installed. If this mod or it's parent is one of the targets, we must install it
	if dependencyMod == nil {
		if commandTargettingParent {
			var err error
			dependencyMod, err = i.install(ctx, requiredModVersion, parent)
			if err != nil {
				return err
			}
		} else {
			// nothing further to do
			return nil
		}
	} else {
		// this mod is already installed - just update the install data
		i.installData.addExisting(dependencyMod, parent)
		slog.Debug(fmt.Sprintf("not installing %s with version constraint %s as version %s is already Installed", requiredModVersion.Name, requiredModVersion.VersionString, dependencyMod.Mod.GetVersion()))
	}

	// to get here we have the dependency mod - either we installed it or it was already installed
	// recursively install its dependencies
	for _, childDependency := range dependencyMod.Mod.GetRequire().Mods {
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

	if len(errors) > 0 {
		return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("%d child %s failed to install", len(errors), utils.Pluralize("dependency", len(errors))), errors...)
	}
	return nil
}

func (i *ModInstaller) install(ctx context.Context, requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (installedMod *DependencyMod, err error) {
	var modDef *modconfig.Mod
	var resolvedRef *versionmap.ResolvedVersionConstraint

	defer func() {
		if err == nil {
			i.installData.onModInstalled(installedMod, parent)
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
		modDef, err = i.installFromTag(resolvedRef)
		if err != nil {
			return nil, err
		}
	case requiredModVersion.Tag != "":
		// get a resolved mod ref that satisfies the version constraints
		resolvedRef, err = i.getModRefForTag(requiredModVersion)
		if err != nil {
			return nil, err
		}
		// install the mod
		modDef, err = i.installFromTag(resolvedRef)
		if err != nil {
			return nil, err
		}
	case requiredModVersion.BranchName != "":
		resolvedRef, modDef, err = i.installFromBranch(ctx, requiredModVersion)
		if err != nil {
			return nil, err
		}

	case requiredModVersion.FilePath != "":
		resolvedRef, modDef, err = i.installFromFilepath(ctx, requiredModVersion, parent)
		if err != nil {
			return nil, err
		}
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
		Alias:                     modDef.GetShortName(),
	}

	return &DependencyMod{
		Mod:              modDef,
		InstalledVersion: installedModVersion,
	}, nil

}

// install a mod
func (i *ModInstaller) installFromTag(dependency *versionmap.ResolvedVersionConstraint) (*modconfig.Mod, error) {
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
			return nil, fmt.Errorf("could not remove existing mod path '%s': %w", destPath, err)
		}
	}

	slog.Debug("installing", "dependency", dependencyPath, "in", destPath)
	if _, err := i.installFromGit(dependency.Name, plumbing.ReferenceName(dependency.GitRefStr), destPath); err != nil {
		return nil, err
	}

	// now load the installed mod and return it
	modDef, err := parse.LoadModfile(destPath)
	if err != nil {
		return nil, err
	}
	if modDef == nil {
		return nil, fmt.Errorf("'%s' has no mod definition file", dependency.Name)
	}
	return modDef, nil
}

func (i *ModInstaller) installFromBranch(_ context.Context, modVersion *modconfig.ModVersionConstraint) (*versionmap.ResolvedVersionConstraint, *modconfig.Mod, error) {
	// build a DependencyVersion
	var dependencyVersion = &modconfig.DependencyVersion{
		Branch: modVersion.BranchName,
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
			return nil, nil, fmt.Errorf("could not remove existing mod path '%s': %w", destPath, err)
		}
	}

	slog.Debug("installing", "dependency", dependencyPath, "in", destPath)
	// build a git ref for the branch
	gitRef := plumbing.NewBranchReferenceName(dependencyVersion.Branch)
	repo, err := i.installFromGit(modVersion.Name, gitRef, destPath)
	if err != nil {
		return nil, nil, err
	}
	// get the commit hash
	ref, err := repo.Reference(gitRef, true)
	if err != nil {
		return nil, nil, err
	}

	// build a ResolvedVersionConstraint
	resolvedRef := versionmap.NewResolvedVersionConstraint(dependencyVersion, modVersion.Name, ref)

	// now load the installed mod and return it
	modDef, err := parse.LoadModfile(destPath)
	if err != nil {
		return nil, nil, err
	}
	if modDef == nil {
		return nil, nil, fmt.Errorf("'%s' has no mod definition file", modVersion.Name)
	}

	return resolvedRef, modDef, nil
}

func (i *ModInstaller) installFromGit(repoName string, gitRefName plumbing.ReferenceName, installPath string) (*git.Repository, error) {
	// get the mod from git = first try https
	gitUrl := getGitUrl(repoName, GitUrlModeHTTPS)
	slog.Debug("installFromGit cloning the repo", gitUrl, gitRefName.String())

	repo, err := i.cloneRepo(gitUrl, gitRefName, installPath)

	if err != nil {
		// if that failed, try ssh
		gitUrl := getGitUrl(repoName, GitUrlModeSSH)
		slog.Debug(">>> cloning", gitUrl, gitRefName.String())
		repo, err = i.cloneRepo(gitUrl, gitRefName, installPath)
		if err != nil {
			return nil, err
		}
	}

	// verify the cloned repo contains a valid modfile
	if err := i.verifyModFile(repoName, installPath); err != nil {
		return nil, err
	}
	return repo, nil
}

func (i *ModInstaller) installFromFilepath(_ context.Context, modVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod) (*versionmap.ResolvedVersionConstraint, *modconfig.Mod, error) {
	// build a DependencyVersion
	// convert the filename to absolute
	filePath := i.toAbsoluteFilepath(modVersion.FilePath, parent.GetModPath())

	var dependencyVersion = &modconfig.DependencyVersion{
		FilePath: filePath,
	}

	slog.Debug("installing a local file mod", "file location", filePath)

	// now load the installed mod and return it
	modDef, err := parse.LoadModfile(dependencyVersion.FilePath)
	if err != nil {
		return nil, nil, err
	}
	if modDef == nil {
		return nil, nil, fmt.Errorf("'%s' has no mod definition file", modVersion.Name)
	}

	// build a ResolvedVersionConstraint
	// NOTE: use the file path as the name
	resolvedRef := versionmap.NewResolvedVersionConstraint(dependencyVersion, filePath, nil)

	return resolvedRef, modDef, nil
}

// is this command targeting this mod - i.e. mod was included in the args
func (i *ModInstaller) isCommandTargetingMod(m *modconfig.ModVersionConstraint) bool {
	if len(i.targetMods) == 0 {
		return true
	}
	// if this mod in the list of updates?
	constraint, isTarget := i.targetMods[m.Name]
	return isTarget && constraint.Equals(m)
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
	var modDefinition *modconfig.Mod

	// construct the dependency path - this is the relative path of the dependency we are installing
	dependencyPath := modVersion.DependencyPath()

	var err error
	// if the mod has a FilePath, just load it
	if modVersion.DependencyVersion.FilePath != "" {
		modDefinition, err = parse.LoadModfile(modVersion.DependencyVersion.FilePath)
		if err != nil {
			return nil, err
		}
	} else {
		// first try loading from the shadow dir
		modDefinition, err = i.loadDependencyModFromRoot(ctx, i.shadowDirPath, dependencyPath)
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
	// user non-method, injecting updateChecker interface to make unit testing easier
	return shouldUpdateMod(installedVersion, requiredModVersion, commandTargettingParent, i)
}

func (i *ModInstaller) getUpdateStrategy() string {
	return i.updateStrategy
}

// determine whether there is a newer mod version available which satisfies the dependency version constraint
func (i *ModInstaller) newerVersionAvailable(requiredVersion *modconfig.ModVersionConstraint, currentVersion *semver.Version) (bool, error) {
	// get available versions for this mod
	includePrerelease := requiredVersion.IsPrerelease()
	availableVersions, err := i.installData.getAvailableModVersions(requiredVersion.Name, includePrerelease)
	if err != nil {
		return false, err
	}

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
	case version.Version != nil, version.Tag != "":
		latestCommit, err = i.getLatestCommitForTag(version)
	case version.FilePath != "":
		// TODO CHECK FILE VERSIONS? OR EXPECT TO NEVER GET HERE
		return false, nil
	default:
		err = fmt.Errorf("no version, branch or file path set for Installed mod")
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

// get the most recent available mod version which satisfies the version constraint
func (i *ModInstaller) getModRefForTag(modVersion *modconfig.ModVersionConstraint) (*versionmap.ResolvedVersionConstraint, error) {
	// mod version MUST have a version constrait to be here
	if modVersion.Tag == "" {
		return nil, fmt.Errorf("getModRefForTag should not be called if mod version has no tag")
	}

	// find a version which satisfies the version constraint
	dependencyVersion, err := getTagFromGit(modVersion.Name, modVersion.Tag)
	if err != nil {
		return nil, err
	}
	if dependencyVersion == nil {
		return nil, fmt.Errorf("tag %s not found for mod %s", modVersion.Tag, modVersion.Name)
	}

	return dependencyVersion, nil
}

func (i *ModInstaller) getLatestCommitForBranch(installedVersion *versionmap.InstalledModVersion) (string, error) {
	branch := installedVersion.Branch
	if branch == "" {
		return "", fmt.Errorf("getLatestCommitForBranch called but Installed version has no branch")
	}
	// Open the local repository
	modPath := path.Join(i.modsPath, installedVersion.DependencyPath())
	repo, err := i.openRepo(modPath)
	if err != nil {
		return "", err
	}

	// Fetch the latest changes from the remote for the specific branch
	fetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("+refs/heads/" + branch + ":refs/remotes/origin/" + branch)},
		Force:    true,
	}

	if gitHubToken := getGitToken(); gitHubToken != "" {
		// if we have a token, use it
		fetchOptions.Auth = getGitAuthForToken(gitHubToken)
	}
	err = repo.Fetch(fetchOptions)
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
	// a version or tag must be set to call this function
	if installedVersion.Version == nil && installedVersion.Tag == "" {
		return "", fmt.Errorf("getLatestCommitForTag called but Installed version has no version or tag")
	}

	// Open the local repository
	modPath := path.Join(i.modsPath, installedVersion.DependencyPath())
	repo, err := i.openRepo(modPath)
	if err != nil {
		return "", err
	}

	// Fetch only tag information from the remote
	// Fetch the latest changes from the remote for the specific branch
	fetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{"+refs/tags/*:refs/tags/*"},
	}

	if gitHubToken := getGitToken(); gitHubToken != "" {
		// if we have a token, use it
		fetchOptions.Auth = getGitAuthForToken(gitHubToken)
	}
	err = repo.Fetch(fetchOptions)
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

// build the path of the temp location to copy this dependency to
func (i *ModInstaller) getDependencyDestPath(dependencyFullName string) string {
	return filepath.Join(i.modsPath, dependencyFullName)
}

// build the path of the temp location to copy this dependency to
func (i *ModInstaller) getDependencyShadowPath(dependencyFullName string) string {
	return filepath.Join(i.shadowDirPath, dependencyFullName)
}

// set the mod dependency path
func (i *ModInstaller) setModDependencyConfig(mod *modconfig.Mod, dependencyPath string) error {
	return mod.SetDependencyConfigFromPath(dependencyPath)
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

// is the given string a file path, and if so, return as an absolute path, realtive to the given base
func (i *ModInstaller) toAbsoluteFilepath(modArg, basePath string) string {
	filePath := modArg
	// Check if the path is already absolute
	if !filepath.IsAbs(filePath) {
		// If it's not absolute, join it with the workspace path
		filePath = filepath.Join(basePath, filePath)
	}

	if filehelpers.DirectoryExists(filePath) {
		return filePath
	}
	return ""
}
