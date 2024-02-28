package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepMessage struct {
	PipelineStepBase

	Text string `json:"text" hcl:"text" cty:"text"`

	// Notifier cty.Value `json:"-" cty:"notify"`
	Notifier NotifierImpl `json:"notify" cty:"-"`

	// overrides
	Cc      []string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc     []string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel *string  `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Subject *string  `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	To      []string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
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

	return p.Text == other.Text &&
		utils.PtrEqual(p.Subject, other.Subject) &&
		helpers.StringSliceEqualIgnoreOrder(p.Cc, other.Cc) &&
		helpers.StringSliceEqualIgnoreOrder(p.Bcc, other.Bcc) &&
		utils.PtrEqual(p.Channel, other.Channel) &&
		helpers.StringSliceEqualIgnoreOrder(p.To, other.To) &&
		p.Notifier.Equals(&other.Notifier)
}

func (p *PipelineStepMessage) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	results := map[string]interface{}{}

	// text is a mandatory attribute
	var text string
	if p.UnresolvedAttributes[schema.AttributeTypeText] == nil {
		text = p.Text
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeText], evalContext, &text)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}
	results[schema.AttributeTypeText] = text

	// subject
	var subject *string
	if p.UnresolvedAttributes[schema.AttributeTypeSubject] == nil {
		subject = p.Subject
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSubject], evalContext, &subject)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	if subject != nil {
		results[schema.AttributeTypeSubject] = *subject
	}

	// notifier
	if attr, ok := p.UnresolvedAttributes[schema.AttributeTypeNotifier]; !ok {
		results[schema.AttributeTypeNotifier] = p.Notifier
	} else {
		notifierCtyVal, moreDiags := attr.Value(evalContext)
		if moreDiags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, moreDiags)
		}

		notifier, err := ctyValueToPipelineStepNotifierValueMap(notifierCtyVal)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse notifier attribute: " + err.Error())
		}
		results[schema.AttributeTypeNotifier] = notifier
	}

	return results, nil

}

func (p *PipelineStepMessage) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {

		case schema.AttributeTypeText:
			stepDiags := setStringAttribute(attr, evalContext, p, "Text", false)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeChannel, schema.AttributeTypeSubject:

			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, structFieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeCc, schema.AttributeTypeBcc, schema.AttributeTypeTo:
			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringSliceAttribute(attr, evalContext, p, structFieldName, false)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeNotifier:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				var err error
				p.Notifier, err = ctyValueToPipelineStepNotifierValueMap(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeNotifier + " attribute to InputNotifier",
						Detail:   err.Error(),
						Subject:  &attr.Range,
					})
				}
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Message Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}
