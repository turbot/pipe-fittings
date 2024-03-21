package modconfig

import (
	"reflect"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type LoopContainerStep struct {
	Until             bool               `json:"until" hcl:"until" cty:"until"`
	Image             *string            `json:"image,omitempty" hcl:"image,optional" cty:"image"`
	Source            *string            `json:"source,omitempty" hcl:"source,optional" cty:"source"`
	Cmd               *[]string          `json:"cmd,omitempty" hcl:"cmd,optional" cty:"cmd"`
	Env               *map[string]string `json:"env,omitempty" hcl:"env,optional" cty:"env"`
	EntryPoint        *[]string          `json:"entrypoint,omitempty" hcl:"entrypoint,optional" cty:"entrypoint"`
	CpuShares         *int64             `json:"cpu_shares,omitempty" hcl:"cpu_shares,optional" cty:"cpu_shares"`
	Memory            *int64             `json:"memory,omitempty" hcl:"memory,optional" cty:"memory"`
	MemoryReservation *int64             `json:"memory_reservation,omitempty" hcl:"memory_reservation,optional" cty:"memory_reservation"`
	MemorySwap        *int64             `json:"memory_swap,omitempty" hcl:"memory_swap,optional" cty:"memory_swap"`
	MemorySwappiness  *int64             `json:"memory_swappiness,omitempty" hcl:"memory_swappiness,optional" cty:"memory_swappiness"`
	ReadOnly          *bool              `json:"read_only,omitempty" hcl:"read_only,optional" cty:"read_only"`
	User              *string            `json:"user,omitempty" hcl:"user,optional" cty:"user"`
	Workdir           *string            `json:"workdir,omitempty" hcl:"workdir,optional" cty:"workdir"`
}

func (s *LoopContainerStep) Equals(other LoopDefn) bool {
	if s == nil && helpers.IsNil(other) {
		return true
	}

	if s == nil && !helpers.IsNil(other) || s != nil && helpers.IsNil(other) {
		return false
	}

	otherLoopContainerStep, ok := other.(*LoopContainerStep)
	if !ok {
		return false
	}

	// compare env using reflection
	if reflect.DeepEqual(s.Env, otherLoopContainerStep.Env) {
		return false
	}

	if s.Cmd == nil && otherLoopContainerStep.Cmd != nil || s.Cmd != nil && otherLoopContainerStep.Cmd == nil {
		return false
	} else if s.Cmd != nil {
		if slices.Compare(*s.Cmd, *otherLoopContainerStep.Cmd) != 0 {
			return false
		}
	}

	if s.EntryPoint == nil && otherLoopContainerStep.EntryPoint != nil || s.EntryPoint != nil && otherLoopContainerStep.EntryPoint == nil {
		return false
	} else if s.EntryPoint != nil {
		if slices.Compare(*s.EntryPoint, *otherLoopContainerStep.EntryPoint) != 0 {
			return false
		}
	}

	if s.Env == nil && otherLoopContainerStep.Env != nil || s.Env != nil && otherLoopContainerStep.Env == nil {
		return false
	} else if s.Env != nil {
		if !reflect.DeepEqual(*s.Env, *otherLoopContainerStep.Env) {
			return false
		}
	}

	return s.Until == otherLoopContainerStep.Until &&
		utils.PtrEqual(s.Image, otherLoopContainerStep.Image) &&
		utils.PtrEqual(s.Source, otherLoopContainerStep.Source) &&
		utils.PtrEqual(s.CpuShares, otherLoopContainerStep.CpuShares) &&
		utils.PtrEqual(s.Memory, otherLoopContainerStep.Memory) &&
		utils.PtrEqual(s.MemoryReservation, otherLoopContainerStep.MemoryReservation) &&
		utils.PtrEqual(s.MemorySwap, otherLoopContainerStep.MemorySwap) &&
		utils.PtrEqual(s.MemorySwappiness, otherLoopContainerStep.MemorySwappiness) &&
		utils.BoolPtrEqual(s.ReadOnly, otherLoopContainerStep.ReadOnly) &&
		utils.PtrEqual(s.User, otherLoopContainerStep.User) &&
		utils.PtrEqual(s.Workdir, otherLoopContainerStep.Workdir)
}

func (s *LoopContainerStep) GetType() string {
	return schema.BlockTypePipelineStepContainer
}

func (s *LoopContainerStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	if s.Cmd != nil {
		input[schema.AttributeTypeCmd] = *s.Cmd
	}

	if s.EntryPoint != nil {
		input[schema.AttributeTypeEntryPoint] = *s.EntryPoint
	}

	if s.Env != nil {
		input[schema.AttributeTypeEnv] = *s.Env
	}

	if s.Image != nil {
		input[schema.AttributeTypeImage] = *s.Image
	}

	if s.Source != nil {
		input[schema.AttributeTypeSource] = *s.Source
	}

	if s.CpuShares != nil {
		input[schema.AttributeTypeCpuShares] = *s.CpuShares
	}

	if s.Memory != nil {
		input[schema.AttributeTypeMemory] = *s.Memory
	}

	if s.MemoryReservation != nil {
		input[schema.AttributeTypeMemoryReservation] = *s.MemoryReservation
	}

	if s.MemorySwap != nil {
		input[schema.AttributeTypeMemorySwap] = *s.MemorySwap
	}

	if s.MemorySwappiness != nil {
		input[schema.AttributeTypeMemorySwappiness] = *s.MemorySwappiness
	}

	if s.ReadOnly != nil {
		input[schema.AttributeTypeReadOnly] = *s.ReadOnly
	}

	if s.User != nil {
		input[schema.AttributeTypeUser] = *s.User
	}

	if s.Workdir != nil {
		input[schema.AttributeTypeWorkdir] = *s.Workdir
	}

	return input, nil
}

func (s *LoopContainerStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}
