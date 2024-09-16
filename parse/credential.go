package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/zclconf/go-cty/cty"
)

func DecodeCredential(configPath string, block *hcl.Block) (credential.Credential, hcl.Diagnostics) {

	if len(block.Labels) != 2 {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid credential block - expected 2 labels, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	credentialType := block.Labels[0]

	cred, err := credential.NewCredential(block)
	if err != nil {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("error creating credential: %s", err),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	if helpers.IsNil(cred) {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid credential type '%s'", credentialType),
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

	diags = DecodeHclBody(body, evalCtx, nil, cred)
	if len(diags) > 0 {
		return nil, diags
	}

	moreDiags := cred.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	return cred, diags
}
