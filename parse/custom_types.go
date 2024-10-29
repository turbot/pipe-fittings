package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

// customTypeFunc is a function that returns a custom cty.Type for a given subtype
type customTypeFunc func(string) cty.Type

// customTypeMappings is a map of resource types to custom cty.Type functions
func getCustomTypeFunc(blockType string) (customTypeFunc, bool) {
	if blockType == schema.BlockTypeConnection {
		return app_specific_connection.ConnectionCtyType, true
	}

	// otherwise try app specific custom types
	ty, ok := app_specific.CustomTypes[blockType]
	if ok {
		f := func(string) cty.Type {
			return ty
		}
		return f, true
	}
	return nil, false
}

// customTypeFromExpr returns the custom cty.Type for the given hcl.Expression, if one is registered
func customTypeFromExpr(expr hcl.Expression) (cty.Type, bool) {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		return customTypeFromScopeTraversalExpr(e)
	case *hclsyntax.FunctionCallExpr:
		return customTypeFromFunctionCallExpr(e)
	default:
		return cty.NilType, true

	}
}

// customTypeFromScopeTraversalExpr returns the custom cty.Type for the given hclsyntax.ScopeTraversalExpr,
// if one is registered
func customTypeFromScopeTraversalExpr(expr *hclsyntax.ScopeTraversalExpr) (cty.Type, bool) {
	dottedString := hclhelpers.TraversalAsString(expr.Traversal)
	parts := strings.Split(dottedString, ".")
	// extract the resource type and (optionally) the subtype
	blockType := parts[0]
	var subtype string
	if len(parts) == 2 {
		subtype = parts[1]
	}

	// do we have a custom type mapping for this type?
	customTypeFunc, ok := getCustomTypeFunc(blockType)
	if !ok {
		return cty.NilType, true
	}

	customType := customTypeFunc(subtype)
	return customType, customType == cty.NilType
}

// customTypeFromFunctionCallExpr returns the custom cty.Type for the given hclsyntax.FunctionCallExpr,
// if one is registered
func customTypeFromFunctionCallExpr(fCallExpr *hclsyntax.FunctionCallExpr) (cty.Type, bool) {
	// curently only handling list function with single args
	if fCallExpr.Name != "list" || len(fCallExpr.Args) != 1 {
		return cty.NilType, true
	}

	dottedString := hclhelpers.TraversalAsString(fCallExpr.Args[0].Variables()[0])
	parts := strings.Split(dottedString, ".")

	// extract the resource type and (optionally) the subtype
	blockType := parts[0]
	var subtype string
	if len(parts) == 2 {
		subtype = parts[1]
	}

	// do we have a custom type mapping for this type?
	customTypeFunc, ok := getCustomTypeFunc(blockType)
	if !ok {
		return cty.NilType, true
	}

	// return a list of the custom  type
	customType := customTypeFunc(subtype)
	if customType == cty.NilType {
		return cty.NilType, true
	}
	return cty.List(customType), false
}
