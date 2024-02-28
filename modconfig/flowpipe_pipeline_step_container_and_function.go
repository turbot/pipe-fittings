package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepContainer struct {
	PipelineStepBase

	Image             *string           `json:"image"`
	Source            *string           `json:"source"`
	Cmd               []string          `json:"cmd"`
	Env               map[string]string `json:"env"`
	EntryPoint        []string          `json:"entrypoint"`
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
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepContainer)
	if !ok {
		return false
	}

	return p.Image == other.Image && reflect.DeepEqual(p.Cmd, other.Cmd) && reflect.DeepEqual(p.Env, other.Env)
}

func (p *PipelineStepContainer) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	results, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, err
	}

	var diags hcl.Diagnostics

	// image
	results, diags = stringPtrInputFromAttribute(p, results, evalContext, schema.AttributeTypeImage, "Image")
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError(p.Name, diags)
	}

	// source
	results, diags = stringPtrInputFromAttribute(p, results, evalContext, schema.AttributeTypeSource, "Source")
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError(p.Name, diags)
	}

	// cmd
	results, diags = stringSliceInputFromAttribute(p, results, evalContext, schema.AttributeTypeCmd, "Cmd")
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError(p.Name, diags)
	}

	// env
	var env map[string]string
	if p.UnresolvedAttributes[schema.AttributeTypeEnv] == nil {
		env = p.Env
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeEnv], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		var err error
		env, err = hclhelpers.CtyToGoMapString(args)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse env attribute to map[string]string: " + err.Error())
		}
	}

	// entry_point
	results, diags = stringSliceInputFromAttribute(p, results, evalContext, schema.AttributeTypeEntryPoint, "EntryPoint")
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError(p.Name, diags)
	}

	var cpuShares *int64
	if p.UnresolvedAttributes[schema.AttributeTypeCpuShares] == nil {
		cpuShares = p.CpuShares
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCpuShares], evalContext, &cpuShares)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var memory *int64
	if p.UnresolvedAttributes[schema.AttributeTypeMemory] == nil {
		memory = p.Memory
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeMemory], evalContext, &memory)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var memoryReservation *int64
	if p.UnresolvedAttributes[schema.AttributeTypeMemoryReservation] == nil {
		memoryReservation = p.MemoryReservation
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeMemoryReservation], evalContext, &memoryReservation)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var memorySwap *int64
	if p.UnresolvedAttributes[schema.AttributeTypeMemorySwap] == nil {
		memorySwap = p.MemorySwap
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeMemorySwap], evalContext, &memorySwap)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var memorySwappiness *int64
	if p.UnresolvedAttributes[schema.AttributeTypeMemorySwappiness] == nil {
		memorySwappiness = p.MemorySwappiness
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeMemorySwappiness], evalContext, &memorySwappiness)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var containerUser *string
	if p.UnresolvedAttributes[schema.AttributeTypeUser] == nil {
		containerUser = p.User
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeUser], evalContext, &containerUser)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var workDir *string
	if p.UnresolvedAttributes[schema.AttributeTypeWorkdir] == nil {
		workDir = p.Workdir
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeWorkdir], evalContext, &workDir)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var readOnly *bool
	if p.UnresolvedAttributes[schema.AttributeTypeReadOnly] == nil {
		readOnly = p.ReadOnly
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeReadOnly], evalContext, &readOnly)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	results[schema.LabelName] = p.Name
	results[schema.AttributeTypeEnv] = env

	if cpuShares != nil {
		results[schema.AttributeTypeCpuShares] = *cpuShares
	}

	if memory != nil {
		results[schema.AttributeTypeMemory] = *memory
	}

	if memoryReservation != nil {
		results[schema.AttributeTypeMemoryReservation] = *memoryReservation
	}

	if memorySwap != nil {
		results[schema.AttributeTypeMemorySwap] = *memorySwap
	}

	if memorySwappiness != nil {
		// If the attribute is using any reference, it can only be resolved at the runtime
		if !(*memorySwappiness >= 0 && *memorySwappiness <= 100) {
			return nil, perr.BadRequestWithMessage("The value of '" + schema.AttributeTypeMemorySwappiness + "' attribute must be between 0 and 100")
		}
		results[schema.AttributeTypeMemorySwappiness] = *memorySwappiness
	}

	if containerUser != nil {
		results[schema.AttributeTypeUser] = *containerUser
	}

	if workDir != nil {
		results[schema.AttributeTypeWorkdir] = *workDir
	}

	if readOnly != nil {
		results[schema.AttributeTypeReadOnly] = *readOnly
	}

	return results, nil
}

