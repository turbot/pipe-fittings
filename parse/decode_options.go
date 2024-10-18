package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
)

// DecodeOptions decodes an options block
func DecodeOptions(block *hcl.Block, blockFactory modconfig.OptionsBlockFactory) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	destination, diags := blockFactory(block)
	if diags.HasErrors() {
		return nil, diags
	}

	if timingOptions, ok := destination.(options.CanSetTiming); ok {
		morediags := decodeTimingFlag(block, timingOptions)
		if morediags.HasErrors() {
			diags = append(diags, morediags...)
			return nil, diags
		}
	}

	diags = gohcl.DecodeBody(block.Body, nil, destination)
	if diags.HasErrors() {
		return nil, diags
	}

	return destination, nil
}

// for Query options block,  if timing attribute is set to "verbose", replace with true and set verbose to true
func decodeTimingFlag(block *hcl.Block, timingOptions options.CanSetTiming) hcl.Diagnostics {
	body := block.Body.(*hclsyntax.Body)
	timingAttribute := body.Attributes["timing"]
	if timingAttribute == nil {
		return nil
	}
	// remove the attribute so subsequent decoding does not see it
	delete(body.Attributes, "timing")

	val, diags := timingAttribute.Expr.Value(&hcl.EvalContext{
		Variables: map[string]cty.Value{
			constants.ArgOn:      cty.StringVal(constants.ArgOn),
			constants.ArgOff:     cty.StringVal(constants.ArgOff),
			constants.ArgVerbose: cty.StringVal(constants.ArgVerbose),
		},
	})
	if diags.HasErrors() {
		return diags
	}
	// support legacy boolean values
	if val == cty.True {
		val = cty.StringVal(constants.ArgOn)
	}
	if val == cty.False {
		val = cty.StringVal(constants.ArgOff)
	}
	return timingOptions.SetTiming(val.AsString(), timingAttribute.Range())

}
