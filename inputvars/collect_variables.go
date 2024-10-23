package inputvars

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/terraform-components/terraform"
	"github.com/turbot/terraform-components/tfdiags"
	"github.com/zclconf/go-cty/cty"
)

// CollectVariableValues inspects the various places that configuration input variable
// values can come from and constructs a map ready to be passed to the
// backend as part of a Operation.
//
// This method returns diagnostics relating to the collection of the values,
// but the values themselves may produce additional diagnostics when finally
// parsed.
func CollectVariableValues(workspacePath string, variableFileArgs []string, variablesArgs []string, workspaceModName string) (map[string]UnparsedVariableValue, error) {
	ret := map[string]UnparsedVariableValue{}

	// First we'll deal with environment variables
	// since they have the lowest precedence.
	// (apart from values in the mod Require proeprty, which are handled separately later)
	{
		env := os.Environ()
		for _, raw := range env {
			if !strings.HasPrefix(raw, app_specific.EnvInputVarPrefix) {
				continue
			}
			raw = raw[len(app_specific.EnvInputVarPrefix):] // trim the prefix

			eq := strings.Index(raw, "=")
			if eq == -1 {
				// Seems invalid, so we'll ignore it.
				continue
			}

			name := raw[:eq]
			rawVal := raw[eq+1:]

			ret[name] = UnparsedVariableValueString{
				str:        rawVal,
				name:       name,
				sourceType: terraform.ValueFromEnvVar,
			}
		}
	}

	// Next up we have some implicit files that are loaded automatically
	// if they are present. There's the original terraform.tfvars
	// (constants.DefaultVarsFilename) along with the later-added search for all files
	// ending in .auto.spvars.
	defaultVarsPath := filepaths.DefaultVarsFilePath(workspacePath)
	if _, err := os.Stat(defaultVarsPath); err == nil {
		diags := addVarsFromFile(defaultVarsPath, terraform.ValueFromAutoFile, ret)
		if diags.HasErrors() {
			return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to load variables from '%s'", defaultVarsPath), diags)
		}
	} else {
		// TACTICAL if the default vars file does not exist, check for a steampipe vars file
		legacyDefaultVarsPath := filepaths.LegacyDefaultVarsFilePath(workspacePath)

		if _, err := os.Stat(legacyDefaultVarsPath); err == nil {
			diags := addVarsFromFile(legacyDefaultVarsPath, terraform.ValueFromAutoFile, ret)
			if diags.HasErrors() {
				return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to load variables from '%s'", legacyDefaultVarsPath), diags)
			}
		}
	}

	if infos, err := os.ReadDir(workspacePath); err == nil {
		// "infos" is already sorted by name, so we just need to filter it here.
		for _, info := range infos {
			name := info.Name()
			if !isAutoVarFile(name) {
				continue
			}

			diags := addVarsFromFile(filepath.Join(workspacePath, name), terraform.ValueFromAutoFile, ret)
			if diags.HasErrors() {
				return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to load variables from '%s'", name), diags)
			}

		}
	}

	// Finally we process values given explicitly on the command line, either
	// as individual literal settings or as additional files to read.
	for _, fileArg := range variableFileArgs {
		diags := addVarsFromFile(fileArg, terraform.ValueFromNamedFile, ret)
		if diags.HasErrors() {
			return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to load variables from '%s'", fileArg), diags)
		}
	}

	var diags tfdiags.Diagnostics
	for _, variableArg := range variablesArgs {
		// Value should be in the form "name=value", where value is a
		// raw string whose interpretation will depend on the variable's
		// parsing mode.
		raw := variableArg
		eq := strings.Index(raw, "=")
		if eq == -1 {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				fmt.Sprintf("The given --var option %q is not correctly specified. It must be a variable name and value separated an equals sign: --var key=value", raw),
				"",
			))
			continue
		}

		name := raw[:eq]
		rawVal := raw[eq+1:]
		ret[name] = UnparsedVariableValueString{
			str:        rawVal,
			name:       name,
			sourceType: terraform.ValueFromCLIArg,
		}
	}

	if diags.HasErrors() {
		return nil, error_helpers.DiagsToError("failed to evaluate var args:", diags)
	}

	// check viper for any interactively added variables
	if varMap := viper.GetStringMap(constants.ConfigInteractiveVariables); varMap != nil {
		for name, rawVal := range varMap {
			// Value should be in the form "name=value", where value is a
			// raw string whose interpretation will depend on the variable's
			// parsing mode.
			ret[name] = UnparsedInteractiveVariableValue{
				Name:     name,
				RawValue: rawVal.(string),
			}
		}
	}

	// now map any variable names of form <modname>.<variablename> to <modname>.var.<varname>
	// also if any var value is qualified with the workspace mod, remove the qualification
	ret = transformVarNames(ret, workspaceModName)
	return ret, nil
}

// map any variable names of form <modname>.<variablename> to <modname>.var.<varname>
func transformVarNames(rawValues map[string]UnparsedVariableValue, workspaceModName string) map[string]UnparsedVariableValue {

	ret := make(map[string]UnparsedVariableValue, len(rawValues))
	for k, v := range rawValues {
		if parts := strings.Split(k, "."); len(parts) == 2 {
			if parts[0] == workspaceModName {
				k = parts[1]
			} else {
				k = fmt.Sprintf("%s.var.%s", parts[0], parts[1])
			}
		}
		ret[k] = v
	}
	return ret
}

