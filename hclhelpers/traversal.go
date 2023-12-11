package hclhelpers

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// TraversalAsString converts a traversal to a path string
// (if an absolute traversal is passed - convert to relative)
func TraversalAsString(traversal hcl.Traversal) string {
	var parts = make([]string, len(traversal))
	offset := 0

	if !traversal.IsRelative() {
		s := traversal.SimpleSplit()
		parts[0] = s.Abs.RootName()
		offset++
		traversal = s.Rel
	}
	for i, r := range traversal {
		switch t := r.(type) {
		case hcl.TraverseAttr:
			parts[i+offset] = t.Name
		case hcl.TraverseIndex:
			idx, err := CtyToString(t.Key)
			if err != nil {
				// we do not expect this to fail
				continue
			}
			parts[i+offset] = idx
		}
	}
	return strings.Join(parts, ".")
}

// TraversalAsStringSlice converts a traversal to a path string
// (if an absolute traversal is passed - convert to relative)
func TraversalAsStringSlice(traversal hcl.Traversal) []string {
	var parts = make([]string, len(traversal))
	offset := 0

	if !traversal.IsRelative() {
		s := traversal.SimpleSplit()
		parts[0] = s.Abs.RootName()
		offset++
		traversal = s.Rel
	}
	for i, r := range traversal {
		switch t := r.(type) {
		case hcl.TraverseAttr:
			parts[i+offset] = t.Name
		case hcl.TraverseIndex:
			idx, err := CtyToString(t.Key)
			if err != nil {
				// we do not expect this to fail
				continue
			}
			parts[i+offset] = idx
		}
	}
	return parts
}

// operationToString maps an operation to its string representation
func operationToString(op *hclsyntax.Operation) string {
	switch op {
	case hclsyntax.OpLogicalOr:
		return "||"
	case hclsyntax.OpLogicalAnd:
		return "&&"
	case hclsyntax.OpLogicalNot:
		return "!"
	case hclsyntax.OpEqual:
		return "=="
	case hclsyntax.OpNotEqual:
		return "!="
	case hclsyntax.OpGreaterThan:
		return ">"
	case hclsyntax.OpGreaterThanOrEqual:
		return ">="
	case hclsyntax.OpLessThan:
		return "<"
	case hclsyntax.OpLessThanOrEqual:
		return "<="
	case hclsyntax.OpAdd:
		return "+"
	case hclsyntax.OpSubtract:
		return "-"
	case hclsyntax.OpMultiply:
		return "*"
	case hclsyntax.OpDivide:
		return "/"
	case hclsyntax.OpModulo:
		return "%"
	case hclsyntax.OpNegate:
		return "-"
	default:
		return "unknown operation"
	}
}

func ExpressionAsLiteralSlice(expr hcl.Expression) ([]string, error) {
	allParts := make([]string, 0)

	switch v := expr.(type) {
	case *hclsyntax.TemplateExpr:
		for _, p := range v.Parts {
			parts, err := ExpressionAsLiteralSlice(p)
			if err != nil {
				return nil, err
			}

			_, ok := p.(*hclsyntax.LiteralValueExpr)
			if !ok {
				allParts = append(allParts, "${"+strings.Join(parts, ".")+"}")
			} else {
				allParts = append(allParts, parts...)
			}
		}

	case *hclsyntax.BinaryOpExpr:
		parts, err := ExpressionAsLiteralSlice(v.LHS)
		if err != nil {
			return nil, err
		}

		allParts = append(allParts, parts...)

		allParts = append(allParts, operationToString(v.Op))

		parts, err = ExpressionAsLiteralSlice(v.RHS)
		if err != nil {
			return nil, err
		}

		allParts = append(allParts, parts...)

	case *hclsyntax.LiteralValueExpr:
		goVal, err := CtyToGo(v.Val)
		if err != nil {
			return nil, err
		}
		allParts = append(allParts, fmt.Sprintf("%v", goVal))

	default:
		for _, tss := range expr.Variables() {
			parts := TraversalAsStringSlice(tss)
			partString := strings.Join(parts, ".")
			allParts = append(allParts, partString)
		}
	}

	return allParts, nil
}
func AttributeAsLiteral(attr *hcl.Attribute) string {
	// This is one attempt .. but difficult to implement because we merged all our hcl files into one single body,
	// tracking back which one is which is difficult to track
	// rng := attr.Expr.Range()
	// source := file.Bytes[rng.Start.Byte:rng.End.Byte]
	// fmt.Printf("Original expression: %s\n", string(source))
	// manual reconstruction of the expression

	if attr == nil {
		return ""
	}

	allParts, err := ExpressionAsLiteralSlice(attr.Expr)
	if err != nil {
		return "Unable to evaluate expression " + err.Error()
	}

	return strings.Join(allParts, "")
}

func TraversalsEqual(t1, t2 hcl.Traversal) bool {
	return TraversalAsString(t1) == TraversalAsString(t2)
}

// ResourceNameFromTraversal converts a traversal to the name of the referenced resource
// We must take into account possible mod-name as first traversal element
func ResourceNameFromTraversal(resourceType string, traversal hcl.Traversal) (string, bool) {
	traversalString := TraversalAsString(traversal)
	split := strings.Split(traversalString, ".")

	// the resource reference will be of the form
	// var.<var_name>
	// or
	// <resource_type>.<resource_name>.<property>
	// or
	// <mod_name>.<resource_type>.<resource_name>.<property>

	if split[0] == "var" {
		return strings.Join(split, "."), true
	}
	if len(split) >= 2 && split[0] == resourceType {
		return strings.Join(split[:2], "."), true
	}
	if len(split) >= 3 && split[1] == resourceType {
		return strings.Join(split[:3], "."), true
	}
	return "", false
}
