package load_mod

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
)

// LoadMod parses all hcl files in modPath and returns a single mod
// NOTE: it is an error if there is more than 1 mod defined, however zero mods is acceptable
// - a default mod will be created assuming there are any resource files
func LoadMod(ctx context.Context, modPath string, parseCtx *parse.ModParseContext) (mod *modconfig.Mod, ew error_helpers.ErrorAndWarnings) {
	utils.LogTime(fmt.Sprintf("LoadMod start: %s", modPath))
	defer utils.LogTime(fmt.Sprintf("LoadMod end: %s", modPath))

	defer func() {
		if r := recover(); r != nil {
			ew.Error = helpers.ToError(r)
		}
	}()

	mod, ew = loadModDefinition(modPath, parseCtx)
	if ew.Error != nil {
		return nil, ew
	}

	// if this is a dependency mod, initialise the dependency config
	if parseCtx.DependencyConfig != nil {
		parseCtx.DependencyConfig.SetModProperties(mod)
	}

	// set the current mod on the run context
	if ew.Error = parseCtx.SetCurrentMod(mod); ew.Error != nil {
		return nil, ew
	}

	// load the mod dependencies
	if ew.Error = loadModDependenciesAsync(ctx, mod, parseCtx); ew.Error != nil {
		return nil, ew
	}

	// populate the resource maps of the current mod using the dependency mods
	mod.SetModResources(parseCtx.GetModResources())

	// now load the mod resource hcl (
	var resourceResult error_helpers.ErrorAndWarnings
	mod, resourceResult = loadModResources(ctx, mod, parseCtx)

	ew.Merge(resourceResult)
	return mod, ew
}

func loadModDefinition(modPath string, parseCtx *parse.ModParseContext) (*modconfig.Mod, error_helpers.ErrorAndWarnings) {
	utils.LogTime(fmt.Sprintf("loadModDefinition start: %s", modPath))
	defer utils.LogTime(fmt.Sprintf("loadModDefinition end: %s", modPath))

	var mod *modconfig.Mod
	ew := error_helpers.ErrorAndWarnings{}

	// verify the mod folder exists
	modFilePath, exists := parse.ModFileExists(modPath)
	if exists {
		// load the mod definition to get the dependencies
		var res *parse.DecodeResult
		mod, res = parse.ParseModDefinition(modFilePath, parseCtx.EvalCtx)
		if res.Diags.HasErrors() {
			ew.Error = error_helpers.HclDiagsToError("mod load failed", res.Diags)
			return nil, ew
		}
		// if there are variable dependencies, add warnings
		// this is a workaround to cover the case where a mod require arg block bas a variable dependency
		ew.AddWarning(res.VariableDependenciesToWarnings()...)
	} else {
		// so there is no mod file - should we create a default?
		if !parseCtx.ShouldCreateDefaultMod() {
			ew.Error = perr.BadRequestWithMessage(fmt.Sprintf("mod folder does not contain a mod resource definition '%s'", modPath))
			// ShouldCreateDefaultMod flag NOT set - fail
			return nil, ew
		}
		// just create a default mod
		mod = modconfig.CreateDefaultMod(modPath)
	}
	// add metadata
	// NOTE: set the current mod on the parse context before adding metadata
	parseCtx.CurrentMod = mod
	diags := parse.AddResourceMetadata(mod, mod.GetHclResourceImpl().DeclRange, parseCtx)
	moreEw := error_helpers.DiagsToErrorsAndWarnings("", diags)
	ew.Merge(moreEw)

	return mod, ew
}

func loadModDependenciesAsync(ctx context.Context, parent *modconfig.Mod, parseCtx *parse.ModParseContext) error {
	utils.LogTime(fmt.Sprintf("loadModDependenciesAsync for %s start", parent.GetModPath()))
	defer utils.LogTime(fmt.Sprintf("loadModDependenciesAsync for %s end", parent.GetModPath()))

	parentRequire := parent.GetRequire()
	if parentRequire == nil || len(parentRequire.Mods) == 0 {
		return nil
	}

	// now ensure there is a lock file - if we have any mod dependencies there MUST be a lock file -
	// otherwise 'steampipe install' must be run
	if err := parseCtx.EnsureWorkspaceLock(parent); err != nil {
		return err
	}

	var errors []error
	var wg sync.WaitGroup
	var errChan = make(chan error, len(parentRequire.Mods))

	for _, r := range parentRequire.Mods {
		wg.Add(1)
		go func(requiredModVersion *modconfig.ModVersionConstraint) {
			defer wg.Done()
			if err := loadModDependency(ctx, requiredModVersion, parent, parseCtx); err != nil {
				errChan <- err
			}
		}(r)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		errors = append(errors, err)
	}
	return error_helpers.CombineErrors(errors...)
}

