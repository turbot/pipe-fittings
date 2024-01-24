package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func DecodeCredentialImport(configPath string, block *hcl.Block) (*modconfig.CredentialImport, hcl.Diagnostics) {

	if len(block.Labels) != 1 {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid credential_import block - expected 1 label, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	credentialImportName := block.Labels[0]

	credentialImport := modconfig.NewCredentialImport(block)
	if credentialImport == nil {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid credential_import '%s'", credentialImportName),
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

	diags = decodeHclBody(body, evalCtx, nil, credentialImport)
	if len(diags) > 0 {
		return nil, diags
	}

	// moreDiags := credential.Validate()
	// if len(moreDiags) > 0 {
	// 	diags = append(diags, moreDiags...)
	// }

	return credentialImport, diags
}
