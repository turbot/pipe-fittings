package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func NewThrowConfig(p *PipelineStepBase) *ThrowConfig {
	return &ThrowConfig{
		PipelineStepBase:     p,
		UnresolvedAttributes: make(map[string]hcl.Expression),
	}
}

type ThrowConfig struct {
	// Circular reference to its parent
	PipelineStepBase     *PipelineStepBase
	UnresolvedAttributes map[string]hcl.Expression

	If      *bool
	Message *string
}

func (t *ThrowConfig) AppendDependsOn(dependsOn ...string) {
	t.PipelineStepBase.AppendDependsOn(dependsOn...)
}

func (t *ThrowConfig) AppendCredentialDependsOn(...string) {
	// not implemented
}

func (t *ThrowConfig) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	t.UnresolvedAttributes[name] = expr
}

func (t *ThrowConfig) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeIf:
			t.AddUnresolvedAttribute(name, attr.Expr)
		case schema.AttributeTypeMessage:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, t)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				valString, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMessage + " attribute to string",
						Subject:  &attr.Range,
					})
					continue
				}

				t.Message = &valString
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   "Unsupported attribute '" + name + "' in throw block",
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

func (t *ThrowConfig) Resolve() *ThrowConfig {

	return t
}