func loadModDependency(ctx context.Context, requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod, parseCtx *parse.ModParseContext) error {
	// get the locked version of this dependency
	modDependency, err := parseCtx.WorkspaceLock.GetLockedModVersion(requiredModVersion, parent)
	if err != nil {
		return err
	}
	if modDependency == nil {
		return perr.BadRequestWithTypeAndMessage(perr.ErrorCodeDependencyFailure, "not all dependencies are installed - run '"+app_specific.AppName+" mod install'")
	}

	utils.LogTime(fmt.Sprintf("loadModDependency for %s %s start", parseCtx.CurrentMod.Name(), modDependency.Name))
	defer utils.LogTime(fmt.Sprintf("loadModDependency for %s %s end", parseCtx.CurrentMod.Name(), modDependency.Name))

	// dependency mods are installed to <mod path>/<mod nam>@version
	// for example workspace_folder/.steampipe/mods/github.com/turbot/steampipe-mod-aws-compliance@v1.0

	// we need to list all mod folder in the parent folder: workspace_folder/.steampipe/mods/github.com/turbot/
	// for each folder we parse the mod name and version and determine whether it meets the version constraint

	// search the parent folder for a mod installation which satisfied the given mod dependency
	dependencyDir, err := parseCtx.WorkspaceLock.FindInstalledDependency(modDependency.ResolvedVersionConstraint)
	if err != nil {
		return err
	}
	// create a parse context for the dependency mod
	childParseCtx, err := parse.NewChildModParseContext(parseCtx, modDependency.ResolvedVersionConstraint, dependencyDir)
	if err != nil {
		return err
	}

	// NOTE: pass in the version and dependency path of the mod - these must be set before it loads its dependencies
	dependencyMod, errAndWarnings := LoadMod(ctx, dependencyDir, childParseCtx)
	if errAndWarnings.GetError() != nil {
		return errAndWarnings.GetError()
	}

	// update loaded dependency mods
	parseCtx.AddLoadedDependencyMod(dependencyMod)

	return nil
}

func loadModResources(ctx context.Context, mod *modconfig.Mod, parseCtx *parse.ModParseContext) (*modconfig.Mod, error_helpers.ErrorAndWarnings) {
	utils.LogTime(fmt.Sprintf("loadModResources %s start", mod.GetModPath()))
	defer utils.LogTime(fmt.Sprintf("loadModResources %s end", mod.GetModPath()))

	// get the source files
	sourcePaths, err := getSourcePaths(ctx, mod.GetModPath(), parseCtx.ListOptions)
	if err != nil {
		slog.Warn("LoadMod: failed to get mod file paths", "error", err)
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// load the raw file data
	fileData, diags := parse.LoadFileData(sourcePaths...)
	if diags.HasErrors() {
		return nil, error_helpers.NewErrorsAndWarning(error_helpers.HclDiagsToError("Failed to load all mod files", diags))
	}

	// parse all hcl files (NOTE - this reads the CurrentMod out of ParseContext and adds to it)
	mod, errAndWarnings := parse.ParseMod(ctx, fileData, parseCtx)

	return mod, errAndWarnings
}

// GetModFileExtensions returns list of all file extensions we care about
// this will be the mod data extension, plus any registered extensions registered in fileToResourceMap
func GetModFileExtensions() []string {
	return append(app_specific.ModDataExtensions, app_specific.VariablesExtensions...)
}

// build list of all filepaths we need to parse/load the mod
// this will include hcl files (with .sp extension)
// as well as any other files with extensions that have been registered for pseudo resource creation
// (see steampipeconfig/modconfig/resource_type_map.go)
func getSourcePaths(ctx context.Context, modPath string, listOpts filehelpers.ListOptions) ([]string, error) {
	sourcePaths, err := filehelpers.ListFilesWithContext(ctx, modPath, &listOpts)
	if err != nil {
		return nil, err
	}
	return sourcePaths, nil
}

// Deprecated
// TODO this function is included for backwards compatibility - it is used for Flowpipe LoadPipelines
func LoadModWithFileName(ctx context.Context, modPath, modFile string, parseCtx *parse.ModParseContext) (mod *modconfig.Mod, errorsAndWarnings error_helpers.ErrorAndWarnings) {
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = error_helpers.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	mod, loadModResult := loadModDefinitionWithFileName(modPath, modFile, parseCtx)
	if loadModResult.Error != nil {
		return nil, loadModResult
	}

	// if this is a dependency mod, initialise the dependency config
	if parseCtx.DependencyConfig != nil {
		parseCtx.DependencyConfig.SetModProperties(mod)
	}

	// set the current mod on the run context
	if err := parseCtx.SetCurrentMod(mod); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// load the mod dependencies
	if err := loadModDependenciesAsync(ctx, mod, parseCtx); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// populate the resource maps of the current mod using the dependency mods
	mod.SetModResources(parseCtx.GetModResources())
	// now load the mod resource hcl (
	mod, errorsAndWarnings = loadModResources(ctx, mod, parseCtx)

	// add in any warnings from mod load
	errorsAndWarnings.AddWarning(loadModResult.Warnings...)
	return mod, errorsAndWarnings
}

func loadModDefinitionWithFileName(modPath, modFileName string, parseCtx *parse.ModParseContext) (mod *modconfig.Mod, errorsAndWarnings error_helpers.ErrorAndWarnings) {
	modFilePath := filepath.Join(modPath, modFileName)

	// only create transient local mod if the mod file does not exist
	filehelpers.FileExists(modFilePath)
	modFileFound := filehelpers.FileExists(modFilePath)
	if parseCtx.ShouldCreateDefaultMod() && !modFileFound {
		mod = modconfig.NewMod("local", modPath, hcl.Range{})
		return mod, errorsAndWarnings
	}

	if modFileFound {
		// load the mod definition to get the dependencies
		var res *parse.DecodeResult
		mod, res = parse.ParseModDefinition(modFilePath, parseCtx.EvalCtx)
		if res.Diags.HasErrors() {
			return nil, error_helpers.DiagsToErrorsAndWarnings("mod load failed", res.Diags)
		}
	} else {
		// so there is no mod file - should we create a default?
		if !parseCtx.ShouldCreateDefaultMod() {
			errorsAndWarnings.Error = perr.BadRequestWithMessage(fmt.Sprintf("mod folder does not contain a mod resource definition '%s'", modPath))
			// ShouldCreateDefaultMod flag NOT set - fail
			return nil, errorsAndWarnings
		}
		// just create a default mod
		mod = modconfig.CreateDefaultMod(modPath)

	}
	return mod, errorsAndWarnings
}
