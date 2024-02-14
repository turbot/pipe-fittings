package workspace

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"log/slog"
)

func (w *Workspace) GetQueryProvider(queryName string) (modconfig.QueryProvider, bool) {
	parsedName, err := modconfig.ParseResourceName(queryName)
	if err != nil {
		return nil, false
	}
	// try to find the resource
	if resource, ok := w.GetResource(parsedName); ok {
		// found a resource - is it a query provider
		if qp := resource.(modconfig.QueryProvider); ok {
			return qp, true
		}
		slog.Debug("GetQueryProviderImpl found a mod resource resource for query but it is not a query provider", "resourceName", queryName)
	}

	return nil, false
}

// GetResourceMaps implements ResourceMapsProvider
func (w *Workspace) GetResourceMaps() *modconfig.ResourceMaps {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if this a source snapshot workspace, create a ResourceMaps containing ONLY source snapshot paths
	if len(w.SourceSnapshots) != 0 {
		return modconfig.NewSourceSnapshotModResources(w.SourceSnapshots)
	}
	return w.Mod.ResourceMaps
}

func (w *Workspace) GetResource(parsedName *modconfig.ParsedResourceName) (resource modconfig.HclResource, found bool) {
	return w.GetResourceMaps().GetResource(parsedName)
}
