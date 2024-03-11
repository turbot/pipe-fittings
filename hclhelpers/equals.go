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
	} else if expr1LiteralValue, ok := expr1.(*hclsyntax.LiteralValueExpr); ok {
		if expr2LiteralValue, ok := expr2.(*hclsyntax.LiteralValueExpr); !ok {
			return false
		} else {
			return expr1LiteralValue.Val.Equals(expr2LiteralValue.Val) == cty.True
		}
	} else if expr1BinaryOp, ok := expr1.(*hclsyntax.BinaryOpExpr); ok {
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
	} else if expr1ScopeTraversal, ok := expr1.(*hclsyntax.ScopeTraversalExpr); ok {
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
	} else if expr1TupleConsExpr, ok := expr1.(*hclsyntax.TupleConsExpr); ok {
		if expr2TupleConsExpr, ok := expr2.(*hclsyntax.TupleConsExpr); !ok {
			return false
		} else {
			if len(expr1TupleConsExpr.Exprs) != len(expr2TupleConsExpr.Exprs) {
				return false
			}

			for i, e1 := range expr1TupleConsExpr.Exprs {
				if !ExpressionsEqual(e1, expr2TupleConsExpr.Exprs[i]) {
					return false
				}
			}
		}
	} else if expr1FuncCallExpr, ok := expr1.(*hclsyntax.FunctionCallExpr); ok {
		if expr2FuncCallExpr, ok := expr2.(*hclsyntax.FunctionCallExpr); !ok {
			return false
		} else {
			if expr1FuncCallExpr.Name != expr2FuncCallExpr.Name {
				return false
			}

			if len(expr1FuncCallExpr.Args) != len(expr2FuncCallExpr.Args) {
				return false
			}

			for i, a1 := range expr1FuncCallExpr.Args {
				if !ExpressionsEqual(a1, expr2FuncCallExpr.Args[i]) {
					return false
				}
			}
		}
	} else if expr1ObjConsExpr, ok := expr1.(*hclsyntax.ObjectConsExpr); ok {
		if expr2ObjConsExpr, ok := expr2.(*hclsyntax.ObjectConsExpr); !ok {
			return false
		} else {
			if len(expr1ObjConsExpr.Items) != len(expr2ObjConsExpr.Items) {
				return false
			}

			for i, item1 := range expr1ObjConsExpr.Items {
				if !ExpressionsEqual(item1.KeyExpr, expr2ObjConsExpr.Items[i].KeyExpr) {
					return false
				}

				if !ExpressionsEqual(item1.ValueExpr, expr2ObjConsExpr.Items[i].ValueExpr) {
					return false
				}
			}
		}
	} else if expr1ConditionalExpr, ok := expr1.(*hclsyntax.ConditionalExpr); ok {
		if expr2ConditionalExpr, ok := expr2.(*hclsyntax.ConditionalExpr); !ok {
			return false
		} else {
			if !ExpressionsEqual(expr1ConditionalExpr.Condition, expr2ConditionalExpr.Condition) {
				return false
			}

			if !ExpressionsEqual(expr1ConditionalExpr.TrueResult, expr2ConditionalExpr.TrueResult) {
				return false
			}

			if !ExpressionsEqual(expr1ConditionalExpr.FalseResult, expr2ConditionalExpr.FalseResult) {
				return false
			}
		}
	}

	return true
}
