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
	AppendDependsOn(...string)
	AppendCredentialDependsOn(...string)
	AddUnresolvedAttribute(string, hcl.Expression)
	GetUnresolvedAttributes() map[string]hcl.Expression
}

func GetLoopDefn(stepType string, p *PipelineStepBase) LoopDefn {
	loopStep := LoopStep{
		PipelineStepBase:     p,
		UnresolvedAttributes: map[string]hcl.Expression{},
	}

	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		return &LoopHttpStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepSleep:
		return &LoopSleepStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepQuery:
		return &LoopQueryStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepPipeline:
		return &LoopPipelineStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepTransform:
		return &LoopTransformStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepContainer:
		return &LoopContainerStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepInput:
		return &LoopInputStep{
			LoopStep: loopStep,
		}
	}

	return nil
}

type LoopStep struct {
	// circular link to its "parent"
	PipelineStepBase *PipelineStepBase

	UnresolvedAttributes map[string]hcl.Expression
	Until                *bool
}

func (l *LoopStep) GetUnresolvedAttributes() map[string]hcl.Expression {
	return l.UnresolvedAttributes
}

func (l *LoopStep) AppendDependsOn(dependsOn ...string) {
	l.PipelineStepBase.AppendDependsOn(dependsOn...)
}

func (*LoopStep) AppendCredentialDependsOn(...string) {
	// not implemented
}

func (l *LoopStep) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	l.UnresolvedAttributes[name] = expr
}

func (s *LoopStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	if attr, ok := hclAttributes[schema.AttributeTypeUntil]; ok {
		stepDiags := setBoolAttribute(attr, evalContext, s, "Until", true)
		if stepDiags.HasErrors() {
			diags = append(diags, stepDiags...)
		}
	}

	return diags
}

type LoopSleepStep struct {
	LoopStep

	UnresolvedAttributes hcl.Attributes `json:"-"`
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

	simpleTypeInputFromAttribute(l.LoopStep, input, evalContext, schema.AttributeTypeDuration, l.Duration)
	return input, nil
}

func (*LoopSleepStep) GetType() string {
	return schema.BlockTypePipelineStepSleep
}

func (s *LoopSleepStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := s.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDuration:
			stepDiags := setStringAttribute(attr, evalContext, s, "Duration", true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeUntil:
			// already handled in SetAttributes
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   "Invalid attribute " + name + " for loop sleep step",
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

type LoopTransformStep struct {
	LoopStep

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
