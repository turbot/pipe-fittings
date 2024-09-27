package modconfig

import (
	"github.com/hashicorp/hcl/v2"
)

type EvalContext struct {
	*hcl.EvalContext
	LateBindingVariables map[string]*Variable
}

func NewEvalContext(ctx *hcl.EvalContext) *EvalContext {
	return &EvalContext{
		EvalContext:          ctx,
		LateBindingVariables: make(map[string]*Variable),
	}
}
