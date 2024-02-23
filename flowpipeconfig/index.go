package flowpipeconfig

import (
	"context"
	"log/slog"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/modconfig"
)

type FlowpipeConfig struct {
	ConfigPaths []string

	CredentialImports map[string]credential.CredentialImport
	Credentials       map[string]credential.Credential
	Integrations      map[string]modconfig.Integration
	Notifiers         map[string]modconfig.Notifier

	watcher                 *filewatcher.FileWatcher
	fileWatcherErrorHandler func(context.Context, error)
	// watcherError               error
	onFileWatcherEventMessages func()

	// Hooks
	OnFileWatcherError func(context.Context, error)
	OnFileWatcherEvent func(context.Context)
}

func (f *FlowpipeConfig) Equals(other *FlowpipeConfig) bool {
	if len(f.Credentials) != len(other.Credentials) {
		return false
	}

	for k := range f.Credentials {
		if _, ok := other.Credentials[k]; !ok {
			return false
		}

		// if !other.Credentials[k].Equals(v) {
		// 	return false
		// }
	}

	if len(f.Integrations) != len(other.Integrations) {
		return false
	}

	// for k, v := range f.Integrations {
	// check if k exists in other
	// 	if !other.Integrations[k].Equals(v) {
	// 		return false
	// 	}
	// }

	if len(f.Notifiers) != len(other.Notifiers) {
		return false
	}

	// for k, v := range f.Notifiers {
	// check if k exists in other
	// 	if !other.Notifiers[k].Equals(v) {
	// 		return false
	// 	}
	// }

	if len(f.CredentialImports) != len(other.CredentialImports) {
		return false
	}

	for k, v := range f.CredentialImports {

		if _, ok := other.CredentialImports[k]; !ok {
			return false
		}

		if !other.CredentialImports[k].Equals(v) {
			return false
		}
	}

	return true
}

func (f *FlowpipeConfig) SetupWatcher(ctx context.Context, errorHandler func(context.Context, error)) error {
	watcherOptions := &filewatcher.WatcherOptions{
		Directories: f.ConfigPaths,
		Include:     filehelpers.InclusionsFromExtensions([]string{".fpc"}),
		ListFlag:    filehelpers.FilesRecursive,
		EventMask:   fsnotify.Create | fsnotify.Remove | fsnotify.Rename | fsnotify.Write,
		// we should look into passing the callback function into the underlying watcher
		// we need to analyze the kind of errors that come out from the watcher and
		// decide how to handle them
		// OnError: errCallback,
		OnChange: func(events []fsnotify.Event) {
			f.handleFileWatcherEvent(ctx)
		},
	}
	watcher, err := filewatcher.NewWatcher(watcherOptions)
	if err != nil {
		return err
	}
	f.watcher = watcher

	// start the watcher
	watcher.Start()

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	f.fileWatcherErrorHandler = errorHandler

	return nil
}

func (f *FlowpipeConfig) handleFileWatcherEvent(ctx context.Context) {
	slog.Debug("FlowpipeConfig handleFileWatcherEvent")

	newFpConfig, errAndWarnings := LoadFlowpipeConfig(f.ConfigPaths)

	if errAndWarnings.GetError() != nil {
		// call error hook
		if f.OnFileWatcherError != nil {
			f.OnFileWatcherError(ctx, errAndWarnings.Error)
		}

		// Flag on workspace?
		return
	}

	if !newFpConfig.Equals(f) {
		if f.onFileWatcherEventMessages != nil {
			f.onFileWatcherEventMessages()
		}
	}

	// call hook
	if f.OnFileWatcherEvent != nil {
		f.OnFileWatcherEvent(ctx)
	}
}

func NewFlowpipeConfig(configPaths []string) *FlowpipeConfig {
	defaultCreds, err := credential.DefaultCredentials()
	if err != nil {
		slog.Error("Unable to create default credentials", "error", err)
		return nil
	}

	defaultIntegrations, err := modconfig.DefaultIntegrations()
	if err != nil {
		slog.Error("Unable to create default integrations", "error", err)
		return nil
	}

	defaultNotifiers, err := modconfig.DefaultNotifiers(defaultIntegrations["webform.default"])
	if err != nil {
		slog.Error("Unable to create default notifiers", "error", err)
		return nil
	}

	fpConfig := FlowpipeConfig{
		CredentialImports: make(map[string]credential.CredentialImport),
		Credentials:       defaultCreds,
		Integrations:      defaultIntegrations,
		Notifiers:         defaultNotifiers,
		ConfigPaths:       configPaths,
	}

	return &fpConfig
}
