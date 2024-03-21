package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type LoopHttpStep struct {
	LoopStep

	URL            *string                 `json:"url,omitempty" hcl:"url,optional" cty:"url"`
	Method         *string                 `json:"method,omitempty" hcl:"method,optional" cty:"method"`
	RequestBody    *string                 `json:"request_body,omitempty" hcl:"request_body,optional" cty:"request_body"`
	RequestHeaders *map[string]interface{} `json:"request_headers,omitempty" hcl:"request_headers,optional" cty:"request_headers"`
	CaCertPem      *string                 `json:"ca_cert_pem,omitempty" hcl:"ca_cert_pem,optional" cty:"ca_cert_pem"`
	Insecure       *bool                   `json:"insecure,omitempty" hcl:"insecure,optional" cty:"insecure"`
}

func (l *LoopHttpStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopHttpStep, ok := other.(*LoopHttpStep)
	if !ok {
		return false
	}

	if l.RequestHeaders == nil && otherLoopHttpStep.RequestHeaders != nil || l.RequestHeaders != nil && otherLoopHttpStep.RequestHeaders == nil {
		return false
	}

	if l.RequestHeaders != nil {
		if !reflect.DeepEqual(*l.RequestHeaders, *otherLoopHttpStep.RequestHeaders) {
			return false
		}
	}

	return l.Until == otherLoopHttpStep.Until &&
		utils.PtrEqual(l.URL, otherLoopHttpStep.URL) &&
		utils.PtrEqual(l.Method, otherLoopHttpStep.Method) &&
		utils.PtrEqual(l.RequestBody, otherLoopHttpStep.RequestBody) &&
		utils.PtrEqual(l.CaCertPem, otherLoopHttpStep.CaCertPem) &&
		utils.BoolPtrEqual(l.Insecure, otherLoopHttpStep.Insecure)
}

func (l *LoopHttpStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	if l.URL != nil {
		input["url"] = *l.URL
	}
	if l.Method != nil {
		input["method"] = *l.Method
	}
	if l.RequestBody != nil {
		input["request_body"] = *l.RequestBody
	}
	if l.RequestHeaders != nil {
		input["request_headers"] = *l.RequestHeaders
	}
	if l.CaCertPem != nil {
		input["ca_cert_pem"] = *l.CaCertPem
	}
	if l.Insecure != nil {
		input["insecure"] = *l.Insecure
	}

	return input, nil
}

func (*LoopHttpStep) GetType() string {
	return schema.BlockTypePipelineStepHttp
}

func (l *LoopHttpStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}
