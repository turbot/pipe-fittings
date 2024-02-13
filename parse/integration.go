package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func DecodeIntegration(configPath string, block *hcl.Block) (modconfig.Integration, hcl.Diagnostics) {
	if len(block.Labels) != 2 {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid integration block - expected 2 labels, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	integrationType := block.Labels[0]

	integration := modconfig.NewIntegrationFromBlock(block)
	if integration == nil {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid integration type '%s'", integrationType),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	if len(diags) > 0 {
		return nil, diags
	}

	body := r.(*hclsyntax.Body)

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: funcs.ContextFunctions(configPath),
		Variables: make(map[string]cty.Value),
	}

	diags = decodeHclBody(body, evalCtx, nil, integration)
	if len(diags) > 0 {
		return nil, diags
	}

	moreDiags := integration.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	return integration, diags
}