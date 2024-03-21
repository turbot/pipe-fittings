package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type LoopPipelineStep struct {
	LoopStep

	Args interface{} `json:"args,omitempty" hcl:"args,optional" cty:"args"`
}

func (l *LoopPipelineStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopPipelineStep, ok := other.(*LoopPipelineStep)
	if !ok {
		return false
	}

	return l.Until == otherLoopPipelineStep.Until &&
		reflect.DeepEqual(l.Args, otherLoopPipelineStep.Args)
}

func (l *LoopPipelineStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	expr, ok := l.Args.(hcl.Expression)
	if ok {
		val, err := expr.Value(nil)
		if err != nil {
			return nil, err
		}

		if !val.IsNull() {
			goVal, err := hclhelpers.CtyToGoMapInterface(val)
			if err != nil {
				return nil, err
			}
			input["args"] = goVal
		}
	} else {
		hclAttr, ok := l.Args.(*hcl.Attribute)
		if !ok {
			input["args"] = l.Args
		} else {
			var ctyValue cty.Value
			diags := gohcl.DecodeExpression(hclAttr.Expr, evalContext, &ctyValue)
			if len(diags) > 0 {
				return nil, error_helpers.BetterHclDiagsToError("pipeline loop", diags)
			}
			goVal, err := hclhelpers.CtyToGoMapInterface(ctyValue)
			if err != nil {
				return nil, err
			}
			input["args"] = goVal
		}
	}

	return input, nil
}

func (*LoopPipelineStep) GetType() string {
	return schema.BlockTypePipelineStepPipeline
}

func (s *LoopPipelineStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}
