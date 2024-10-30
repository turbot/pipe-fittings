package workspace

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

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
	_ = w.GetModResources().WalkResources(resourceFunc)

	return res
}

// FilterWorkspaceResourcesOfType returns all resources of type T from a workspace which satisf          y the filter,
// which is specified as a SQL syntax where clause
func FilterWorkspaceResourcesOfType[T modconfig.HclResource](w *Workspace, filter ResourceFilter) (map[string]T, error) {
	var res = map[string]T{}

	filterPredicate, err := filter.getPredicate()
	if err != nil {
		return nil, err
	}

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		// if item is correct type and matches the predicate, add it to the result
		if item, ok := item.(T); ok && filterPredicate(item) {
			res[item.Name()] = item
		}
		return true, nil
	}

	// resource func does not return error
	_ = w.GetModResources().WalkResources(resourceFunc)

	return res, nil
}
