package modconfig

import (
	"github.com/hashicorp/hcl/v2"
)

func NewThrowConfig(p *PipelineStepBase) *ThrowConfig {
	return &ThrowConfig{
		PipelineStepBase:     p,
		UnresolvedAttributes: make(map[string]hcl.Expression),
	}
}

type ThrowConfig struct {
	// Circular reference to its parent
	PipelineStepBase     *PipelineStepBase
	UnresolvedAttributes map[string]hcl.Expression

	If      *bool
	Message *string
}
