package workspace

import (
	"context"
	"log/slog"

	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

var EventCount int64 = 0

func (w *WorkspaceBase[T]) handleFileWatcherEvent(ctx context.Context) {
	slog.Debug("handleFileWatcherEvent")
	prevResourceMaps, resourceMaps, errAndWarnings := w.ReloadResourceMaps(ctx)

	if errAndWarnings.GetError() != nil {
		slog.Debug("handleFileWatcherEvent reloadResourceMaps returned error - call PublishDashboardEvent")
		// call error hook
		if w.OnFileWatcherError != nil {
			w.OnFileWatcherError(ctx, errAndWarnings.Error)
		}

		slog.Debug("back from PublishDashboardEvent")
		// Flag on workspace?
		return
	}
	// if resources have changed, update introspection tables
	if !prevResourceMaps.Equals(resourceMaps) {
		if w.onFileWatcherEventMessages != nil {
			w.onFileWatcherEventMessages()
		}
	}

	// call hook
	if w.OnFileWatcherEvent != nil {
		w.OnFileWatcherEvent(ctx, resourceMaps, prevResourceMaps)
	}
}

func (w *WorkspaceBase[T]) ReloadResourceMaps(ctx context.Context) (modconfig.ResourceMapsI, modconfig.ResourceMapsI, error_helpers.ErrorAndWarnings) {
	w.LoadLock()
	defer w.LoadUnlock()

	// get the pre-load resource maps
	// NOTE: do not call GetResourceMaps - we DO NOT want to lock LoadLock
	prevResourceMaps := w.Mod.GetResourceMaps()
	// if there is an outstanding watcher error, set prevResourceMaps to empty to force refresh
	if w.WatcherError != nil {
		prevResourceMaps = modconfig.NewResourceMaps(w.Mod)
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
	resourceMaps := w.Mod.GetResourceMaps()

	return prevResourceMaps, resourceMaps, errAndWarnings

}