func (p *PipelineStepContainer) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeImage:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				image, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeImage + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Image = &image
			}
		case schema.AttributeTypeSource:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				source, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSource + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Source = &source
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
		case schema.AttributeTypeEntryPoint:
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
						Summary:  "Unable to parse '" + schema.AttributeTypeEntryPoint + "' attribute to string slice",
						Subject:  &attr.Range,
					})
					continue
				}
				p.EntryPoint = ep
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
		case schema.AttributeTypeUser:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				containerUser, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeUser + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.User = &containerUser
			}
		case schema.AttributeTypeWorkdir:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				workDir, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeWorkdir + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Workdir = &workDir
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

type PipelineStepFunction struct {
	PipelineStepBase

	Function cty.Value `json:"-"`

	Runtime string `json:"runtime" cty:"runtime"`
	Source  string `json:"source" cty:"source"`
	Handler string `json:"handler" cty:"handler"`

	Event map[string]interface{} `json:"event"`
	Env   map[string]string      `json:"env"`
}

func (p *PipelineStepFunction) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepFunction)
	if !ok {
		return false
	}

	return p.Name == other.Name &&
		p.Runtime == other.Runtime &&
		p.Handler == other.Handler &&
		p.Source == other.Source
}

func (p *PipelineStepFunction) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	results, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, err
	}

	var env map[string]string
	if p.UnresolvedAttributes[schema.AttributeTypeEnv] == nil {
		env = p.Env
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeEnv], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		var err error
		env, err = hclhelpers.CtyToGoMapString(args)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse env attribute to map[string]string: " + err.Error())
		}
	}

	var event map[string]interface{}
	if p.UnresolvedAttributes[schema.AttributeTypeEvent] == nil {
		event = p.Event
	} else {
		val, diags := p.UnresolvedAttributes[schema.AttributeTypeEvent].Value(evalContext)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		var err error
		event, err = hclhelpers.CtyToGoMapInterface(val)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse event attribute to map[string]interface{}: " + err.Error())
		}
	}

	var src string
	if p.UnresolvedAttributes[schema.AttributeTypeSource] == nil {
		src = p.Source
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSource], evalContext, &src)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var runtime string
	if p.UnresolvedAttributes[schema.AttributeTypeRuntime] == nil {
		runtime = p.Runtime
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRuntime], evalContext, &runtime)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var handler string
	if p.UnresolvedAttributes[schema.AttributeTypeHandler] == nil {
		handler = p.Handler
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeHandler], evalContext, &handler)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	results[schema.LabelName] = p.PipelineName + "." + p.GetFullyQualifiedName()
	results[schema.AttributeTypeSource] = src
	results[schema.AttributeTypeRuntime] = runtime
	results[schema.AttributeTypeHandler] = handler
	results[schema.AttributeTypeEvent] = event
	results[schema.AttributeTypeEnv] = env

	return results, nil
}

func (p *PipelineStepFunction) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSource:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				p.Source = val.AsString()
			}

		case schema.AttributeTypeHandler:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				p.Handler = val.AsString()
			}

		case schema.AttributeTypeRuntime:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val != cty.NilVal {
				p.Runtime = val.AsString()
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
		case schema.AttributeTypeEvent:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val != cty.NilVal {
				events, err := hclhelpers.CtyToGoMapInterface(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeEvent + "' attribute to string map",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Event = events
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

func (p *PipelineStepFunction) Validate() hcl.Diagnostics {
	// validate the base attributes
	diags := p.ValidateBaseAttributes()
	return diags
}
