package workspace

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/versionmap"
	"github.com/zclconf/go-cty/cty"
)

type Workspace struct {
	Path                string
	ModInstallationPath string
	Mod                 *modconfig.Mod

	PipelingConnections map[string]connection.PipelingConnection

	Mods map[string]*modconfig.Mod

	// the input variables used in the parse
	VariableValues map[string]string

	// items from the global config which need to be added to the parse context as a value map
	// Flowpipe uses this to populate notifiers
	// it is a map of cty value maps - keyed by the typ ename (e.g. notifier)
	configValueMaps map[string]map[string]cty.Value

	watcher     *filewatcher.FileWatcher
	loadLock    *sync.Mutex
	exclusions  []string
	modFilePath string

	FileWatcherErrorHandler func(context.Context, error)
	WatcherError            error
	// callback function called when there is a file watcher event
	onFileWatcherEventMessages func()

	// hooks
	OnFileWatcherError  func(context.Context, error)
	OnFileWatcherEvent  func(context.Context, modconfig.ModResources, modconfig.ModResources)
	BlockTypeInclusions []string
	ValidateVariables   bool
	SupportLateBinding  bool
	decoderOptions      []parse.DecoderOption
}

func (w *Workspace) SetupWatcher(ctx context.Context, errorHandler func(context.Context, error)) error {
	watcherOptions := &filewatcher.WatcherOptions{
		Directories: []string{w.Path},
		Include:     filehelpers.InclusionsFromExtensions(load_mod.GetModFileExtensions()),
		Exclude:     w.exclusions,
		ListFlag:    filehelpers.FilesRecursive,
		EventMask:   fsnotify.Create | fsnotify.Remove | fsnotify.Rename | fsnotify.Write,
		// we should look into passing the callback function into the underlying watcher
		// we need to analyze the kind of errors that come out from the watcher and
		// decide how to handle them
		// OnError: errCallback,
		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(ctx)
		},
	}
	watcher, err := filewatcher.NewWatcher(watcherOptions)
	if err != nil {
		return err
	}
	w.watcher = watcher
	// start the watcher
	watcher.Start()

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	w.FileWatcherErrorHandler = errorHandler

	return nil
}

func (w *Workspace) SetOnFileWatcherEventMessages(f func()) {
	w.onFileWatcherEventMessages = f
}

func (w *Workspace) Close() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}

func (w *Workspace) ModfileExists() bool {
	return len(w.modFilePath) > 0
}

// check  whether the workspace contains a modfile
// this will determine whether we load files recursively, and create pseudo resources for sql files
func (w *Workspace) SetModfileExists() {
	modFile, err := FindModFilePath(w.Path)
	modFileExists := !errors.Is(err, ErrorNoModDefinition)

	if modFileExists {
		w.modFilePath = modFile

		// also set it in the viper config, so that it is available to whoever is using it
		viper.Set(constants.ArgModLocation, filepath.Dir(modFile))
		w.Path = filepath.Dir(modFile)
		w.Mod.SetFilePath(modFile)
	}
}

func (w *Workspace) LoadWorkspaceMod(ctx context.Context) error_helpers.ErrorAndWarnings {
	utils.LogTime("LoadWorkspaceMod start")
	defer utils.LogTime("LoadWorkspaceMod end")

	// check if your workspace path is home dir and if modfile exists - if yes then warn and ask user to continue or not
	if err := HomeDirectoryModfileCheck(ctx, w.Path); err != nil {
		return error_helpers.NewErrorsAndWarning(err)
	}

	// resolve values of all input variables and add to parse context
	// we WILL validate missing variables when loading
	// NOTE: this does an initial mod load, loading only variable blocks
	inputVariables, ew := w.resolveVariableValues(ctx)
	if ew.Error != nil {
		return ew
	}
	// build run context which we use to load the workspace
	parseCtx, err := w.GetParseContext(ctx)
	if err != nil {
		ew.Error = err
		return ew
	}

	// add evaluated variables to the context
	parseCtx.AddInputVariableValues(inputVariables)

	// if we are ONLY loading variables, we can skip loading resources
	if parseCtx.LoadVariablesOnly() {
		return w.populateVariablesOnlyMod(parseCtx)
	}

	// do not reload variables or mod block, as we already have them
	parseCtx.SetBlockTypeExclusions(schema.BlockTypeVariable, schema.BlockTypeMod)
	if len(w.BlockTypeInclusions) > 0 {
		parseCtx.SetBlockTypes(w.BlockTypeInclusions...)
	}

	// load the workspace mod
	m, otherErrorAndWarning := load_mod.LoadMod(ctx, w.Path, parseCtx)
	ew.Merge(otherErrorAndWarning)
	if ew.Error != nil {
		return ew
	}

	// set the mod
	w.Mod = m
	// set the child mods
	w.Mods = parseCtx.GetTopLevelDependencyMods()
	// NOTE: add in the workspace mod to the dependency mods
	w.Mods[w.Mod.Name()] = w.Mod

	return ew
}

// resolve values of all input variables
// we may need to load the mod more than once to resolve all variable dependencies
func (w *Workspace) resolveVariableValues(ctx context.Context) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	lastDependCount := -1

	var inputVariables *modconfig.ModVariableMap
	var ew error_helpers.ErrorAndWarnings

	for {
		variablesParseCtx, ew := w.getVariablesParseContext(ctx, inputVariables)
		if ew.Error != nil {
			return nil, ew
		}

		utils.LogTime("getInputVariables start")

		var otherEw error_helpers.ErrorAndWarnings
		inputVariables, otherEw = w.getVariableValues(ctx, variablesParseCtx, w.ValidateVariables)
		utils.LogTime("getInputVariables end")
		ew.Merge(otherEw)
		if ew.Error != nil {
			slog.Error("Error loading input variables", "error", ew.Error)
			return nil, ew
		}

		// populate the parsed variable values
		w.VariableValues, ew.Error = inputVariables.GetPublicVariableValues()
		if ew.Error != nil {
			return nil, error_helpers.ErrorAndWarnings{}
		}

		// do we have any variable dependencies? If so there will be warnings
		dependCount := getVariableDependencyCount(ew)
		if dependCount == 0 {
			break
		}
		if dependCount == lastDependCount {
			slog.Warn("Failed to resolve all variable dependencies")
			break
		}

		lastDependCount = dependCount
	}

	return inputVariables, ew
}

