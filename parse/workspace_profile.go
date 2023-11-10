package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"log"
)

func LoadWorkspaceProfiles[T modconfig.WorkspaceProfile](workspaceProfilePath string) (profileMap map[string]T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}

	}()

	// create profile map to populate
	profileMap = make(map[string]T)

	configPaths, err := filehelpers.ListFiles(workspaceProfilePath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
	})
	if err != nil {
		return nil, err
	}
	if len(configPaths) == 0 {
		return profileMap, nil
	}

	fileData, diags := LoadFileData(configPaths...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, diags := body.Content(ConfigBlockSchema)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	parseCtx := NewWorkspaceProfileParseContext[T](workspaceProfilePath)
	parseCtx.SetDecodeContent(content, fileData)

	// build parse context
	return parseWorkspaceProfiles(parseCtx)

}

func parseWorkspaceProfiles[T modconfig.WorkspaceProfile](parseCtx *WorkspaceProfileParseContext[T]) (map[string]T, error) {
	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		_, diags := decodeWorkspaceProfiles(parseCtx)
		if diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to decode all workspace profile files", diags)
		}

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			log.Printf("[TRACE] parse complete after %d decode passes", attempts+1)
			break
		}
		// if the number of unresolved blocks has NOT reduced, fail
		if prevUnresolvedBlocks != 0 && unresolvedBlocks >= prevUnresolvedBlocks {
			str := parseCtx.FormatDependencies()
			return nil, fmt.Errorf("failed to resolve workspace profile dependencies after %d attempts\nDependencies:\n%s", attempts+1, str)
		}
		// update prevUnresolvedBlocks
		prevUnresolvedBlocks = unresolvedBlocks
	}

	return parseCtx.workspaceProfiles, nil

}

func decodeWorkspaceProfiles[T modconfig.WorkspaceProfile](parseCtx *WorkspaceProfileParseContext[T]) (map[string]T, hcl.Diagnostics) {
	profileMap := make(map[string]T)

	var diags hcl.Diagnostics
	blocksToDecode, err := parseCtx.BlocksToDecode()
	// build list of blocks to decode
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
		return nil, diags
	}

	// now clear dependencies from run context - they will be rebuilt
	parseCtx.ClearDependencies()

	for _, block := range blocksToDecode {
		if block.Type == schema.BlockTypeWorkspaceProfile {
			workspaceProfile, res := decodeWorkspaceProfile[T](block, parseCtx)

			if res.Success() {
				// success - add to map
				profileMap[workspaceProfile.ShortName()] = workspaceProfile
			}
			diags = append(diags, res.Diags...)
		}
	}
	return profileMap, diags
}

func decodeWorkspaceProfile[T modconfig.WorkspaceProfile](block *hcl.Block, parseCtx *WorkspaceProfileParseContext[T]) (T, *DecodeResult) {
	var emptyProfile T
	res := newDecodeResult()
	// get shell resource
	resource, diags := modconfig.NewWorkspaceProfile[T](block)
	if diags.HasErrors() {
		res.handleDecodeDiags(diags)
		return emptyProfile, res
	}

	// do a partial decode to get options blocks into workspaceProfileOptions, with all other attributes in rest
	workspaceProfileOptions, rest, diags := block.Body.PartialContent(WorkspaceProfileBlockSchema)
	if diags.HasErrors() {
		res.handleDecodeDiags(diags)
		return emptyProfile, res
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, resource)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
	}
	// lookup of options blocks
	foundOptions := map[string]struct{}{}
	for _, block := range workspaceProfileOptions.Blocks {
		switch block.Type {
		case "options":
			optionsBlockType := block.Labels[0]
			if _, found := foundOptions[optionsBlockType]; found {
				// fail
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Subject:  hclhelpers.BlockRangePointer(block),
					Summary:  fmt.Sprintf("Duplicate options type '%s'", optionsBlockType),
				})
			}
			opts, moreDiags := DecodeOptions(block, resource.GetOptionsForBlock)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			moreDiags = resource.SetOptions(opts, block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
			foundOptions[optionsBlockType] = struct{}{}
		default:
			// this should never happen
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for workspace profiles", block.Type),
				Subject:  hclhelpers.BlockRangePointer(block),
			})
		}
	}

	res.addDiags(diags)

	handleWorkspaceProfileDecodeResult(resource, res, block, parseCtx)
	return resource, res
}

func handleWorkspaceProfileDecodeResult[T modconfig.WorkspaceProfile](resource T, res *DecodeResult, block *hcl.Block, parseCtx *WorkspaceProfileParseContext[T]) {
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded()
		res.addDiags(moreDiags)

		moreDiags = parseCtx.AddResource(resource)
		res.addDiags(moreDiags)
		return
	}

	// failure :(
	if len(res.Depends) > 0 {
		moreDiags := parseCtx.AddDependencies(block, resource.Name(), res.Depends)
		res.addDiags(moreDiags)
	}
}
