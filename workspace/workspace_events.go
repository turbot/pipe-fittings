package workspace

import (
	"context"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"log/slog"
)

var EventCount int64 = 0

func (w *Workspace) handleFileWatcherEvent(ctx context.Context) {
	slog.Log(ctx, constants.LevelTrace, "handleFileWatcherEvent")
	prevResourceMaps, resourceMaps, errAndWarnings := w.reloadResourceMaps(ctx)

	if errAndWarnings.GetError() != nil {
		slog.Log(ctx, constants.LevelTrace, "handleFileWatcherEvent reloadResourceMaps returned error - call PublishDashboardEvent")
		// call error hook
		if w.OnFileWatcherError != nil {
			w.OnFileWatcherError(ctx, errAndWarnings.Error)
		}

		slog.Log(ctx, constants.LevelTrace, "back from PublishDashboardEvent")
		// Flag on workspace?
		return
	}
	// if resources have changed, update introspection tables
	if !prevResourceMaps.Equals(resourceMaps) {
		// TODO KAI STEAMPIPE workspacres should not know about introspection data - STEAMPIPE will need a hook here <INTROSPECTION>
		// maybe workspace could provide a file changed hook which Steampipe uses

		//// update the client with the new introspection data
		//w.onNewIntrospectionData(ctx, client)

		if w.onFileWatcherEventMessages != nil {
			w.onFileWatcherEventMessages()
		}
	}

	// call hook
	if w.OnFileWatcherEvent != nil {
		w.OnFileWatcherEvent(ctx, resourceMaps, prevResourceMaps)
	}
}

// TODO KAI STEAMPIPE workspaces should not know about introspection data - STEAMPIPE will need a hook here <INTROSPECTION>
// maybe workspace could provide a file changed hook which Steampipe uses
//func (w *Workspace) onNewIntrospectionData(ctx context.Context, client *db_client.DbClient) {
//	if viper.GetString(constants.ArgIntrospection) == constants.IntrospectionNone {
//		// nothing to do here
//		return
//	}
//	client.ResetPools(ctx)
//	res := client.AcquireSession(ctx)
//	if res.Session != nil {
//		res.Session.Close(error_helpers.IsContextCanceled(ctx))
//	}
//	if res != nil {
//		fmt.Println()
//		error_helpers.ShowErrorWithMessage(ctx, res.Error, "error when refreshing session data")
//		error_helpers.ShowWarning(strings.Join(res.Warnings, "\n"))
//	}
//}

func (w *Workspace) reloadResourceMaps(ctx context.Context) (*modconfig.ResourceMaps, *modconfig.ResourceMaps, *error_helpers.ErrorAndWarnings) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// get the pre-load resource maps
	// NOTE: do not call GetResourceMaps - we DO NOT want to lock loadLock
	prevResourceMaps := w.Mod.ResourceMaps
	// if there is an outstanding watcher error, set prevResourceMaps to empty to force refresh
	if w.watcherError != nil {
		prevResourceMaps = modconfig.NewModResources(w.Mod)
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
