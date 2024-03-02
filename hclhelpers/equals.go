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
			if len(expr1TemplateExpr.Parts) != len(expr2TemplateExpr.Parts) {
				return false
			}

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

	if expr1BinaryOp, ok := expr1.(*hclsyntax.BinaryOpExpr); ok {
		if expr2BinaryOp, ok := expr2.(*hclsyntax.BinaryOpExpr); !ok {
			return false
		} else {
			if expr1BinaryOp.Op != expr2BinaryOp.Op {
				return false
			}

			if !ExpressionsEqual(expr1BinaryOp.LHS, expr2BinaryOp.LHS) {
				return false
			}

			if !ExpressionsEqual(expr1BinaryOp.RHS, expr2BinaryOp.RHS) {
				return false
			}
		}
	}

	if expr1ScopeTraversal, ok := expr1.(*hclsyntax.ScopeTraversalExpr); ok {
		if expr2ScopeTraversal, ok := expr2.(*hclsyntax.ScopeTraversalExpr); !ok {
			return false
		} else {

			if len(expr1ScopeTraversal.Traversal) != len(expr2ScopeTraversal.Traversal) {
				return false
			}

			if len(expr1ScopeTraversal.Traversal) != len(expr2ScopeTraversal.Traversal) {
				return false
			}

			for i, t1 := range expr1ScopeTraversal.Traversal {
				if t1Root, ok := t1.(hcl.TraverseRoot); ok {
					t2 := expr2ScopeTraversal.Traversal[i]
					if t2Root, ok := t2.(hcl.TraverseRoot); ok {
						if t1Root.Name != t2Root.Name {
							return false
						}
					}
				} else if t1Attr, ok := t1.(hcl.TraverseAttr); ok {
					t2 := expr2ScopeTraversal.Traversal[i]
					if t2Attr, ok := t2.(hcl.TraverseAttr); ok {
						if t1Attr.Name != t2Attr.Name {
							return false
						}
					}
				}
			}
		}
	}

	return true
}
