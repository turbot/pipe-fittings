package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepContainer struct {
	PipelineStepBase

	Image             *string           `json:"image"`
	Source            *string           `json:"source"`
	Cmd               []string          `json:"cmd"`
	Env               map[string]string `json:"env"`
	Entrypoint        []string          `json:"entrypoint"`
	CpuShares         *int64            `json:"cpu_shares"`
	Memory            *int64            `json:"memory"`
	MemoryReservation *int64            `json:"memory_reservation"`
	MemorySwap        *int64            `json:"memory_swap"`
	MemorySwappiness  *int64            `json:"memory_swappiness"`
	ReadOnly          *bool             `json:"read_only"`
	User              *string           `json:"user"`
	Workdir           *string           `json:"workdir"`
}

func (p *PipelineStepContainer) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || p != nil && helpers.IsNil(iOther) {
		return false
	}

	other, ok := iOther.(*PipelineStepContainer)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	return utils.PtrEqual(p.Image, other.Image) &&
		reflect.DeepEqual(p.Cmd, other.Cmd) &&
		reflect.DeepEqual(p.Env, other.Env)
}

func (p *PipelineStepContainer) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	results, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, err
	}

	var diags hcl.Diagnostics

	// image
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeImage, p.Image)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// source
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeSource, p.Source)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// cmd
	results, diags = stringSliceInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeCmd, &p.Cmd)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// env
	var env map[string]string
	if p.UnresolvedAttributes[schema.AttributeTypeEnv] == nil {
		env = p.Env
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeEnv], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}

		var err error
		env, err = hclhelpers.CtyToGoMapString(args)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse env attribute to map[string]string: " + err.Error())
		}
	}
	results[schema.AttributeTypeEnv] = env

	// entry_point
	results, diags = stringSliceInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeEntrypoint, &p.Entrypoint)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// cpu_shares
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeCpuShares, p.CpuShares)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// memory
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeMemory, p.Memory)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// memory_reservation
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeMemoryReservation, p.MemoryReservation)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// memory_swap
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeMemorySwap, p.MemorySwap)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// memory_swappiness
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeMemorySwappiness, p.MemorySwappiness)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// user
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeUser, p.User)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// workdir
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeWorkdir, p.Workdir)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// read_only
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeReadOnly, p.ReadOnly)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	results[schema.LabelName] = p.Name

	memorySwappinessI, ok := results[schema.AttributeTypeMemorySwappiness]
	if ok {
		memorySwappiness := memorySwappinessI.(int64)
		// If the attribute is using any reference, it can only be resolved at the runtime
		if !(memorySwappiness >= 0 && memorySwappiness <= 100) {
			return nil, perr.BadRequestWithMessage("The value of '" + schema.AttributeTypeMemorySwappiness + "' attribute must be between 0 and 100")
		}
		results[schema.AttributeTypeMemorySwappiness] = memorySwappiness
	}

	return results, nil
}

func (p *PipelineStepContainer) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeImage, schema.AttributeTypeSource, schema.AttributeTypeUser,
			schema.AttributeTypeWorkdir:

			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, structFieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeCmd:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				cmds, moreErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if moreErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeCmd + "' attribute to string slice",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Cmd = cmds
			}
		case schema.AttributeTypeEnv:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				env, moreErr := hclhelpers.CtyToGoMapString(val)
				if moreErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeEnv + "' attribute to string map",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Env = env
			}
		case schema.AttributeTypeEntrypoint:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				ep, moreErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if moreErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeEntrypoint + "' attribute to string slice",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Entrypoint = ep
			}
		case schema.AttributeTypeCpuShares:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				cpuShares, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeCpuShares + " attribute to integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.CpuShares = cpuShares
			}
		case schema.AttributeTypeMemory:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				memory, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMemory + " attribute to integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Memory = memory
			}
		case schema.AttributeTypeMemoryReservation:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				memoryReservation, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMemoryReservation + " attribute to integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.MemoryReservation = memoryReservation
			}
		case schema.AttributeTypeMemorySwap:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				memorySwap, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMemorySwap + " attribute to integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.MemorySwap = memorySwap
			}
		case schema.AttributeTypeMemorySwappiness:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				memorySwappiness, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMemorySwappiness + " attribute to integer",
						Subject:  &attr.Range,
					})
					continue
				}

				if !(*memorySwappiness >= 0 && *memorySwappiness <= 100) {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "The value of '" + schema.AttributeTypeMemorySwappiness + "' attribute must be between 0 and 100",
						Subject:  &attr.Range,
					})
				}

				p.MemorySwappiness = memorySwappiness
			}

		case schema.AttributeTypeReadOnly:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				readOnly, err := hclhelpers.CtyToGo(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeReadOnly + " attribute to integer",
						Subject:  &attr.Range,
					})
					continue
				}

				if boolVal, ok := readOnly.(bool); ok {
					p.ReadOnly = &boolVal
				}
			}
		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Function Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepContainer) Validate() hcl.Diagnostics {

	diags := hcl.Diagnostics{}

	// validate the base attributes
	stepBaseDiags := p.ValidateBaseAttributes()
	if stepBaseDiags.HasErrors() {
		diags = append(diags, stepBaseDiags...)
	}

	// Either source or image must be specified, but not both
	if p.Image != nil && p.Source != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Image and Source attributes are mutually exclusive: " + p.GetFullyQualifiedName(),
		})
	}

	return diags
}
