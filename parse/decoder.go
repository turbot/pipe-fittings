package parse

import (
	"fmt"
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type DecoderOption func(Decoder)

type Decoder interface {
	Decode(*ModParseContext) hcl.Diagnostics
	ShouldAddToMod(modconfig.HclResource, *hcl.Block, *ModParseContext) bool
}

type DecodeFunc func(*hcl.Block, *ModParseContext) (modconfig.HclResource, *DecodeResult)

type DecoderImpl struct {
	// registered block types
	DecodeFuncs map[string]DecodeFunc
	// optional default decode function
	DefaultDecodeFunc DecodeFunc
	// optional resource validation func
	ValidateFunc func(resource modconfig.HclResource) hcl.Diagnostics
}

func NewDecoderImpl() DecoderImpl {
	d := DecoderImpl{
		DecodeFuncs: make(map[string]DecodeFunc),
	}
	d.DecodeFuncs[schema.BlockTypeVariable] = d.decodeVariable
	return d
}

func (d *DecoderImpl) Decode(parseCtx *ModParseContext) hcl.Diagnostics {
	utils.LogTime(fmt.Sprintf("decode %s start", parseCtx.CurrentMod.Name()))
	defer utils.LogTime(fmt.Sprintf("decode %s end", parseCtx.CurrentMod.Name()))

	var diags hcl.Diagnostics

	blocks, err := parseCtx.BlocksToDecode()
	// build list of blocks to decode
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
		return diags
	}

	// now clear dependencies from run context - they will be rebuilt
	parseCtx.ClearDependencies()

	for _, block := range blocks {
		utils.LogTime(fmt.Sprintf("decode block %s - %v start", block.Type, block.Labels))
		switch block.Type {
		case schema.BlockTypeLocals:
			resources, res := d.decodeLocalsBlock(block, parseCtx)
			if !res.Success() {
				diags = append(diags, res.Diags...)
				continue
			}
			for _, resource := range resources {
				resourceDiags := AddResourceToMod(resource, block, d, parseCtx)
				diags = append(diags, resourceDiags...)
			}
		default:
			resource, res := d.DecodeBlock(block, parseCtx)
			diags = append(diags, res.Diags...)
			if !res.Success() || resource == nil {
				continue
			}

			resourceDiags := AddResourceToMod(resource, block, d, parseCtx)
			diags = append(diags, resourceDiags...)
		}
		utils.LogTime(fmt.Sprintf("decode block %s - %v end", block.Type, block.Labels))
	}

	return diags
}

// special case decode logic for locals
func (d *DecoderImpl) decodeLocalsBlock(block *hcl.Block, parseCtx *ModParseContext) ([]modconfig.HclResource, *DecodeResult) {
	var resources []modconfig.HclResource
	var res = NewDecodeResult()

	// check name is valid
	diags := ValidateName(block)
	if diags.HasErrors() {
		res.AddDiags(diags)
		return nil, res
	}

	var locals []*modconfig.Local
	locals, res = d.decodeLocals(block, parseCtx)
	for _, local := range locals {
		resources = append(resources, local)
		d.HandleModDecodeResult(local, res, block, parseCtx)
	}

	return resources, res
}

func (d *DecoderImpl) DecodeBlock(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	var resource modconfig.HclResource
	var res = NewDecodeResult()

	// has this block already been decoded?
	// (this could happen if it is a child block and has been decoded before its parent as part of second decode phase)
	if resource, ok := parseCtx.GetDecodedResourceForBlock(block); ok {
		return resource, res
	}

	// check name is valid
	diags := ValidateName(block)
	if diags.HasErrors() {
		res.AddDiags(diags)
		return nil, res
	}

	decodeFunc, moreDiags := d.getDecodeFunc(block)
	if diags.HasErrors() {
		res.AddDiags(moreDiags)
		return nil, res
	}
	resource, res = decodeFunc(block, parseCtx)
	// Note that an interface value that holds a nil concrete value is itself non-nil.
	if !helpers.IsNil(resource) {
		// handle the result
		// - if there are dependencies, add to run context
		d.HandleModDecodeResult(resource, res, block, parseCtx)
	}

	return resource, res
}

func (d *DecoderImpl) getDecodeFunc(block *hcl.Block) (DecodeFunc, hcl.Diagnostics) {
	decodeFunc, ok := d.DecodeFuncs[block.Type]
	if !ok {
		if d.DefaultDecodeFunc != nil {
			return d.DefaultDecodeFunc, nil
		}
		return nil,
			hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("no decode function registered for block type %s", block.Type),
				Subject:  &block.DefRange,
			}}
	}
	return decodeFunc, nil
}

