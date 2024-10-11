package modconfig

// github.com/turbot/terraform-components/configs/parser_config.go
import (
	"fmt"
	"unicode"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const BadIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."

// RawVariable represents a "variable" block in a module or file.
type RawVariable struct {
	Name           string
	Title          string
	Description    string
	Default        cty.Value
	Type           cty.Type
	TypeString     string
	ParsingMode    VariableParsingMode
	Enum           cty.Value
	EnumGo         []any
	DescriptionSet bool
	Format         string
	DeclRange      hcl.Range
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *RawVariable) Required() bool {
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
func (m VariableParsingMode) Parse(evalCtx *hcl.EvalContext, name, value string) (cty.Value, hcl.Diagnostics) {
	switch m {
	case VariableParseLiteral:
		return cty.StringVal(value), nil
	case VariableParseHCL:
		fakeFilename := fmt.Sprintf("<value for var.%s>", name)
		expr, diags := hclsyntax.ParseExpression([]byte(value), fakeFilename, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return cty.DynamicVal, diags
		}
		val, valDiags := expr.Value(evalCtx)
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
