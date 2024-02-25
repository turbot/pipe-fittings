package hclhelpers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func ExpressionsEqual(expr1, expr2 hcl.Expression) bool {
	if expr1 == nil && expr2 == nil {
		return true
	}

	if expr1 == nil || expr2 == nil {
		return false
	}

	if len(expr1.Variables()) != len(expr2.Variables()) {
		return false
	}

	for i, v := range expr1.Variables() {
		v2 := expr2.Variables()[i]

		if v.RootName() != v2.RootName() {
			return false
		}
	}

	if expr1TemplateExpr, ok := expr1.(*hclsyntax.TemplateExpr); ok {

		if expr2TemplateExpr, ok := expr2.(*hclsyntax.TemplateExpr); !ok {
			return false
		} else {
			for i, part := range expr1TemplateExpr.Parts {
				if !ExpressionsEqual(part, expr2TemplateExpr.Parts[i]) {
					return false
				}
			}
		}
	}

	if expr1LiteralValue, ok := expr1.(*hclsyntax.LiteralValueExpr); ok {
		if expr2LiteralValue, ok := expr2.(*hclsyntax.LiteralValueExpr); !ok {
			return false
		} else {
			return expr1LiteralValue.Val.Equals(expr2LiteralValue.Val) == cty.True
		}
	}

	return true
}
