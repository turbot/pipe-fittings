package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"
)

func DecodeTypeExpression(attr *hcl.Attribute) (cty.Type, hcl.Diagnostics) {
	expr := attr.Expr

	ty, diags := typeexpr.TypeConstraint(expr)
	if !diags.HasErrors() {
		return ty, nil
	}

	// so we failed to parse the type constraint - special case handling required

	var typeErr bool
	switch hcl.ExprAsKeyword(expr) {
	// Handle shorthand forms for list, map, and set
	case "list":
		ty = cty.List(cty.DynamicPseudoType)
		typeErr = false
	case "map":
		ty = cty.Map(cty.DynamicPseudoType)
		typeErr = false
	case "set":
		ty = cty.Set(cty.DynamicPseudoType)
		typeErr = false
	default:
		// Try to parse the expression as a custom type
		ty, typeErr = customTypeFromExpr(expr)
	}

	// did we manage to determine the type?
	if typeErr {
		// create new diagnostics
		diags = hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "A type specification is either a primitive type keyword (bool, number, string), complex type constructor call or Turbot custom type (connection, notifier)",
			Subject:  &attr.Range,
		}}
		return cty.Type{}, diags
	}

	return ty, nil
}
