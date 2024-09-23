package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"reflect"
	"strings"
)

type customTypeFunc func(string) cty.Type

var customTypeMappings = map[string]customTypeFunc{
	schema.BlockTypeConnection: app_specific_connection.ConnectionCtyType,
	schema.BlockTypeNotifier: func(string) cty.Type {
		return cty.Capsule("NotifierCtyType", reflect.TypeOf(&modconfig.NotifierImpl{}))
	},
}

//}
//
//var ConnectionCtyType = cty.Capsule("ConnectionCtyType", reflect.TypeOf(&connection.ConnectionImpl{}))
//var NotifierCtyType =

func customTypeFromExpr(expr hcl.Expression) (cty.Type, bool) {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		return customTypeFromScopeTraversalExpr(e)
	case *hclsyntax.FunctionCallExpr:
		return customTypeFromFunctionCallExpr(e)
	default:
		// TODO KAI verify - this is new error behaviour - add test?
		return cty.NilType, true

	}
}

func customTypeFromScopeTraversalExpr(expr *hclsyntax.ScopeTraversalExpr) (cty.Type, bool) {
	dottedString := hclhelpers.TraversalAsString(expr.Traversal)
	parts := strings.Split(dottedString, ".")
	// extract the resource type and (optionally) the subtype
	ty := parts[0]
	var subtype string
	if len(parts) == 2 {
		subtype = parts[1]
	}

	// do we have a custom type mapping for this type?
	customTypeFunc, ok := customTypeMappings[ty]
	if !ok {
		return cty.NilType, true
	}

	return customTypeFunc(subtype), false
}

func customTypeFromFunctionCallExpr(fCallExpr *hclsyntax.FunctionCallExpr) (cty.Type, bool) {
	// curently only handling list function with single args
	if fCallExpr.Name != "list" || len(fCallExpr.Args) != 1 {
		return cty.NilType, true
	}

	dottedString := hclhelpers.TraversalAsString(fCallExpr.Args[0].Variables()[0])
	parts := strings.Split(dottedString, ".")

	// extract the resource type and (optionally) the subtype
	ty := parts[0]
	var subtype string
	if len(parts) == 2 {
		subtype = parts[1]
	}

	// do we have a custom type mapping for this type?
	customTypeFunc, ok := customTypeMappings[ty]
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
