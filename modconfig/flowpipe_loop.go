package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type LoopDefn interface {
	GetType() string
	UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error)
	SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics
	Equals(LoopDefn) bool
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
	case schema.BlockTypePipelineStepContainer:
		return &LoopContainerStep{}
	case schema.BlockTypePipelineStepInput:
		return &LoopInputStep{}
	}

	return nil
}

type LoopSleepStep struct {
	UnresolvedAttributes hcl.Attributes `json:"-"`
	Until                bool           `json:"until"`
	Duration             *string        `json:"duration,omitempty"`
}

func (l *LoopSleepStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopSleepStep, ok := other.(*LoopSleepStep)
	if !ok {
		return false
	}

	return l.Until == otherLoopSleepStep.Until &&
		utils.PtrEqual(l.Duration, otherLoopSleepStep.Duration)
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

func (s *LoopSleepStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// for name, attr := range hclAttributes {
	// 	switch name {
	// 	case schema.AttributeTypeDuration:
	// 		s.Duration = hclhelpers.StringPtr(attr)
	// 	}
	// }

	return diags
}

type LoopTransformStep struct {
	Until bool        `json:"until" hcl:"until" cty:"until"`
	Value interface{} `json:"value,omitempty" hcl:"value,optional" cty:"value"`
}

func (l *LoopTransformStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopTransformStep, ok := other.(*LoopTransformStep)
	if !ok {
		return false
	}

	return l.Until == otherLoopTransformStep.Until &&
		reflect.DeepEqual(l.Value, otherLoopTransformStep.Value)
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
				return nil, error_helpers.BetterHclDiagsToError("transform loop", diags)
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

func (s *LoopTransformStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}
