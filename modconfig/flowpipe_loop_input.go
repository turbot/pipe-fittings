package modconfig

import (
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type LoopInputStep struct {
	Until bool `json:"until" hcl:"until" cty:"until"`

	Notifier *cty.Value `json:"notifier" cty:"-" hcl:"-"`

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Prompt  *string   `json:"prompt" cty:"prompt" hcl:"prompt,optional"`
	Cc      *[]string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc     *[]string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel *string   `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Subject *string   `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	To      *[]string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
}

func (s *LoopInputStep) Equals(other LoopDefn) bool {

	if s == nil && helpers.IsNil(other) {
		return true
	}

	if s == nil && !helpers.IsNil(other) || s != nil && helpers.IsNil(other) {
		return false
	}

	otherLoopInputStep, ok := other.(*LoopInputStep)
	if !ok {
		return false
	}

	if s.Cc == nil && otherLoopInputStep.Cc != nil || s.Cc != nil && otherLoopInputStep.Cc == nil {
		return false
	} else if s.Cc != nil {
		if slices.Compare(*s.Cc, *otherLoopInputStep.Cc) != 0 {
			return false
		}
	}

	if s.Bcc == nil && otherLoopInputStep.Bcc != nil || s.Bcc != nil && otherLoopInputStep.Bcc == nil {
		return false
	} else if s.Bcc != nil {
		if slices.Compare(*s.Bcc, *otherLoopInputStep.Bcc) != 0 {
			return false
		}
	}

	if s.To == nil && otherLoopInputStep.To != nil || s.To != nil && otherLoopInputStep.To == nil {
		return false
	} else if s.To != nil {
		if slices.Compare(*s.To, *otherLoopInputStep.To) != 0 {
			return false
		}
	}

	return utils.PtrEqual(s.Prompt, otherLoopInputStep.Prompt) &&
		utils.PtrEqual(s.Channel, otherLoopInputStep.Channel) &&
		utils.PtrEqual(s.Subject, otherLoopInputStep.Subject)
}

func (s *LoopInputStep) GetType() string {
	return schema.BlockTypeInput
}

func (s *LoopInputStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	if s.Prompt != nil {
		input[schema.AttributeTypePrompt] = *s.Prompt
	}

	if s.Cc != nil {
		input[schema.AttributeTypeCc] = *s.Cc
	}

	if s.Bcc != nil {
		input[schema.AttributeTypeBcc] = *s.Bcc
	}

	if s.Channel != nil {
		input[schema.AttributeTypeChannel] = *s.Channel
	}

	if s.Subject != nil {
		input[schema.AttributeTypeSubject] = *s.Subject
	}

	if s.To != nil {
		input[schema.AttributeTypeTo] = *s.To
	}

	return input, nil
}

func (s *LoopInputStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}
