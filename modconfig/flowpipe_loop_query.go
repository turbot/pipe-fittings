package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type LoopQueryStep struct {
	LoopStep

	ConnnectionString *string        `json:"connection_string,omitempty" hcl:"connection_string,optional" cty:"connection_string"`
	Sql               *string        `json:"sql,omitempty" hcl:"sql,optional" cty:"sql"`
	Args              *[]interface{} `json:"args,omitempty" hcl:"args,optional" cty:"args"`
}

func (l *LoopQueryStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopQueryStep, ok := other.(*LoopQueryStep)
	if !ok {
		return false
	}

	if l.Args == nil && otherLoopQueryStep.Args != nil || l.Args != nil && otherLoopQueryStep.Args == nil {
		return false
	}

	if l.Args != nil {
		if !reflect.DeepEqual(*l.Args, *otherLoopQueryStep.Args) {
			return false
		}
	}

	return l.Until == otherLoopQueryStep.Until &&
		utils.PtrEqual(l.ConnnectionString, otherLoopQueryStep.ConnnectionString) &&
		utils.PtrEqual(l.Sql, otherLoopQueryStep.Sql)
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

func (s *LoopQueryStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}
