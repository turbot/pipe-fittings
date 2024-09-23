package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

	// create an empty connection struct of appropriate type
	conn, err := app_specific_connection.NewPipelingConnection(block.Labels[0], block.Labels[1], block.DefRange)
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
	// if the connection hcl defines a 'pipes' block, decode this separately
	// and remove this block from the hcl body before decoding the rest of the connection
	body, pipes, diags := extractPipesBlock(block.Body, evalCtx)
	if diags.HasErrors() {
		return nil, diags
	}
	// if a pipes block was decoded, set it on the connection
	if pipes != nil {
		conn.SetPipesMetadata(pipes)
	}

	diags = gohcl.DecodeBody(body, evalCtx, conn)
	if diags.HasErrors() {
		return nil, diags
	}

	moreDiags := conn.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	return conn, diags
}

// if the connection hcl defines a 'pipes' block, decode this separately and remove this block from the body
// before decoding the rest of the connection
func extractPipesBlock(body hcl.Body, evalCtx *hcl.EvalContext) (hcl.Body, *connection.PipesConnectionMetadata, hcl.Diagnostics) {
	syntaxBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return body, nil, nil
	}
	var pipes *connection.PipesConnectionMetadata
	// build a new body without the 'pipes' block
	var updatedBlocks []*hclsyntax.Block
	for _, childBlock := range syntaxBody.Blocks {
		if childBlock.Type == "pipes" {
			pipes = &connection.PipesConnectionMetadata{}
			diags := gohcl.DecodeBody(childBlock.Body, evalCtx, pipes)
			if diags.HasErrors() {
				return nil, nil, diags
			}
		} else {
			updatedBlocks = append(updatedBlocks, childBlock)
		}
	}

	syntaxBody.Blocks = updatedBlocks
	return syntaxBody, pipes, nil
}
