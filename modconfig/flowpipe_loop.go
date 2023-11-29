package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type LoopDefn interface {
	UntilReached() bool
	GetType() string
	UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error)
}

func GetLoopDefn(stepType string) LoopDefn {
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		return &LoopHttpStep{}
	case schema.BlockTypePipelineStepSleep:
		return &LoopSleepStep{}
	case schema.BlockTypePipelineStepQuery:
		return &LoopQueryStep{}
	case schema.BlockTypePipelineStepPipeline:
		return &LoopPipelineStep{}
	case schema.BlockTypePipelineStepTransform:
		return &LoopTransformStep{}
	}

	return nil
}

type LoopEmailStep struct {
	Until            bool      `json:"until" hcl:"until" cty:"until"`
	To               *[]string `json:"to,omitempty" hcl:"to,optional" cty:"to"`
	From             *string   `json:"from,omitempty" hcl:"from,optional" cty:"from"`
	SenderCredential *string   `json:"sender_credential,omitempty" hcl:"sender_credential,optional" cty:"sender_credential"`
	Host             *string   `json:"host,omitempty" hcl:"host,optional" cty:"host"`
	Port             *int64    `json:"port,omitempty" hcl:"port,optional" cty:"port"`
	SenderName       *string   `json:"sender_name,omitempty" hcl:"sender_name,optional" cty:"sender_name"`
	Cc               *[]string `json:"cc,omitempty" hcl:"cc,optional" cty:"cc"`
	Bcc              *[]string `json:"bcc,omitempty" hcl:"bcc,optional" cty:"bcc"`
	Body             *string   `json:"body,omitempty" hcl:"body,optional" cty:"body"`
	ContentType      *string   `json:"content_type,omitempty" hcl:"content_type,optional" cty:"content_type"`
	Subject          *string   `json:"subject,omitempty" hcl:"subject,optional" cty:"subject"`
}

func (l *LoopEmailStep) ShouldRun() bool {
	return l.Until
}

func (l *LoopEmailStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	if l.To != nil {
		input["to"] = *l.To
	}
	if l.From != nil {
		input["from"] = *l.From
	}
	if l.SenderCredential != nil {
		input["sender_credential"] = *l.SenderCredential
	}
	if l.Host != nil {
		input["host"] = *l.Host
	}
	if l.Port != nil {
		input["port"] = *l.Port
	}
	if l.SenderName != nil {
		input["sender_name"] = *l.SenderName
	}
	if l.Cc != nil {
		input["cc"] = *l.Cc
	}
	if l.Bcc != nil {
		input["bcc"] = *l.Bcc
	}
	if l.Body != nil {
		input["body"] = *l.Body
	}
	if l.ContentType != nil {
		input["content_type"] = *l.ContentType
	}
	if l.Subject != nil {
		input["subject"] = *l.Subject
	}
	return input, nil
}

func (*LoopEmailStep) GetType() string {
	return schema.BlockTypePipelineStepEmail
}

type LoopQueryStep struct {
	Until             bool           `json:"until" hcl:"until" cty:"until"`
	ConnnectionString *string        `json:"connection_string,omitempty" hcl:"connection_string,optional" cty:"connection_string"`
	Sql               *string        `json:"sql,omitempty" hcl:"sql,optional" cty:"sql"`
	Args              *[]interface{} `json:"args,omitempty" hcl:"args,optional" cty:"args"`
}

func (l *LoopQueryStep) UntilReached() bool {
	return l.Until
}

func (l *LoopQueryStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	if l.ConnnectionString != nil {
		input["connection_string"] = *l.ConnnectionString
	}
	if l.Sql != nil {
		input["sql"] = *l.Sql
	}
	if l.Args != nil {
		input["args"] = *l.Args
	}
	return input, nil
}

func (*LoopQueryStep) GetType() string {
	return schema.BlockTypePipelineStepQuery
}

type LoopHttpStep struct {
	Until            bool                    `json:"until" hcl:"until" cty:"until"`
	URL              *string                 `json:"url,omitempty" hcl:"url,optional" cty:"url"`
	Method           *string                 `json:"method,omitempty" hcl:"method,optional" cty:"method"`
	RequestBody      *string                 `json:"request_body,omitempty" hcl:"request_body,optional" cty:"request_body"`
	RequestHeaders   *map[string]interface{} `json:"request_headers,omitempty" hcl:"request_headers,optional" cty:"request_headers"`
	RequestTimeoutMs *int                    `json:"request_timeout_ms,omitempty" hcl:"request_timeout_ms,optional" cty:"request_timeout_ms"`
	CaCertPem        *string                 `json:"ca_cert_pem,omitempty" hcl:"ca_cert_pem,optional" cty:"ca_cert_pem"`
	Insecure         *bool                   `json:"insecure,omitempty" hcl:"insecure,optional" cty:"insecure"`
}

func (l *LoopHttpStep) UntilReached() bool {
	return l.Until
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
	if l.RequestTimeoutMs != nil {
		input["request_timeout_ms"] = *l.RequestTimeoutMs
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

type LoopSleepStep struct {
	Until    bool    `json:"until" hcl:"until" cty:"until"`
	Duration *string `json:"duration,omitempty" hcl:"duration,optional" cty:"duration"`
}

func (l *LoopSleepStep) UntilReached() bool {
	return l.Until
}

func (l *LoopSleepStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	if l.Duration != nil {
		input["duration"] = *l.Duration
	}
	return input, nil
}

func (*LoopSleepStep) GetType() string {
	return schema.BlockTypePipelineStepSleep
}

type LoopPipelineStep struct {
	Until bool `json:"until" hcl:"until" cty:"until"`
}

func (l *LoopPipelineStep) UntilReached() bool {
	return l.Until
}

func (l *LoopPipelineStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	return input, nil
}

func (*LoopPipelineStep) GetType() string {
	return schema.BlockTypePipelineStepPipeline
}

type LoopTransformStep struct {
	Until bool        `json:"until" hcl:"until" cty:"until"`
	Value interface{} `json:"value,omitempty" hcl:"value,optional" cty:"value"`
}

func (l *LoopTransformStep) UntilReached() bool {
	return l.Until
}

func (l *LoopTransformStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	expr, ok := l.Value.(hcl.Expression)
	if ok {
		val, err := expr.Value(nil)
		if err != nil {
			return nil, err
		}

		if !val.IsNull() {
			goVal, err := hclhelpers.CtyToGo(val)
			if err != nil {
				return nil, err
			}
			input["value"] = goVal
		}
	} else {
		hclAttrib, ok := l.Value.(*hcl.Attribute)
		if !ok {
			input["value"] = l.Value
		} else {
			var ctyValue cty.Value
			diags := gohcl.DecodeExpression(hclAttrib.Expr, evalContext, &ctyValue)
			if len(diags) > 0 {
				return nil, error_helpers.HclDiagsToError("transform loop", diags)
			}
			goVal, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				return nil, err
			}
			input["value"] = goVal
		}

	}

	return input, nil
}

func (*LoopTransformStep) GetType() string {
	return schema.BlockTypePipelineStepTransform
}
