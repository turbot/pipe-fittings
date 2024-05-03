package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/hclhelpers"
	"github.com/turbot/pipe-fittings/v2/perr"
	"github.com/turbot/pipe-fittings/v2/schema"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepPipeline struct {
	PipelineStepBase

	Pipeline cty.Value `json:"-"`
	Args     Input     `json:"args"`
}

func (p *PipelineStepPipeline) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || p != nil && helpers.IsNil(iOther) {
		return false
	}

	other, ok := iOther.(*PipelineStepPipeline)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Check if the maps have the same number of elements
	if len(p.Args) != len(other.Args) {
		return false
	}

	// Iterate through the first map
	for key, value1 := range p.Args {
		// Check if the key exists in the second map
		if value2, ok := other.Args[key]; !ok {
			return false
		} else if !reflect.DeepEqual(value1, value2) {
			return false
		}
	}

	// and reverse
	for key := range other.Args {
		// Check if the key exists in the second map
		if _, ok := p.Args[key]; !ok {
			return false
		}
	}

	if p.Pipeline == cty.NilVal && other.Pipeline != cty.NilVal || p.Pipeline != cty.NilVal && other.Pipeline == cty.NilVal {
		return false
	}

	// this should never happen?
	if p.Pipeline == cty.NilVal && other.Pipeline == cty.NilVal {
		return true
	}

	pValueMap := p.Pipeline.AsValueMap()
	otherValueMap := other.Pipeline.AsValueMap()

	if len(pValueMap) != len(otherValueMap) {
		return false
	}

	if pValueMap[schema.LabelName] != otherValueMap[schema.LabelName] {
		return false
	}

	return p.Pipeline.AsValueMap()[schema.LabelName] == other.Pipeline.AsValueMap()[schema.LabelName]
}

func (p *PipelineStepPipeline) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	var pipeline string
	if p.UnresolvedAttributes[schema.AttributeTypePipeline] == nil {
		if p.Pipeline == cty.NilVal {
			return nil, perr.InternalWithMessage(p.Name + ": pipeline must be supplied")
		}
		valueMap := p.Pipeline.AsValueMap()
		pipelineNameCty := valueMap[schema.LabelName]
		pipeline = pipelineNameCty.AsString()

	} else {
		var pipelineCty cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypePipeline], evalContext, &pipelineCty)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}
		valueMap := pipelineCty.AsValueMap()
		pipelineNameCty := valueMap[schema.LabelName]
		pipeline = pipelineNameCty.AsString()
	}

	results := map[string]interface{}{}

	results[schema.AttributeTypePipeline] = pipeline

	if p.UnresolvedAttributes[schema.AttributeTypeArgs] != nil {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeArgs], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}

		mapValue, err := hclhelpers.CtyToGoMapInterface(args)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse args attribute to map[string]interface{}: " + err.Error())
		}
		results[schema.AttributeTypeArgs] = mapValue

	} else if p.Args != nil {
		results[schema.AttributeTypeArgs] = p.Args
	}

	return results, nil
}

func (p *PipelineStepPipeline) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypePipeline:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)

			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				p.Pipeline = val
			}

		case schema.AttributeTypeArgs:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				goVals, err2 := hclhelpers.CtyToGoMapInterface(val)
				if err2 != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeArgs + " attribute to Go values",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Args = goVals
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Pipeline Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}
