package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type PipelineStepMessage struct {
	PipelineStepBase

	Body     string `json:"body" hcl:"body" cty:"body"`
	Markdown *bool  `json:"markdown" hcl:"markdown,optional" cty:"markdown"`

	// Notifier cty.Value `json:"-" cty:"notify"`
	Notifier NotifierImpl `json:"notify" cty:"-"`
}

func (p *PipelineStepMessage) Equals(iOther PipelineStep) bool {
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || !helpers.IsNil(iOther) && p == nil {
		return false
	}

	other, ok := iOther.(*PipelineStepMessage)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	return p.Body == other.Body &&
		utils.BoolPtrEqual(p.Markdown, other.Markdown) &&
		p.Notifier.Equals(&other.Notifier)
}

func (p *PipelineStepMessage) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var value any

	return map[string]interface{}{
		schema.AttributeTypeValue: value,
	}, nil
}

func (p *PipelineStepMessage) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Transform Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}
