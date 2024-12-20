package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

func DecodePipelingConnection(configPath string, block *hcl.Block) (connection.PipelingConnection, hcl.Diagnostics) {
	if len(block.Labels) != 2 {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid connection block - expected 2 labels, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	utils.LogTime(fmt.Sprintf("decode connection %v start", block.Labels))

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

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: funcs.ContextFunctions(configPath),
		Variables: make(map[string]cty.Value),
	}

	// Decode the connectionImpl - we pass the connectionImpl from con into the decodeConnectionImpl function
	// and it is mutated in place
	remainderBody, diags := decodeConnectionImpl(block, evalCtx, conn.GetConnectionImpl())
	if diags.HasErrors() {
		return nil, diags
	}
	// now decode the rest of the block
	diags = gohcl.DecodeBody(remainderBody, evalCtx, conn)
	if diags.HasErrors() {
		return nil, diags
	}

	// validate the connection
	moreDiags := conn.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	utils.LogTime(fmt.Sprintf("decode connection %v end", block.Labels))

	return conn, diags
}

// decodeConnectionImpl decodes the given block into a connection.ConnectionImpl and returns the remaining body.
func decodeConnectionImpl(block *hcl.Block, evalCtx *hcl.EvalContext, connectionImpl *connection.ConnectionImpl) (hcl.Body, hcl.Diagnostics) {
	
	utils.LogTime(fmt.Sprintf("decode connectionImpl %s start", connectionImpl.FullName))

	schema, err := hclhelpers.HclSchemaForStruct(connectionImpl)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("error getting schema for struct: %s", err),
				Subject:  hclhelpers.BlockRangePointer(block),
			},
		}
	}

	// Perform a partial content decode to split known fields and remaining fields
	bodyContent, remain, diags := block.Body.PartialContent(schema)
	if diags.HasErrors() {
		return nil, diags
	}

	// Now decode - we only decode the pipes block and tty field

	for name, attr := range bodyContent.Attributes {
		// we have applied the schema to the hcl so we should not get any unexpected attributes
		if name == "ttl" {
			moreDiags := gohcl.DecodeExpression(attr.Expr, evalCtx, &connectionImpl.Ttl)
			diags = append(diags, moreDiags...)
		}
	}

	// Decode the pipes block
	for _, childBlock := range bodyContent.Blocks {
		// we have applied the schema to the hcl so we should not get any unexpected blocks
		if childBlock.Type == "pipes" {
			pipes := &connection.PipesConnectionMetadata{}

			moreDiags := gohcl.DecodeBody(childBlock.Body, evalCtx, pipes)
			diags = append(diags, moreDiags...)
			if !diags.HasErrors() {
				connectionImpl.Pipes = pipes
			}
		}
	}

	utils.LogTime(fmt.Sprintf("decode connectionImpl %s end", connectionImpl.FullName))

	// Return the decoded result, remaining body, and any diagnostics
	return remain, diags
}