func addVarsFromFile(filename string, sourceType terraform.ValueSourceType, to map[string]UnparsedVariableValue) tfdiags.Diagnostics {
	var diags tfdiags.Diagnostics

	src, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Failed to read variables file",
				fmt.Sprintf("Given variables file %s does not exist.", filename),
			))
		} else {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Failed to read variables file",
				fmt.Sprintf("Error while reading %s: %s.", filename, err),
			))
		}
		return diags
	}

	// replace syntax `<modname>.<varname>=<var_value>` with `___steampipe_<modname>_<varname>=<var_value>
	sanitisedSrc, depVarAliases := sanitiseVariableNames(src)

	var f *hcl.File
	var hclDiags hcl.Diagnostics

	// attempt to parse the config
	f, hclDiags = hclsyntax.ParseConfig(sanitisedSrc, filename, hcl.Pos{Line: 1, Column: 1})
	diags = diags.Append(hclDiags)
	if f == nil || f.Body == nil {
		return diags
	}

	// Before we do our real decode, we'll probe to see if there are any blocks
	// of type "variable" in this body, since it's a common mistake for new
	// users to put variable declarations in tfvars rather than variable value
	// definitions, and otherwise our error message for that case is not so
	// helpful.
	{
		content, _, _ := f.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "variable",
					LabelNames: []string{"name"},
				},
			},
		})
		for _, block := range content.Blocks {
			name := block.Labels[0]
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Variable declaration in .tfvars file",
				Detail:   fmt.Sprintf("A .tfvars file is used to assign values to variables that have already been declared in .tf files, not to declare new variables. To declare variable %q, place this block in one of your .tf files, such as variables.tf.\n\nTo set a value for this variable in %s, use the definition syntax instead:\n    %s = <value>", name, block.TypeRange.Filename, name),
				Subject:  &block.TypeRange,
			})
		}
		if diags.HasErrors() {
			// If we already found problems then JustAttributes below will find
			// the same problems with less-helpful messages, so we'll bail for
			// now to let the user focus on the immediate problem.
			return diags
		}
	}

	attrs, hclDiags := f.Body.JustAttributes()
	diags = diags.Append(hclDiags)

	for name, attr := range attrs {
		// check for aliases
		if alias, ok := depVarAliases[name]; ok {
			name = alias
		}
		to[name] = unparsedVariableValueExpression{
			expr:       attr.Expr,
			sourceType: sourceType,
		}
	}
	return diags
}

func sanitiseVariableNames(src []byte) ([]byte, map[string]string) {
	// replace syntax `<modname>.<varname>=<var_value>` with `____steampipe_mod_<modname>_<varname>____=<var_value>

	lines := strings.Split(string(src), "\n")
	// make map of varname aliases
	var depVarAliases = make(map[string]string)

	for i, line := range lines {

		r := regexp.MustCompile(`^ *(([a-z0-9\-_]+)\.([a-z0-9\-_]+)) *=`)
		captureGroups := r.FindStringSubmatch(line)
		if len(captureGroups) == 4 {
			fullVarName := captureGroups[1]
			mod := captureGroups[2]
			varName := captureGroups[3]

			aliasedName := fmt.Sprintf("____%s_mod_%s_variable_%s____", app_specific.AppName, mod, varName)
			depVarAliases[aliasedName] = fullVarName
			lines[i] = strings.Replace(line, fullVarName, aliasedName, 1)

		}
	}

	// now try again
	src = []byte(strings.Join(lines, "\n"))
	return src, depVarAliases
}

// unparsedVariableValueLiteral is a UnparsedVariableValue
// implementation that was actually already parsed (!). This is
// intended to deal with expressions inside "tfvars" files.
type unparsedVariableValueExpression struct {
	expr       hcl.Expression
	sourceType terraform.ValueSourceType
}

func (v unparsedVariableValueExpression) ParseVariableValue(evalCtx *hcl.EvalContext, mode modconfig.VariableParsingMode) (*terraform.InputValue, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	val, hclDiags := v.expr.Value(evalCtx)
	diags = diags.Append(hclDiags)

	rng := tfdiags.SourceRangeFromHCL(v.expr.Range())

	return &terraform.InputValue{
		Value:       val,
		SourceType:  v.sourceType,
		SourceRange: rng,
	}, diags
}

func (v unparsedVariableValueExpression) ParseVariableValueToType(evalCtx *hcl.EvalContext, mode modconfig.VariableParsingMode, targetType cty.Type) (*terraform.InputValue, tfdiags.Diagnostics) {
	return v.ParseVariableValue(evalCtx, mode)
}

// UnparsedVariableValueString is a UnparsedVariableValue
// implementation that parses its value from a string. This can be used
// to deal with values given directly on the command line and via environment
// variables.
type UnparsedVariableValueString struct {
	str        string
	name       string
	sourceType terraform.ValueSourceType
}

func (v UnparsedVariableValueString) Raw() string {
	return v.str
}

func (v UnparsedVariableValueString) ParseVariableValue(evalCtx *hcl.EvalContext, mode modconfig.VariableParsingMode) (*terraform.InputValue, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	val, hclDiags := mode.Parse(evalCtx, v.name, v.str)
	diags = diags.Append(hclDiags)

	return &terraform.InputValue{
		Value:      val,
		SourceType: v.sourceType,
	}, diags
}

func (v UnparsedVariableValueString) ParseVariableValueToType(evalCtx *hcl.EvalContext, mode modconfig.VariableParsingMode, targetType cty.Type) (*terraform.InputValue, tfdiags.Diagnostics) {
	return v.ParseVariableValue(evalCtx, mode)
}

// isAutoVarFile determines if the file ends with .auto.spvars or .auto.spvars.json
func isAutoVarFile(path string) bool {
	for _, ext := range app_specific.AutoVariablesExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
