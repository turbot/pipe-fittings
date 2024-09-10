package var_config

// github.com/turbot/terraform-components/configs/parser_config.go
import (
	"fmt"
	"unicode"

	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const badIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."

// Variable represents a "variable" block in a module or file.
type Variable struct {
	Name           string
	Title          string
	Description    string
	Default        cty.Value
	Type           cty.Type
	ParsingMode    VariableParsingMode
	Enum           cty.Value
	EnumGo         []any
	DescriptionSet bool
	DeclRange      hcl.Range
	Subtype        hcl.Expression
	SubtypeString  string
}

func DecodeVariableBlock(block *hcl.Block, content *hcl.BodyContent, override bool) (*Variable, hcl.Diagnostics) {
	v := &Variable{
		Name:      block.Labels[0],
		DeclRange: hclhelpers.BlockRange(block),
	}
	var diags hcl.Diagnostics

	// Unless we're building an override, we'll set some defaults
	// which we might override with attributes below. We leave these
	// as zero-value in the override case so we can recognize whether
	// or not they are set when we merge.
	if !override {
		v.Type = cty.DynamicPseudoType
		v.ParsingMode = VariableParseLiteral
	}

	if !hclsyntax.ValidIdentifier(v.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid variable name",
			Detail:   badIdentifierDetail,
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
		ty, parseMode, tyDiags := decodeVariableType(attr.Expr)
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

	if attr, exists := content.Attributes[schema.AttributeTypeSubtype]; exists {
		v.Subtype = attr.Expr
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

func decodeVariableType(expr hcl.Expression) (cty.Type, VariableParsingMode, hcl.Diagnostics) {
	if exprIsNativeQuotedString(expr) {
		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return cty.DynamicPseudoType, VariableParseHCL, diags
		}
		str := val.AsString()
		switch str {
		case "string":
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid quoted type constraints",
				Subject:  expr.Range().Ptr(),
			})
			return cty.DynamicPseudoType, VariableParseLiteral, diags
		case "list":
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid quoted type constraints",
				Subject:  expr.Range().Ptr(),
			})
			return cty.DynamicPseudoType, VariableParseHCL, diags
		case "map":
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid quoted type constraints",
				Subject:  expr.Range().Ptr(),
			})
			return cty.DynamicPseudoType, VariableParseHCL, diags
		default:
			return cty.DynamicPseudoType, VariableParseHCL, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Invalid legacy variable type hint",
				Subject:  expr.Range().Ptr(),
			}}
		}
	}

	// First we'll deal with some shorthand forms that the HCL-level type
	// expression parser doesn't include. These both emulate pre-0.12 behavior
	// of allowing a list or map of any element type as long as all of the
	// elements are consistent. This is the same as list(any) or map(any).
	switch hcl.ExprAsKeyword(expr) {
	case "list":
		return cty.List(cty.DynamicPseudoType), VariableParseHCL, nil
	case "map":
		return cty.Map(cty.DynamicPseudoType), VariableParseHCL, nil
	}

	ty, diags := typeexpr.TypeConstraint(expr)
	if diags.HasErrors() {
		return cty.DynamicPseudoType, VariableParseHCL, diags
	}

	switch {
	case ty.IsPrimitiveType():
		// Primitive types use literal parsing.
		return ty, VariableParseLiteral, diags
	default:
		// Everything else uses HCL parsing
		return ty, VariableParseHCL, diags
	}
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}

// VariableParsingMode defines how values of a particular variable given by
// text-only mechanisms (command line arguments and environment variables)
// should be parsed to produce the final value.
type VariableParsingMode rune

// VariableParseLiteral is a variable parsing mode that just takes the given
// string directly as a cty.String value.
const VariableParseLiteral VariableParsingMode = 'L'

// VariableParseHCL is a variable parsing mode that attempts to parse the given
// string as an HCL expression and returns the result.
const VariableParseHCL VariableParsingMode = 'H'

// Parse uses the receiving parsing mode to process the given variable value
// string, returning the result along with any diagnostics.
//
// A VariableParsingMode does not know the expected type of the corresponding
// variable, so it's the caller's responsibility to attempt to convert the
// result to the appropriate type and return to the user any diagnostics that
// conversion may produce.
//
// The given name is used to create a synthetic filename in case any diagnostics
// must be generated about the given string value. This should be the name
// of the configuration variable whose value will be populated from the given
// string.
//
// If the returned diagnostics has errors, the returned value may not be
// valid.
func (m VariableParsingMode) Parse(name, value string) (cty.Value, hcl.Diagnostics) {
	switch m {
	case VariableParseLiteral:
		return cty.StringVal(value), nil
	case VariableParseHCL:
		fakeFilename := fmt.Sprintf("<value for var.%s>", name)
		expr, diags := hclsyntax.ParseExpression([]byte(value), fakeFilename, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return cty.DynamicVal, diags
		}
		val, valDiags := expr.Value(nil)
		diags = append(diags, valDiags...)
		return val, diags
	default:
		// Should never happen
		panic(fmt.Errorf("parse called on invalid VariableParsingMode %#v", m))
	}
}

// VariableValidation represents a configuration-defined validation rule
// for a particular input variable, given as a "validation" block inside
// a "variable" block.
type VariableValidation struct {
	// Condition is an expression that refers to the variable being tested
	// and contains no other references. The expression must return true
	// to indicate that the value is valid or false to indicate that it is
	// invalid. If the expression produces an error, that's considered a bug
	// in the module defining the validation rule, not an error in the caller.
	Condition hcl.Expression

	// ErrorMessage is one or more full sentences, which would need to be in
	// English for consistency with the rest of the error message output but
	// can in practice be in any language as long as it ends with a period.
	// The message should describe what is required for the condition to return
	// true in a way that would make sense to a caller of the module.
	ErrorMessage string

	DeclRange hcl.Range
}

// looksLikeSentence is a simple heuristic that encourages writing error
// messages that will be presentable when included as part of a larger error diagnostic
func looksLikeSentences(s string) bool {
	if len(s) < 1 {
		return false
	}
	runes := []rune(s) // HCL guarantees that all strings are valid UTF-8
	first := runes[0]
	last := runes[len(runes)-1]

	// If the first rune is a letter then it must be an uppercase letter.
	// (This will only see the first rune in a multi-rune combining sequence,
	// but the first rune is generally the letter if any are, and if not then
	// we'll just ignore it because we're primarily expecting English messages
	// right now anyway)
	if unicode.IsLetter(first) && !unicode.IsUpper(first) {
		return false
	}

	// The string must be at least one full sentence, which implies having
	// sentence-ending punctuation.
	return last == '.' || last == '?' || last == '!'
}
