package modconfig

import (
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-helpers/hclhelpers"
)

type ResourceDependency struct {
	Range      hcl.Range
	Traversals []hcl.Traversal
}

func (d *ResourceDependency) String() string {
	traversalStrings := make([]string, len(d.Traversals))
	for i, t := range d.Traversals {
		traversalStrings[i] = hclhelpers.TraversalAsString(t)
	}
	return strings.Join(traversalStrings, ",")
}

func (d *ResourceDependency) IsRuntimeDependency() bool {
	// runtime dependency wil only have a single traversal
	if len(d.Traversals) > 1 {
		return false
	}
	// parse the traversal as a property path
	propertyPath, err := ParseResourcePropertyPath(hclhelpers.TraversalAsString(d.Traversals[0]))
	if err != nil {
		return false
	}

	return isRunTimeDependencyProperty(propertyPath)

}

func isRunTimeDependencyProperty(propertyPath *ParsedPropertyPath) bool {
	// supported runtime dependencies
	// map is keyed by resource type and contains a list of properties
	runTimeDependencyPropertyPaths := map[string][]string{
		"input": {"value"},
		"param": {"value"},
	}
	// is this property a supported runtime dependency property
	if supportedProperties, ok := runTimeDependencyPropertyPaths[propertyPath.ItemType]; ok {
		return slices.Contains(supportedProperties, propertyPath.PropertyPathString())
	}
	return false
}