func (d *DecoderImpl) decodeLocals(block *hcl.Block, parseCtx *ModParseContext) ([]*modconfig.Local, *DecodeResult) {
	res := NewDecodeResult()
	attrs, diags := block.Body.JustAttributes()
	if len(attrs) == 0 {
		res.Diags = diags
		return nil, res
	}

	// build list of locals
	locals := make([]*modconfig.Local, 0, len(attrs))
	for name, attr := range attrs {
		if !hclsyntax.ValidIdentifier(name) {
			res.Diags = append(res.Diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid local value name",
				Detail:   badIdentifierDetail,
				Subject:  &attr.NameRange,
			})
			continue
		}
		// try to evaluate expression
		val, diags := attr.Expr.Value(parseCtx.EvalCtx)
		// handle any resulting diags, which may specify dependencies
		res.HandleDecodeDiags(diags)

		// add to our list
		locals = append(locals, modconfig.NewLocal(name, val, attr.Range, parseCtx.CurrentMod))
	}
	return locals, res
}

func (d *DecoderImpl) decodeVariable(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	res := NewDecodeResult()

	var variable *modconfig.Variable
	content, diags := block.Body.Content(VariableBlockSchema)
	res.HandleDecodeDiags(diags)

	v, diags := DecodeVariableBlock(block, content, parseCtx)
	res.HandleDecodeDiags(diags)

	if res.Success() {
		variable = modconfig.NewVariable(v, parseCtx.CurrentMod)
	} else {
		slog.Error("decodeVariable failed", "diags", res.Diags)
		return nil, res
	}
	// if a type property was specified, extract type string from the hcl source
	if attr, exists := content.Attributes[schema.AttributeTypeType]; exists {
		src := parseCtx.FileData[attr.Expr.Range().Filename]
		variable.TypeString = ExtractExpressionString(attr.Expr, src)
	}

	diags = DecodeProperty(content, "tags", &variable.Tags, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = DecodeProperty(content, "tags", &variable.Tags, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	return variable, res
}

func ExtractExpressionString(expr hcl.Expression, src []byte) string {
	rng := expr.Range()
	return string(src[rng.Start.Byte:rng.End.Byte])
}

// HandleModDecodeResult
// if decode was successful:
// - generate and set resource metadata
// - add resource to ModParseContext (which adds it to the mod)HandleModDecodeResult
func (d *DecoderImpl) HandleModDecodeResult(resource modconfig.HclResource, res *DecodeResult, block *hcl.Block, parseCtx *ModParseContext) {
	if !res.Success() {
		if len(res.Depends) > 0 {
			moreDiags := parseCtx.AddDependencies(block, resource.GetUnqualifiedName(), res.Depends)
			res.AddDiags(moreDiags)
		}
		return
	}
	// set whether this is a top level resource
	resource.SetTopLevel(parseCtx.IsTopLevelBlock(block))

	// call post decode hook
	// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
	moreDiags := resource.OnDecoded(block, parseCtx)
	res.AddDiags(moreDiags)

	// add references
	moreDiags = AddReferences(resource, block, parseCtx)
	res.AddDiags(moreDiags)

	if d.ValidateFunc != nil {
		res.AddDiags(d.ValidateFunc(resource))
	}

	// if we failed validation, return
	if !res.Success() {
		return
	}

	// if resource is NOT anonymous, and this is a TOP LEVEL BLOCK, add into the run context
	// NOTE: we can only reference resources defined in a top level block
	if !resourceIsAnonymous(resource) && resource.IsTopLevel() {
		moreDiags = parseCtx.AddResource(resource)
		res.AddDiags(moreDiags)
	}

	// if resource supports metadata, save it
	if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
		moreDiags = AddResourceMetadata(resourceWithMetadata, resource.GetHclResourceImpl().DeclRange, parseCtx)
		res.AddDiags(moreDiags)
	}
}

// ShouldAddToMod determines whether the resource should be added to the mod
// this may be overridden by the app specific decoder to add app-specific resource logic
func (d *DecoderImpl) ShouldAddToMod(resource modconfig.HclResource, block *hcl.Block, parseCtx *ModParseContext) bool {
	// do not add mods
	return resource.GetBlockType() != schema.BlockTypeMod
}
