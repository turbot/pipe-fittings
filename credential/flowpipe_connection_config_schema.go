package credential

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
)

func ParseConfig(configString []byte, filename string, startPos hcl.Pos) (hcl.Body, hcl.Diagnostics) {
	file, diags := hclsyntax.ParseConfig(configString, filename, startPos)
	if diags.HasErrors() {
		// try json
		return ParseJsonConfig(configString, filename)

	}

	return file.Body, nil
}

func ParseJsonConfig(configString []byte, filename string) (hcl.Body, hcl.Diagnostics) {
	file, diags := json.Parse(configString, filename)
	if diags.HasErrors() {
		return nil, diags
	}
	return file.Body, nil
}
