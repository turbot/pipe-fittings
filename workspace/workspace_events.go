package workspace

import (
	"context"
	"log/slog"

	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

var EventCount int64 = 0

func (w *Workspace) handleFileWatcherEvent(ctx context.Context) {
	slog.Debug("handleFileWatcherEvent")
	prevModResources, modResources, errAndWarnings := w.ReloadModResources(ctx)

	if errAndWarnings.GetError() != nil {
		slog.Debug("handleFileWatcherEvent reloadModResources returned error - call PublishDashboardEvent")
		// call error hook
		if w.OnFileWatcherError != nil {
			w.OnFileWatcherError(ctx, errAndWarnings.Error)
		}

		slog.Debug("back from PublishDashboardEvent")
		// Flag on workspace?
		return
	}
	// if resources have changed, update introspection tables
	if !prevModResources.Equals(modResources) {
		if w.onFileWatcherEventMessages != nil {
			w.onFileWatcherEventMessages()
		}
	}

	// call hook
	if w.OnFileWatcherEvent != nil {
		w.OnFileWatcherEvent(ctx, modResources, prevModResources)
	}
}

func (w *Workspace) ReloadModResources(ctx context.Context) (modconfig.ModResources, modconfig.ModResources, error_helpers.ErrorAndWarnings) {
	w.LoadLock()
	defer w.LoadUnlock()

	// get the pre-load resource maps
	// NOTE: do not call GetModResources - we DO NOT want to lock LoadLock
	prevModResources := w.Mod.GetModResources()
	// if there is an outstanding watcher error, set prevModResources to empty to force refresh
	if w.WatcherError != nil {
		prevModResources = modconfig.NewModResources(w.Mod)
	}

	// now reload the workspace
	errAndWarnings := w.LoadWorkspaceMod(ctx)
	if errAndWarnings.GetError() != nil {
		// check the existing watcher error - if we are already in an error state, do not show error
		if w.WatcherError == nil {
			w.FileWatcherErrorHandler(ctx, error_helpers.PrefixError(errAndWarnings.GetError(), "failed to reload workspace"))
		}
		// now set watcher error to new error
		w.WatcherError = errAndWarnings.GetError()
		return nil, nil, errAndWarnings
	}
	// clear watcher error
	w.WatcherError = nil

	// reload the resource maps
	modResources := w.Mod.GetModResources()

	return prevModResources, modResources, errAndWarnings

}
