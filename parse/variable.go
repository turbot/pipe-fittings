package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

func DecodeVariableBlock(block *hcl.Block, content *hcl.BodyContent) (*modconfig.RawVariable, hcl.Diagnostics) {
	v := &modconfig.RawVariable{
		Name:      block.Labels[0],
		DeclRange: hclhelpers.BlockRange(block),
	}
	var diags hcl.Diagnostics

	//  set some defaults which we might override with attributes below.
	v.Type = cty.DynamicPseudoType
	v.ParsingMode = modconfig.VariableParseLiteral

	if !hclsyntax.ValidIdentifier(v.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid variable name",
			Detail:   modconfig.BadIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	if attr, exists := content.Attributes[schema.AttributeTypeTitle]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &v.Title)
		diags = append(diags, valDiags...)
	}

	if attr, exists := content.Attributes[schema.AttributeTypeDescription]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &v.Description)
		diags = append(diags, valDiags...)
		v.DescriptionSet = true
	}

	if attr, exists := content.Attributes[schema.AttributeTypeType]; exists {
		ty, tyDiags := decodeTypeExpression(attr)

		// determine the parse mode - everything but primitive types use HCL parsing
		parseMode := modconfig.VariableParseHCL
		if ty.IsPrimitiveType() {
			parseMode = modconfig.VariableParseLiteral
		}

		diags = append(diags, tyDiags...)
		v.Type = ty
		v.ParsingMode = parseMode
	}

	if attr, exists := content.Attributes[schema.AttributeTypeDefault]; exists {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)

		// Convert the default to the expected type so we can catch invalid
		// defaults early and allow later code to assume validity.
		// Note that this depends on us having already processed any "type"
		// attribute above.
		// However, we can't do this if we're in an override file where
		// the type might not be set; we'll catch that during merge.
		if v.Type != cty.NilType {
			var err error
			val, err = convert.Convert(val, v.Type)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid default value for variable",
					Detail:   fmt.Sprintf("This default value is not compatible with the variable's type constraint: %s.", err),
					Subject:  attr.Expr.Range().Ptr(),
				})
				val = cty.DynamicVal
			}
		}

		v.Default = val
	}

	if attr, exists := content.Attributes[schema.AttributeTypeEnum]; exists {
		if v.Type != cty.String && v.Type != cty.Bool && v.Type != cty.Number &&
			v.Type != cty.List(cty.String) && v.Type != cty.List(cty.Bool) && v.Type != cty.List(cty.Number) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "enum is only supported for string, bool, number, list of string, list of bool, list of number types",
				Subject:  &attr.Range,
			})
			return v, diags
		}

		ctyVal, moreDiags := attr.Expr.Value(nil)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			return v, diags
		}

		if !ctyVal.Type().IsCollectionType() && !ctyVal.Type().IsTupleType() {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "enum values must be a list",
				Subject:  &attr.Range,
			})
			return v, diags
		}

		if !hclhelpers.IsEnumValueCompatibleWithType(v.Type, ctyVal) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "enum values type mismatched",
				Subject:  &attr.Range,
			})
			return v, diags
		}

		// if there's a default, that needs to match the enum
		if v.Default != cty.NilVal {
			if !hclhelpers.IsEnumValueCompatibleWithType(v.Default.Type(), ctyVal) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "variable default value type mismatched with enum in",
					Subject:  &attr.Range,
				})
				return v, diags
			}
			valid, err := hclhelpers.ValidateSettingWithEnum(v.Default, ctyVal)

			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "error validating default value with enum",
					Subject:  &attr.Range,
				})
				return v, diags
			}

			if !valid {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "default value not in enum",
					Subject:  &attr.Range,
				})
				return v, diags
			}
		}

		v.Enum = ctyVal

		enumGo, err := hclhelpers.CtyToGo(v.Enum)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "error converting enum to go",
				Subject:  &attr.Range,
			})
			return v, diags
		}

		enumGoSlice, ok := enumGo.([]any)
		if !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "enum is not a slice",
				Subject:  &attr.Range,
			})
			return v, diags
		}

		v.EnumGo = enumGoSlice

	}

	for _, block := range content.Blocks {
		switch block.Type {

		default:
			// The above cases should be exhaustive for all block types
			// defined in variableBlockSchema
			panic(fmt.Sprintf("unhandled block type %q", block.Type))
		}
	}

	return v, diags
}
