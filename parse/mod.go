package parse

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/perr"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/zclconf/go-cty/cty"
)

func LoadModfile(modPath string) (*modconfig.Mod, error) {
	if !ModfileExists(modPath) {
		return nil, nil
	}

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: funcs.ContextFunctions(modPath),
		Variables: make(map[string]cty.Value),
	}

	mod, res := ParseModDefinition(modPath, evalCtx)
	if res.Diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load mod", res.Diags)
	}

	return mod, nil
}

func ParseModDefinitionWithFileName(modPath string, modFileName string, evalCtx *hcl.EvalContext) (*modconfig.Mod, *DecodeResult) {
	res := newDecodeResult()

	// if there is no mod at this location, return error
	modFilePath := filepath.Join(modPath, modFileName)

	modFileFound := true
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		modFileFound = false
		modFilePath = filepath.Join(modPath, app_specific.ModFileName)
		if _, err := os.Stat(modFilePath); err == nil {
			modFileFound = true
		}
	}

	if !modFileFound {
		res.Diags = append(res.Diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("no mod file found in %s", modPath),
		})
		return nil, res
	}

	fileData, diags := LoadFileData(modFilePath)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	body, diags := ParseHclFiles(fileData)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	workspaceContent, diags := body.Content(WorkspaceBlockSchema)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	block := hclhelpers.GetFirstBlockOfType(workspaceContent.Blocks, schema.BlockTypeMod)
	if block == nil {
		res.Diags = append(res.Diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("no mod definition found in %s", modPath),
		})
		return nil, res
	}
	var defRange = hclhelpers.BlockRange(block)
	mod := modconfig.NewMod(block.Labels[0], path.Dir(modFilePath), defRange)
	// set modFilePath
	mod.SetFilePath(modFilePath)

	mod, res = decodeMod(block, evalCtx, mod)
	if res.Diags.HasErrors() {
		return nil, res
	}

	// NOTE: IGNORE DEPENDENCY ERRORS

	// call decode callback
	diags = mod.OnDecoded(block, nil)
	res.addDiags(diags)

	return mod, res
}

// ParseModDefinition parses the modfile only
// it is expected the calling code will have verified the existence of the modfile by calling ModfileExists
// this is called before parsing the workspace to, for example, identify dependency mods
//
// This function only parse the "mod" block, and does not parse any resources in the mod file
func ParseModDefinition(modPath string, evalCtx *hcl.EvalContext) (*modconfig.Mod, *DecodeResult) {
	return ParseModDefinitionWithFileName(modPath, app_specific.ModFileName, evalCtx)
}

// ParseMod parses all source hcl files for the mod path and associated resources, and returns the mod object
// NOTE: the mod definition has already been parsed (or a default created) and is in opts.RunCtx.RootMod
func ParseMod(ctx context.Context, fileData map[string][]byte, pseudoResources []modconfig.MappableResource, parseCtx *ModParseContext) (*modconfig.Mod, *error_helpers.ErrorAndWarnings) {
	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to load all mod source files", diags))
	}

	content, moreDiags := body.Content(WorkspaceBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to load mod", diags))
	}

	mod := parseCtx.CurrentMod
	if mod == nil {
		return nil, error_helpers.NewErrorsAndWarning(fmt.Errorf("ParseMod called with no Current Mod set in ModParseContext"))
	}
	// get names of all resources defined in hcl which may also be created as pseudo resources
	hclResources, err := loadMappableResourceNames(content)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// if variables were passed in parsecontext, add to the mod
	if parseCtx.Variables != nil {
		for _, v := range parseCtx.Variables.RootVariables {
			if diags = mod.AddResource(v); diags.HasErrors() {
				return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to add resource to mod", diags))
			}
		}
	}

	// collect warnings as we parse
	var res = &error_helpers.ErrorAndWarnings{}

	// add pseudo resources to the mod
	errorsAndWarnings := addPseudoResourcesToMod(pseudoResources, hclResources, mod)

	// merge the warnings generated while adding pseudoresources
	res.Merge(errorsAndWarnings)

	// add the parsed content to the run context
	parseCtx.SetDecodeContent(content, fileData)

	// add the mod to the run context
	// - this it to ensure all pseudo resources get added and build the eval context with the variables we just added

	// ! This is the place where the child mods (dependent mods) resources are "pulled up" into this current evaluation
	// ! context.
	// !
	// ! Step through the code to find the place where the child mod resources are added to the "referencesValue"
	// !
	// ! Note that this resource MUST implement ModItem interface, otherwise it will look "flat", i.e. it will be added
	// ! to the current mod
	// !
	// ! There's also a bug where we test for ModTreeItem, we added a new interface ModItem for resources that are mod
	// ! resources but not necessarily need to be in the mod tree
	// !
	if diags = parseCtx.AddModResources(mod); diags.HasErrors() {
		return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to add mod to run context", diags))
	}

	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		diags = decode(parseCtx)
		if diags.HasErrors() {
			return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to decode mod", diags))
		}
		// now retrieve the warning strings
		res.AddWarning(plugin.DiagsToWarnings(diags)...)

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			log.Printf("[TRACE] parse complete after %d decode passes", attempts+1)
			break
		}
		// if the number of unresolved blocks has NOT reduced, fail
		if prevUnresolvedBlocks != 0 && unresolvedBlocks >= prevUnresolvedBlocks {
			str := parseCtx.FormatDependencies()
			msg := fmt.Sprintf("Failed to resolve dependencies after %d passes. Unresolved blocks:\n%s", attempts+1, str)
			return nil, error_helpers.NewErrorsAndWarning(perr.BadRequestWithTypeAndMessage(perr.ErrorCodeDependencyFailure, msg))
		}
		// update prevUnresolvedBlocks
		prevUnresolvedBlocks = unresolvedBlocks
	}

	// now tell mod to build tree of resources
	res.Error = mod.BuildResourceTree(parseCtx.GetTopLevelDependencyMods())

	return mod, res
}
