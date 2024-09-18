package inputvars

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/terraform-components/terraform"
	"github.com/turbot/terraform-components/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/maps"
)

// UnparsedVariableValue represents a variable value provided by the caller
// whose parsing must be deferred until configuration is available.
//
// This exists to allow processing of variable-setting arguments (e.g. in the
// command package) to be separated from parsing (in the backend package).
type UnparsedVariableValue interface {
	// ParseVariableValue information in the provided variable configuration
	// to parse (if necessary) and return the variable value encapsulated in
	// the receiver.
	//
	// If error diagnostics are returned, the resulting value may be invalid
	// or incomplete.
	ParseVariableValue(mode modconfig.VariableParsingMode) (*terraform.InputValue, tfdiags.Diagnostics)
}

// ParseVariableValues processes a map of unparsed variable values by
// correlating each one with the given variable declarations which should
// be from a configuration.
//
// # It also takes into account parsed values which are declares in the mod require block
//
// The map of unparsed variable values should include variables from all
// possible configuration declarations sources such that it is as complete as
// it can possibly be for the current operation. If any declared variables
// are not included in the map, ParseVariableValues will either substitute
// a configured default value or produce an error.
func ParseVariableValues(inputValueUnparsed map[string]UnparsedVariableValue, requireArgs terraform.InputValues, variablesMap *modconfig.ModVariableMap, validate bool) (terraform.InputValues, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	ret := make(terraform.InputValues, len(inputValueUnparsed)*len(requireArgs))

	// initialise ret to contain any values that were set in the mod Require block using Args property
	maps.Copy(ret, requireArgs)

	publicVariables := variablesMap.PublicVariables

	// Currently we're generating only warnings for undeclared variables
	// defined in files (see below) but we only want to generate a few warnings
	// at a time because existing deployments may have lots of these and
	// the result can therefore be overwhelming.
	seenUndeclaredInFile := 0

	for name, unparsedVal := range inputValueUnparsed {
		var mode modconfig.VariableParsingMode
		config, declared := publicVariables[name]
		if declared {
			mode = config.ParsingMode
		} else {
			mode = modconfig.VariableParseLiteral
		}

		val, valDiags := unparsedVal.ParseVariableValue(mode)
		diags = diags.Append(valDiags)
		if valDiags.HasErrors() {
			continue
		}

		if !declared {
			switch val.SourceType {
			case terraform.ValueFromConfig, terraform.ValueFromAutoFile, terraform.ValueFromNamedFile:
				// We allow undeclared names for variable values from files and warn in case
				// users have forgotten a variable {} declaration or have a typo in their var name.
				if seenUndeclaredInFile < 2 {
					diags = diags.Append(tfdiags.Sourceless(
						tfdiags.Warning,
						"Value for undeclared variable",
						getUndeclaredVariableError(name, variablesMap), //, val.SourceRange.Filename),
					))
				}
				seenUndeclaredInFile++

			case terraform.ValueFromEnvVar:
				// We allow and ignore undeclared names for environment
				// variables, because users will often set these globally
				// when they are used across many (but not necessarily all)
				// configurations.
			case terraform.ValueFromCLIArg:
				diags = diags.Append(tfdiags.Sourceless(
					tfdiags.Error,
					"Value for undeclared variable",
					getUndeclaredVariableError(name, variablesMap),
				))
			default:
				// For all other source types we are more vague, but other situations
				// don't generally crop up at this layer in practice.
				diags = diags.Append(tfdiags.Sourceless(
					tfdiags.Error,
					"Value for undeclared variable",
					getUndeclaredVariableError(name, variablesMap),
				))
			}
			continue
		}

		ret[name] = val
	}

	if seenUndeclaredInFile > 2 {
		extras := seenUndeclaredInFile - 2
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "Values for undeclared variables",
			Detail:   fmt.Sprintf("In addition to the other similar warnings shown, %d other variable(s) defined without being declared.", extras),
		})
	}

	// By this point we should've gathered all of the required variables
	// from one of the many possible sources.
	// We'll now populate any we haven't gathered as their defaults and fail if any of the
	// missing ones are required.
	for name, vc := range publicVariables {
		if _, defined := ret[name]; defined {
			continue
		}

		//  are we missing a required variable?
		if vc.Required() {

			// We'll include a placeholder value anyway, just so that our
			// result is complete for any calling code that wants to cautiously
			// analyze it for diagnostic purposes. Since our diagnostics now
			// includes an error, normal processing will ignore this result.
			ret[name] = &terraform.InputValue{
				Value:       cty.DynamicVal,
				SourceType:  terraform.ValueFromConfig,
				SourceRange: tfdiags.SourceRangeFromHCL(vc.DeclRange),
			}

			// if validation flag is set, raise an error
			if validate {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No value for required variable",
					Detail:   fmt.Sprintf("The input variable %q is not set, and has no default value. Use a --var or --var-file command line argument to provide a value for this variable.", name),
					Subject:  vc.DeclRange.Ptr(),
				})
			}
		} else {
			// not required - use default
			ret[name] = &terraform.InputValue{
				Value:       vc.Default,
				SourceType:  terraform.ValueFromConfig,
				SourceRange: tfdiags.SourceRangeFromHCL(vc.DeclRange),
			}
		}
	}

	return ret, diags
}

func getUndeclaredVariableError(name string, variablesMap *modconfig.ModVariableMap) string {
	// is this a qualified variable?
	if len(strings.Split(name, ".")) == 1 {
		// unqualifid
		return fmt.Sprintf("\"%s\". If you meant to use this value, add a \"variable\" block to the mod.\n", name)
	}

	// parse to extract the mod name
	parsedVarName, err := modconfig.ParseResourceName(name)
	if err != nil {
		return fmt.Sprintf("Invalid variable name: \"%s\". It should be of form \"var_name\" or \"mod_name.var_name\".", name)
	}

	for _, m := range variablesMap.Mod.ResourceMaps.Mods {
		if m.ShortName == parsedVarName.Mod {
			return fmt.Sprintf("\"%s\": Dependency mod \"%s\" has no variable \"%s\"", name, m.DependencyName, parsedVarName.Name)
			// so it is a dependency mod
		}
	}
	return fmt.Sprintf("\"%s\": Mod \"%s\" is not a dependency of the current workspace.", name, parsedVarName.Mod)

}

type UnparsedInteractiveVariableValue struct {
	Name, RawValue string
}

func (v UnparsedInteractiveVariableValue) ParseVariableValue(mode modconfig.VariableParsingMode) (*terraform.InputValue, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	val, valDiags := mode.Parse(v.Name, v.RawValue)
	diags = diags.Append(valDiags)
	if diags.HasErrors() {
		return nil, diags
	}
	return &terraform.InputValue{
		Value:      val,
		SourceType: terraform.ValueFromInput,
	}, diags
}
