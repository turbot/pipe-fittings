package parse

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

// DecodeResult holds the result of a decoding operation - diags and resource dependencies
type DecodeResult struct {
	Diags   hcl.Diagnostics
	Depends map[string]*modconfig.ResourceDependency
}

func NewDecodeResult() *DecodeResult {
	return &DecodeResult{Depends: make(map[string]*modconfig.ResourceDependency)}
}

// Merge merges this decode result with another
func (p *DecodeResult) Merge(other *DecodeResult) *DecodeResult {
	p.Diags = append(p.Diags, other.Diags...)
	for k, v := range other.Depends {
		p.Depends[k] = v
	}

	return p
}

// Success returns if the was parsing successful - true if there are no errors and no dependencies
func (p *DecodeResult) Success() bool {
	return !p.Diags.HasErrors() && len(p.Depends) == 0
}

// HandleDecodeDiags adds dependencies to the result if the diags contains dependency errors,
// otherwise adds diags to the result
func (p *DecodeResult) HandleDecodeDiags(diags hcl.Diagnostics) {
	for _, diag := range diags {
		if dependency := DiagsToDependency(diag); dependency != nil {
			p.Depends[dependency.String()] = dependency
		}
	}
	// only register errors if there are NOT any missing variables
	if len(p.Depends) == 0 {
		p.AddDiags(diags)
	}
}

// DiagsToDependency determines whether the diag is a dependency error, and if so, return a dependency object
func DiagsToDependency(diag *hcl.Diagnostic) *modconfig.ResourceDependency {
	if helpers.StringSliceContains(missingVariableErrors, diag.Summary) {
		return &modconfig.ResourceDependency{Range: diag.Expression.Range(), Traversals: diag.Expression.Variables()}
	}
	return nil
}

func (p *DecodeResult) AddDiags(diags hcl.Diagnostics) {
	p.Diags = append(p.Diags, diags...)
}

func (p *DecodeResult) VariableDependenciesToWarnings() []string {
	var res []string
	for _, dep := range p.Depends {
		for _, t := range dep.Traversals {
			traversalString := hclhelpers.TraversalAsString(t)
			if strings.HasPrefix(traversalString, "var.") {
				res = append(res, constants.MissingVariableWarning+traversalString)
			}
		}
	}
	return res
}
