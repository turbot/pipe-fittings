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

// GetWorkspaceResourcesOfType returns all resources of type T from a workspace
func GetWorkspaceResourcesOfType[T modconfig.HclResource](w *Workspace) map[string]T {
	var res = map[string]T{}

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		if item, ok := item.(T); ok {
			res[item.Name()] = item
		}
		return true, nil
	}

	// resource func does not return error
	_ = w.GetResourceMaps().WalkResources(resourceFunc)

	return res
}
