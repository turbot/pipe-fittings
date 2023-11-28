package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

func PropertyPathFromExpression(expr hcl.Expression) (bool, *ParsedPropertyPath, error) {
	var propertyPathStr string
	var isArray bool

dep_loop:
	for {
		switch e := expr.(type) {
		case *hclsyntax.ScopeTraversalExpr:
			propertyPathStr = hclhelpers.TraversalAsString(e.Traversal)
			break dep_loop
		case *hclsyntax.SplatExpr:
			root := hclhelpers.TraversalAsString(e.Source.(*hclsyntax.ScopeTraversalExpr).Traversal)
			var suffix string
			// if there is a property path, add it
			if each, ok := e.Each.(*hclsyntax.RelativeTraversalExpr); ok {
				suffix = fmt.Sprintf(".%s", hclhelpers.TraversalAsString(each.Traversal))
			}
			propertyPathStr = fmt.Sprintf("%s.*%s", root, suffix)
			break dep_loop
		case *hclsyntax.TupleConsExpr:
			// TACTICAL
			// handle the case where an arg value is given as a runtime dependency inside an array, for example
			// arns = [input.arn]
			// this is a common pattern where a runtime depdency gives a scalar value, but an array is needed for the arg
			// NOTE: this code only supports a SINGLE item in the array
			if len(e.Exprs) != 1 {
				return false, nil, fmt.Errorf("unsupported runtime dependency expression - only a single runtime dependency item may be wrapped in an array")
			}
			isArray = true
			expr = e.Exprs[0]
			// fall through to rerun loop with updated expr
		default:
			// unhandled expression type
			return false, nil, fmt.Errorf("unexpected runtime dependency expression type")
		}
	}

	propertyPath, err := ParseResourcePropertyPath(propertyPathStr)
	if err != nil {
		return false, nil, err
	}
	return isArray, propertyPath, nil
}

type ParsedPropertyPath struct {
	Mod          string
	ItemType     string
	Name         string
	PropertyPath []string
	// optional scope of this property path ("self")
	Scope    string
	Original string
}

func (p *ParsedPropertyPath) PropertyPathString() string {
	return strings.Join(p.PropertyPath, ".")
}

func (p *ParsedPropertyPath) ToParsedResourceName() *ParsedResourceName {
	return &ParsedResourceName{
		Mod:      p.Mod,
		ItemType: p.ItemType,
		Name:     p.Name,
	}
}

func (p *ParsedPropertyPath) ToResourceName() string {
	return BuildModResourceName(p.ItemType, p.Name)
}

func (p *ParsedPropertyPath) String() string {
	return p.Original
}

func ParseResourcePropertyPath(propertyPath string) (*ParsedPropertyPath, error) {
	res := &ParsedPropertyPath{Original: propertyPath}

	// valid property paths:
	// <mod>.<resource>.<name>.<property path...>
	// <resource>.<name>.<property path...>
	// so either the first or second slice must be a valid resource type

	//
	// unless they are some flowpipe resources:
	//
	// mod.trigger.trigger_type.trigger_name.<property_path>
	// trigger.trigger_type.trigger_name.<property_path>
	//
	// We can have trigger and integration in this current format

	parts := strings.Split(propertyPath, ".")
	if len(parts) < 2 {
		return nil, perr.BadRequestWithMessage("invalid property path: " + propertyPath)
	}

	// special case handling for runtime dependencies which may have use the "self" qualifier
	// const RuntimeDependencyDashboardScope = "self"
	if parts[0] == "self" {
		res.Scope = parts[0]
		parts = parts[1:]
	}

	// special case if the first part is "each" or "result", both are Flowpipe magic keywords that mean self reference,
	// they are not dependency to other resources
	if parts[0] == "each" || parts[0] == "result" {
		return nil, nil
	}

	if schema.IsValidResourceItemType(parts[0]) {
		// put empty mod as first part
		parts = append([]string{""}, parts...)
	}

	if len(parts) < 3 {
		return nil, perr.BadRequestWithMessage("invalid property path: " + propertyPath)
	}

	switch len(parts) {
	case 3:
		// no property path specified
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
	default:
		if parts[1] == "integration" || parts[1] == "trigger" || parts[1] == "credential" {
			res.Mod = parts[0]
			res.ItemType = parts[1]
			res.Name = parts[2] + "." + parts[3]
			if len(parts) > 4 {
				res.PropertyPath = parts[3:]
			}
		} else {
			res.Mod = parts[0]
			res.ItemType = parts[1]
			res.Name = parts[2]
			res.PropertyPath = parts[3:]
		}
	}

	if !schema.IsValidResourceItemType(res.ItemType) {
		return nil, perr.BadRequestWithMessage("invalid resource item type passed to ParseResourcePropertyPath: " + propertyPath)
	}

	return res, nil
}
