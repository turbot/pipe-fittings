package workspace

import (
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/modconfig"
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

// ResolveQueryFromQueryProvider resolves the query for the given QueryProvider
func (w *Workspace) ResolveQueryFromQueryProvider(queryProvider modconfig.QueryProvider, runtimeArgs *modconfig.QueryArgs) (*modconfig.ResolvedQuery, error) {
	slog.Debug("ResolveQueryFromQueryProvider", "resourceName", queryProvider.Name())

	query := queryProvider.GetQuery()
	sql := queryProvider.GetSQL()

	params := queryProvider.GetParams()

	// merge the base args with the runtime args
	var err error
	runtimeArgs, err = modconfig.MergeArgs(queryProvider, runtimeArgs)
	if err != nil {
		return nil, err
	}

	// determine the source for the query
	// - this will either be the control itself or any named query the control refers to
	// either via its SQL proper ty (passing a query name) or Query property (using a reference to a query object)

	// if a query is provided, use that to resolve the sql
	if query != nil {
		return w.ResolveQueryFromQueryProvider(query, runtimeArgs)
	}

	// must have sql is there is no query
	if sql == nil {
		return nil, fmt.Errorf("%s does not define  either a 'sql' property or a 'query' property\n", queryProvider.Name())
	}

	queryProviderSQL := typehelpers.SafeString(sql)
	slog.Debug("control defines inline SQL")

	// if the SQL refers to a named query, this is the same as if the 'Query' property is set
	if namedQueryProvider, ok := w.GetQueryProvider(queryProviderSQL); ok {
		// in this case, it is NOT valid for the query provider to define its own Param definitions
		if params != nil {
			return nil, fmt.Errorf("%s has an 'SQL' property which refers to %s, so it cannot define 'param' blocks", queryProvider.Name(), namedQueryProvider.Name())
		}
		return w.ResolveQueryFromQueryProvider(namedQueryProvider, runtimeArgs)
	}

	// so the  sql is NOT a named query
	return queryProvider.GetResolvedQuery(runtimeArgs)

}