func getVariableDependencyCount(ew error_helpers.ErrorAndWarnings) int {
	count := 0
	for _, w := range ew.Warnings {
		if strings.Contains(w, constants.MissingVariableWarning) {
			count++
		}
	}
	return count
}

func (w *Workspace) getVariablesParseContext(ctx context.Context, inputVariable *modconfig.ModVariableMap) (*parse.ModParseContext, error_helpers.ErrorAndWarnings) {
	// build a run context just to use to load variable definitions
	variablesParseCtx, err := w.GetParseContext(ctx)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}
	// only load variables blocks
	variablesParseCtx.SetBlockTypes(schema.BlockTypeVariable)
	// NOTE: exclude mod block as we have already loaded the mod definition
	variablesParseCtx.SetBlockTypeExclusions(schema.BlockTypeMod)
	// add the connections and notifiers to the eval context
	variablesParseCtx.SetIncludeLateBindingResources(true)

	if inputVariable != nil {
		variablesParseCtx.AddInputVariableValues(inputVariable)
	}
	return variablesParseCtx, error_helpers.ErrorAndWarnings{}
}

func (w *Workspace) getVariableValues(ctx context.Context, variablesParseCtx *parse.ModParseContext, validateMissing bool) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	utils.LogTime("getInputVariables start")
	defer utils.LogTime("getInputVariables end")

	// load variable definitions
	variableMap, ew := load_mod.LoadVariableDefinitions(ctx, w.Path, variablesParseCtx)
	if ew.Error != nil {
		return nil, ew
	}
	// get the values
	m, moreEw := load_mod.GetVariableValues(variablesParseCtx, variableMap, validateMissing)
	ew.Merge(moreEw)
	return m, ew
}

func (w *Workspace) GetParseContext(ctx context.Context) (*parse.ModParseContext, error) {
	workspaceLock, err := w.loadWorkspaceLock(ctx)
	if err != nil {
		return nil, err
	}
	listOptions := filehelpers.ListOptions{
		Flags:   filehelpers.FilesRecursive,
		Exclude: w.exclusions,
		// load files specified by inclusions
		Include: filehelpers.InclusionsFromExtensions(app_specific.ModDataExtensions),
	}

	parseCtx, err := parse.NewModParseContext(workspaceLock, w.Path,
		parse.WithParseFlags(parse.CreateDefaultMod),
		parse.WithListOptions(listOptions),
		parse.WithConnections(w.PipelingConnections),
		parse.WithLateBinding(w.SupportLateBinding),
		parse.WithConfigValueMap(w.configValueMaps),
		parse.WithDecoderOptions(w.decoderOptions...))

	if err != nil {
		return nil, err
	}

	return parseCtx, nil
}

// load the workspace lock, migrating it if necessary
func (w *Workspace) loadWorkspaceLock(ctx context.Context) (*versionmap.WorkspaceLock, error) {
	workspaceLock, err := versionmap.LoadWorkspaceLock(w.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load installation cache from %s: %s", w.Path, err)
	}

	// if this is the old format, migrate by reinstalling dependencies
	if workspaceLock.StructVersion() != versionmap.WorkspaceLockStructVersion {
		return nil, fmt.Errorf("workspace lock file is out of date, please run 'steampipe install' to update")
	}

	return workspaceLock, nil
}

func (w *Workspace) LoadExclusions() error {
	// default to ignoring hidden files and folders
	w.exclusions = []string{
		// ignore any hidden folder
		fmt.Sprintf("%s/.*", w.Path),
		// and sub files/folders of hidden folders
		fmt.Sprintf("%s/.*/**", w.Path),
	}

	ignorePath := filepath.Join(w.Path, app_specific.WorkspaceIgnoreFile)
	file, err := os.Open(ignorePath)
	if err != nil {
		// if file does not exist, just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) != 0 && !strings.HasPrefix(line, "#") {
			// add exclusion to the workspace path (to ensure relative patterns work)
			absoluteExclusion := filepath.Join(w.Path, line)
			w.exclusions = append(w.exclusions, absoluteExclusion)
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

// populate the mod resource maps with variables from the parse context
func (w *Workspace) populateVariablesOnlyMod(parseCtx *parse.ModParseContext) error_helpers.ErrorAndWarnings {
	var diags hcl.Diagnostics
	for _, v := range parseCtx.Variables.ToArray() {
		diags = append(diags, w.Mod.GetModResources().AddResource(v)...)
	}
	return error_helpers.DiagsToErrorsAndWarnings("", diags)
}

func (w *Workspace) LoadLock() {
	if w.loadLock == nil {
		w.loadLock = &sync.Mutex{}
	}
	w.loadLock.Lock()
}

func (w *Workspace) LoadUnlock() {
	w.loadLock.Unlock()
}

// GetModResources implements ModResourcesProvider
func (w *Workspace) GetModResources() modconfig.ModResources {

	w.LoadLock()
	defer w.LoadUnlock()

	return w.Mod.GetModResources()
}

func (w *Workspace) GetResource(parsedName *modconfig.ParsedResourceName) (resource modconfig.HclResource, found bool) {
	return w.GetModResources().GetResource(parsedName)
}
