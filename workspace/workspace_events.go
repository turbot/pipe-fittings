package workspace

import (
	"context"
	"log/slog"

	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

var EventCount int64 = 0

func (w *Workspace) handleFileWatcherEvent(ctx context.Context) {
	slog.Debug("handleFileWatcherEvent")
	prevResourceMaps, resourceMaps, errAndWarnings := w.reloadResourceMaps(ctx)

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

func (w *Workspace) ReloadResourceMaps(ctx context.Context) (*modconfig.ResourceMaps, *modconfig.ResourceMaps, error_helpers.ErrorAndWarnings) {
	return w.reloadResourceMaps(ctx)
}

func (w *Workspace) reloadResourceMaps(ctx context.Context) (*modconfig.ResourceMaps, *modconfig.ResourceMaps, error_helpers.ErrorAndWarnings) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// get the pre-load resource maps
	// NOTE: do not call GetResourceMaps - we DO NOT want to lock loadLock
	prevResourceMaps := w.Mod.ResourceMaps
	// if there is an outstanding watcher error, set prevResourceMaps to empty to force refresh
	if w.watcherError != nil {
		prevResourceMaps = modconfig.NewResourceMaps(w.Mod)
	}

	// now reload the workspace
	errAndWarnings := w.loadWorkspaceMod(ctx)
	if errAndWarnings.GetError() != nil {
		// check the existing watcher error - if we are already in an error state, do not show error
		if w.watcherError == nil {
			w.fileWatcherErrorHandler(ctx, error_helpers.PrefixError(errAndWarnings.GetError(), "failed to reload workspace"))
		}
		// now set watcher error to new error
		w.watcherError = errAndWarnings.GetError()
		return nil, nil, errAndWarnings
	}
	// clear watcher error
	w.watcherError = nil

	// reload the resource maps
	resourceMaps := w.Mod.ResourceMaps

	return prevResourceMaps, resourceMaps, errAndWarnings

}
