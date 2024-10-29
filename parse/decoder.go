package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"log/slog"
)

type DecoderOption func(Decoder)

type Decoder interface {
	Decode(*ModParseContext) hcl.Diagnostics
}

type DecodeFunc func(*hcl.Block, *ModParseContext) (modconfig.HclResource, *DecodeResult)

type DecoderImpl struct {
	// registered block types
	DecodeFuncs map[string]DecodeFunc
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
		if block.Type == schema.BlockTypeLocals {
			// TODO K remove spceial casing - decodeBVlock could return an array
			resources, res := d.decodeLocalsBlock(block, parseCtx)
			if !res.Success() {
				diags = append(diags, res.Diags...)
				continue
			}
			for _, resource := range resources {
				resourceDiags := AddResourceToMod(resource, block, parseCtx)
				diags = append(diags, resourceDiags...)
			}
		} else {
			resource, res := d.DecodeBlock(block, parseCtx)
			diags = append(diags, res.Diags...)
			if !res.Success() || resource == nil {
				continue
			}

			resourceDiags := AddResourceToMod(resource, block, parseCtx)
			diags = append(diags, resourceDiags...)
		}
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
		HandleModDecodeResult(local, res, block, parseCtx)
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

	decodeFunc := d.getDecodeFunc(block.Type)
	resource, res = decodeFunc(block, parseCtx)
	// Note that an interface value that holds a nil concrete value is itself non-nil.
	if !helpers.IsNil(resource) {
		// handle the result
		// - if there are dependencies, add to run context
		HandleModDecodeResult(resource, res, block, parseCtx)
	}

	return resource, res
}

func (d *DecoderImpl) getDecodeFunc(t string) DecodeFunc {
	decodeFunc, ok := d.DecodeFuncs[t]
	if !ok {
		// default to generic decode function
		decodeFunc = d.decodeResource
	}
	return decodeFunc
}

// generic decode function for any resource we do not have custom decode logic for
func (d *DecoderImpl) decodeResource(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	res := NewDecodeResult()
	// get shell resource
	resource, diags := ResourceForBlock(block, parseCtx)
	res.HandleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	diags = DecodeHclBody(block.Body, parseCtx.EvalCtx, parseCtx, resource)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
	}
	return resource, res
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
