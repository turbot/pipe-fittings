package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/zclconf/go-cty/cty"
)

func DecodePipelingConnection(configPath string, block *hcl.Block) (connection.PipelingConnection, hcl.Diagnostics) {
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

	conn, err := app_specific_connection.NewPipelingConnection(block)
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
				Summary:  fmt.Sprintf("invalid connection type '%s'", block.Labels[0]),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: funcs.ContextFunctions(configPath),
		Variables: make(map[string]cty.Value),
	}

	// Decode the body of the block into the Go struct
	diags := gohcl.DecodeBody(block.Body, evalCtx, conn)
	if diags.HasErrors() {
		return nil, diags
	}

	moreDiags := conn.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	return conn, diags
}
