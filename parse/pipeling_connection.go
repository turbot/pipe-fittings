package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func DecodeFlowpipeConnection(configPath string, block *hcl.Block) (modconfig.PipelingConnection, hcl.Diagnostics) {
	if len(block.Labels) != 2 {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid Flowpipe connection block - expected 2 labels, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	connectionType := block.Labels[0]

	conn, err := modconfig.NewPipelingConnection(block)
	if err != nil {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("error creating connection: %s", err),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	if helpers.IsNil(conn) {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid connection type '%s'", connectionType),
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

	diags = DecodeHclBody(body, evalCtx, nil, conn)
	if len(diags) > 0 {
		return nil, diags
	}

	moreDiags := conn.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	return conn, diags
}
