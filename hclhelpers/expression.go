package hclhelpers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ExprIsNativeQuotedString determines whether the given expression looks like
// it's a quoted string in the HCL native syntax.
//
// This should be used sparingly only for situations where our legacy HCL
// decoding would've expected a keyword or reference in quotes but our new
// decoding expects the keyword or reference to be provided directly as
// an identifier-based expression.
func ExprIsNativeQuotedString(expr hcl.Expression) bool {
	_, ok := expr.(*hclsyntax.TemplateExpr)
	return ok
}
