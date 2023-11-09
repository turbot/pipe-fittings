package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/options"
)

// DecodeOptions decodes an options block
func DecodeOptions(block *hcl.Block, blockFactory OptionsBlockFactory) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	destination, diags := blockFactory(block)
	if diags.HasErrors() {
		return nil, diags
	}

	diags = gohcl.DecodeBody(block.Body, nil, destination)
	if diags.HasErrors() {
		return nil, diags
	}

	return destination, nil
}
