package load_mod

import (
	"context"
	"sort"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/inputvars"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/versionmap"
	"github.com/turbot/terraform-components/terraform"
	"github.com/turbot/terraform-components/tfdiags"
	"golang.org/x/exp/maps"
)

func LoadVariableDefinitions(ctx context.Context, variablePath string, parseCtx *parse.ModParseContext) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	mod, ew := LoadMod(ctx, variablePath, parseCtx)
	if ew.GetError() != nil {
		return nil, ew
	}

	m, err := modconfig.NewModVariableMap(mod)
	ew.Error = err
	return m, ew

}

func GetVariableValues(parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, validate bool) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	// now resolve all input variables
	inputValues, errorsAndWarnings := getInputVariables(parseCtx, variableMap, validate)
	if errorsAndWarnings.Error == nil {
		// now update the variables map with the input values
		err := inputvars.SetVariableValues(inputValues, variableMap)
		if err != nil {
			errorsAndWarnings.Error = err
		}
	}

	return variableMap, errorsAndWarnings
}

func getInputVariables(parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, validate bool) (terraform.InputValues, error_helpers.ErrorAndWarnings) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	// get mod and mod path from run context
	mod := parseCtx.CurrentMod
	path := mod.ModPath

	var inputValuesUnparsed, err = inputvars.CollectVariableValues(path, variableFileArgs, variableArgs, parseCtx.CurrentMod.ShortName)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	if validate {
		if err := identifyAllMissingVariables(parseCtx, variableMap, inputValuesUnparsed); err != nil {
			return nil, error_helpers.NewErrorsAndWarning(err)
		}
	}

	// read any args set in the mod require block
	depModArgs, err := inputvars.CollectVariableValuesFromModRequire(variableMap.Mod, parseCtx.WorkspaceLock)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// parse the input values (only parse values for public variables)
	// NOTE: pass in variable values set in mod require block to ensure validation passes
	parsedValues, diags := inputvars.ParseVariableValues(parseCtx.EvalCtx, inputValuesUnparsed, depModArgs, variableMap, validate)

	if validate {
		moreDiags := inputvars.CheckInputVariables(variableMap.PublicVariables, parsedValues)
		diags = append(diags, moreDiags...)
	}

	return parsedValues, newVariableValidationResult(diags)
}

func newVariableValidationResult(diags tfdiags.Diagnostics) error_helpers.ErrorAndWarnings {
	warnings := error_helpers.HclDiagsToWarnings(diags.ToHCL())
	var err error
	if diags.HasErrors() {
		err = steampipeconfig.NewVariableValidationFailedError(diags)
	}
	return error_helpers.NewErrorsAndWarning(err, warnings...)
}

func identifyAllMissingVariables(parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, variableValues map[string]inputvars.UnparsedVariableValue) error {
	// convert variableValues into a lookup
	var variableValueLookup = utils.SliceToLookup(maps.Keys(variableValues))
	missingVarsMap, err := identifyMissingVariablesForDependencyTree(parseCtx.WorkspaceLock, variableMap, variableValueLookup, nil)

	if err != nil {
		return err
	}
	if len(missingVarsMap) == 0 {
		// all good
		return nil
	}

	// build a MissingVariableError
	missingVarErr := steampipeconfig.NewMissingVarsError(parseCtx.CurrentMod)

	// build a lookup with the dependency path of the root mod and all top level dependencies
	rootName := variableMap.Mod.ShortName
	topLevelModLookup := map[steampipeconfig.DependencyPathKey]struct{}{steampipeconfig.DependencyPathKey(rootName): {}}
	for dep := range parseCtx.WorkspaceLock.InstallCache {
		depPathKey := steampipeconfig.NewDependencyPathKey(rootName, dep)
		topLevelModLookup[depPathKey] = struct{}{}
	}
	for depPath, missingVars := range missingVarsMap {
		if _, isTopLevel := topLevelModLookup[depPath]; isTopLevel {
			missingVarErr.MissingVariables = append(missingVarErr.MissingVariables, missingVars...)
		} else {
			missingVarErr.MissingTransitiveVariables[depPath] = missingVars
		}
	}

	return missingVarErr
}

func identifyMissingVariablesForDependencyTree(workspaceLock *versionmap.WorkspaceLock, variableMap *modconfig.ModVariableMap, parentVariableValuesLookup map[string]struct{}, dependencyPath []string) (map[steampipeconfig.DependencyPathKey][]*modconfig.Variable, error) {
	// return a map of missing variables, keyed by dependency path
	res := make(map[steampipeconfig.DependencyPathKey][]*modconfig.Variable)

	// update the path to this dependency
	dependencyPath = append(dependencyPath, variableMap.Mod.GetInstallCacheKey())

	// clone parentVariableValuesLookup so we can mutate it with dependency specific args overrides
	var variableValueLookup = make(map[string]struct{}, len(parentVariableValuesLookup))
	for k := range parentVariableValuesLookup {
		// convert the variable name to the short name if it is fully qualified and belongs to the current mod
		k = getVariableValueMapKey(k, variableMap)

		// add into lookup
		variableValueLookup[k] = struct{}{}
	}

	// first get any args specified in the mod requires
	// note the actual value of these may be unknown as we have not yet resolved
	depModArgs, err := inputvars.CollectVariableValuesFromModRequire(variableMap.Mod, workspaceLock)
	for varName := range depModArgs {
		// convert the variable name to the short name if it is fully qualified and belongs to the current mod
		varName = getVariableValueMapKey(varName, variableMap)
		variableValueLookup[varName] = struct{}{}
	}
	if err != nil {
		return nil, err
	}

	//  handle root variables
	missingVariables := identifyMissingVariables(variableMap.RootVariables, variableValueLookup)
	if len(missingVariables) > 0 {
		res[steampipeconfig.NewDependencyPathKey(dependencyPath...)] = missingVariables
	}

	// now iterate through all the dependency variable maps
	for _, dependencyVariableMap := range variableMap.DependencyVariables {
		childMissingMap, err := identifyMissingVariablesForDependencyTree(workspaceLock, dependencyVariableMap, variableValueLookup, dependencyPath)
		if err != nil {
			return nil, err
		}
		// add results into map
		for k, v := range childMissingMap {
			res[k] = v
		}
	}
	return res, nil
}

// getVariableValueMapKey checks whether the variable is fully qualified and belongs to the current mod,
// if so use the short name
func getVariableValueMapKey(k string, variableMap *modconfig.ModVariableMap) string {
	// attempt to parse the variable name.
	// Note: if the variable is not fully qualified (e.g. "var_name"),  ParseResourceName will return an error
	// in which case we add it to our map unchanged
	parsedName, err := modconfig.ParseResourceName(k)
	// if this IS a dependency variable, the parse will success
	// if the mod name is the same as the current mod (variableMap.Mod)
	// then add a map entry with the variable short name
	// this will allow us to match the variable value to a variable defined in this mod
	if err == nil && parsedName.Mod == variableMap.Mod.ShortName {
		k = parsedName.Name
	}
	return k
}

func identifyMissingVariables(variableMap map[string]*modconfig.Variable, variableValuesLookup map[string]struct{}) []*modconfig.Variable {

	var needed []*modconfig.Variable

	for name, v := range variableMap {
		if !v.Required() {
			continue // We only prompt for required variables
		}
		_, unparsedValExists := variableValuesLookup[name]

		if !unparsedValExists {
			needed = append(needed, v)
		}
	}
	sort.SliceStable(needed, func(i, j int) bool {
		return needed[i].Name() < needed[j].Name()
	})
	return needed

}
