package workspace

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/credential"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/modinstaller"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/versionmap"
)

type Workspace struct {
	Path                string
	ModInstallationPath string
	Mod                 *modconfig.Mod

	Mods map[string]*modconfig.Mod

	// the input variables used in the parse
	VariableValues map[string]string

	// Credentials are something different, it's not part of the mod, it's not part of the workspace, it is at the same level
	// with mod and workspace. However, it can be referenced by the mod, so it needs to be in the parse context
	Credentials  map[string]credential.Credential
	Integrations map[string]modconfig.Integration
	Notifiers    map[string]modconfig.Notifier

	CloudMetadata *steampipeconfig.CloudMetadata

	// source snapshot paths
	// if this is set, no other mod resources are loaded and
	// the ResourceMaps returned by GetModResources will contain only the snapshots
	SourceSnapshots []string

	watcher     *filewatcher.FileWatcher
	loadLock    sync.Mutex
	exclusions  []string
	modFilePath string

	fileWatcherErrorHandler func(context.Context, error)
	watcherError            error
	// callback function called when there is a file watcher event
	onFileWatcherEventMessages func()

	// hooks
	OnFileWatcherError  func(context.Context, error)
	OnFileWatcherEvent  func(context.Context, *modconfig.ResourceMaps, *modconfig.ResourceMaps)
	BlockTypeInclusions []string
}

// Load_ creates a Workspace and loads the workspace mod

func Load(ctx context.Context, workspacePath string, opts ...LoadWorkspaceOption) (w *Workspace, ew error_helpers.ErrorAndWarnings) {
	cfg := newLoadWorkspaceConfig()
	for _, o := range opts {
		o(cfg)
	}

	utils.LogTime("w.Load_ start")
	defer utils.LogTime("w.Load_ end")

	w, err := createShellWorkspace(workspacePath)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	w.Credentials = cfg.credentials
	w.Integrations = cfg.integrations
	w.Notifiers = cfg.notifiers
	w.BlockTypeInclusions = cfg.blockTypeInclusions

	if !w.ModfileExists() {
		// just return the shell workspace
		return w, ew
	}
	// load the w mod
	errAndWarnings := w.loadWorkspaceMod(ctx)
	return w, errAndWarnings
}

func createShellWorkspace(workspacePath string) (*Workspace, error) {
	// create shell workspace
	workspace := &Workspace{
		Path:           workspacePath,
		VariableValues: make(map[string]string),
	}

	// check whether the workspace contains a modfile
	// this will determine whether we load files recursively, and create pseudo resources for sql files
	workspace.setModfileExists()

	// load the .steampipe ignore file
	if err := workspace.loadExclusions(); err != nil {
		return nil, err
	}

	return workspace, nil
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
	w.fileWatcherErrorHandler = errorHandler

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
func (w *Workspace) setModfileExists() {
	modFile, err := FindModFilePath(w.Path)
	modFileExists := !errors.Is(err, ErrorNoModDefinition)

	if modFileExists {
		w.modFilePath = modFile

		// also set it in the viper config, so that it is available to whoever is using it
		viper.Set(constants.ArgModLocation, filepath.Dir(modFile))
		w.Path = filepath.Dir(modFile)
	}
}

func (w *Workspace) loadWorkspaceMod(ctx context.Context) error_helpers.ErrorAndWarnings {
	// check if your workspace path is home dir and if modfile exists - if yes then warn and ask user to continue or not
	if err := HomeDirectoryModfileCheck(ctx, w.Path); err != nil {
		return error_helpers.NewErrorsAndWarning(err)
	}

	// resolve values of all input variables
	// we WILL validate missing variables when loading
	// NOTE: this does an initial mod load, loading only variable blocks
	validateMissing := true
	inputVariables, errorsAndWarnings := w.getInputVariables(ctx, validateMissing)
	if errorsAndWarnings.Error != nil {
		return errorsAndWarnings
	}
	// populate the parsed variable values
	w.VariableValues, errorsAndWarnings.Error = inputVariables.GetPublicVariableValues()
	if errorsAndWarnings.Error != nil {
		return errorsAndWarnings
	}

	// build run context which we use to load the workspace
	parseCtx, err := w.getParseContext(ctx)
	if err != nil {
		errorsAndWarnings.Error = err
		return errorsAndWarnings
	}

	// add evaluated variables to the context
	parseCtx.AddInputVariableValues(inputVariables)
	// do not reload variables as we already have them
	parseCtx.BlockTypeExclusions = []string{schema.BlockTypeVariable}
	if len(w.BlockTypeInclusions) > 0 {
		parseCtx.BlockTypes = w.BlockTypeInclusions
	}
	// load the workspace mod
	m, otherErrorAndWarning := load_mod.LoadMod(ctx, w.Path, parseCtx)
	errorsAndWarnings.Merge(otherErrorAndWarning)
	if errorsAndWarnings.Error != nil {
		return errorsAndWarnings
	}

	// now set workspace properties
	// populate the mod references map references
	m.ResourceMaps.PopulateReferences()

	// set the mod
	w.Mod = m
	// set the child mods
	w.Mods = parseCtx.GetTopLevelDependencyMods()
	// NOTE: add in the workspace mod to the dependency mods
	w.Mods[w.Mod.Name()] = w.Mod

	// verify all runtime dependencies can be resolved
	errorsAndWarnings.Error = w.verifyResourceRuntimeDependencies()

	return errorsAndWarnings
}

func (w *Workspace) getInputVariables(ctx context.Context, validateMissing bool) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	// build a run context just to use to load variable definitions
	variablesParseCtx, err := w.getParseContext(ctx)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	return w.getVariableValues(ctx, variablesParseCtx, validateMissing)
}

func (w *Workspace) getVariableValues(ctx context.Context, variablesParseCtx *parse.ModParseContext, validateMissing bool) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	// load variable definitions
	variableMap, err := load_mod.LoadVariableDefinitions(ctx, w.Path, variablesParseCtx)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}
	// get the values
	return load_mod.GetVariableValues(variablesParseCtx, variableMap, validateMissing)
}

// build options used to load workspace
func (w *Workspace) getParseContext(ctx context.Context) (*parse.ModParseContext, error) {
	workspaceLock, err := w.loadWorkspaceLock(ctx)
	if err != nil {
		return nil, err
	}

	parseCtx := parse.NewModParseContext(workspaceLock,
		w.Path,
		parse.WithListOptions(&filehelpers.ListOptions{
			Flags:   filehelpers.FilesRecursive,
			Exclude: w.exclusions,
			// load files specified by inclusions
			Include: filehelpers.InclusionsFromExtensions(app_specific.ModDataExtensions),
		}))

	parseCtx.Credentials = w.Credentials
	parseCtx.Integrations = w.Integrations
	parseCtx.Notifiers = w.Notifiers

	// I don't think we need CredentialImports here .. it's fully resolved to Credentials at startup

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
		opts := &modinstaller.InstallOpts{WorkspaceMod: w.Mod}
		installData, err := modinstaller.InstallWorkspaceDependencies(ctx, opts)
		if err != nil {
			return nil, err
		}
		workspaceLock = installData.NewLock
	}
	return workspaceLock, nil
}

func (w *Workspace) loadExclusions() error {
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

func (w *Workspace) verifyResourceRuntimeDependencies() error {
	for _, d := range w.Mod.ResourceMaps.Dashboards {
		if err := d.ValidateRuntimeDependencies(w); err != nil {
			return err
		}
	}
	return nil
}
