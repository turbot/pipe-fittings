package workspace

import (
	"github.com/turbot/pipe-fittings/modconfig/powerpipe"
	"log/slog"

	"github.com/turbot/pipe-fittings/modconfig"
)

func (w *Workspace[T]) GetQueryProvider(queryName string) (powerpipe.QueryProvider, bool) {
	parsedName, err := modconfig.ParseResourceName(queryName)
	if err != nil {
		return nil, false
	}
	// try to find the resource
	if resource, ok := w.GetResource(parsedName); ok {
		// found a resource - is it a query provider
		if qp := resource.(powerpipe.QueryProvider); ok {
			return qp, true
		}
		slog.Debug("GetQueryProviderImpl found a mod resource resource for query but it is not a query provider", "resourceName", queryName)
	}

	return nil, false
}

// GetResourceMaps implements ResourceMapsProvider
func (w *Workspace[T]) GetResourceMaps() modconfig.ResourceMapsI {

	w.LoadLock()
	defer w.LoadUnlock()

	// if this a source snapshot workspace, create a ModResources containing ONLY source snapshot paths
	if len(w.SourceSnapshots) != 0 {
		return powerpipe.NewSourceSnapshotModResources(w.SourceSnapshots)
	}
	return w.Mod.GetResourceMaps()
}

func (w *Workspace[T]) GetResource(parsedName *modconfig.ParsedResourceName) (resource modconfig.HclResource, found bool) {
	return w.GetResourceMaps().GetResource(parsedName)
}
