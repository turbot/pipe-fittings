package modconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepInput struct {
	PipelineStepBase

	InputType  string  `json:"type" cty:"type"`
	Prompt     *string `json:"prompt" cty:"prompt"`
	OptionList []PipelineStepInputOption

	// Notifier cty.Value `json:"-" cty:"notify"`
	Notifier NotifierImpl `json:"notify" cty:"-"`
}

func (p *PipelineStepInput) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	_, ok := iOther.(*PipelineStepInput)
	if !ok {
		return false
	}

	return p.Name == iOther.GetName()
}

func (p *PipelineStepInput) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	results := map[string]interface{}{}
	results[schema.AttributeTypeType] = p.InputType

	// prompt
	var prompt *string
	if p.UnresolvedAttributes[schema.AttributeTypePrompt] == nil {
		prompt = p.Prompt
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypePrompt], evalContext, &prompt)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}
	if prompt != nil {
		results[schema.AttributeTypePrompt] = *prompt
	}

	// options
	var err error
	var resolvedOpts []PipelineStepInputOption
	unresolvedOptBlocks := make(map[string]int)
	unresolvedBlockKeys := utils.SortedMapKeys(p.UnresolvedBodies)

	for _, ubk := range unresolvedBlockKeys {
		if strings.HasPrefix(ubk, schema.BlockTypeOption) && strings.Contains(ubk, ":") {
			if optIndex, err := strconv.Atoi(strings.Split(ubk, ":")[1]); err != nil {
				return results, perr.InternalWithMessage(fmt.Sprintf("unable to parse option index to int: %s", err.Error()))
			} else {
				unresolvedOptBlocks[ubk] = optIndex
			}
		}
	}

	if p.UnresolvedAttributes[schema.AttributeTypeOptions] == nil && len(unresolvedOptBlocks) == 0 && len(p.OptionList) > 0 {
		// everythings already resolved
		resolvedOpts = p.OptionList
	} else {
		if p.UnresolvedAttributes[schema.AttributeTypeOptions] != nil {
			// attribute needs resolving
			var opts cty.Value
			diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeOptions], evalContext, &opts)
			if diags.HasErrors() {
				return nil, error_helpers.HclDiagsToError(p.Name, diags)
			}
			resolvedOpts, err = CtyValueToPipelineStepInputOptionList(opts)
			if err != nil {
				return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse options attribute: " + err.Error())
			}
		} else if len(unresolvedOptBlocks) > 0 {
			// blocks need resolving
			for key, optIndex := range unresolvedOptBlocks {
				var o PipelineStepInputOption
				diags := gohcl.DecodeBody(p.UnresolvedBodies[key], evalContext, &o)
				if len(diags) > 0 {
					return nil, error_helpers.HclDiagsToError(p.Name, diags)
				}
				p.OptionList[optIndex] = o
			}
			resolvedOpts = p.OptionList
		}
	}
	results[schema.AttributeTypeOptions] = resolvedOpts

	// notifier
	if attr, ok := p.UnresolvedAttributes[schema.AttributeTypeNotifier]; !ok {
		results[schema.AttributeTypeNotifier] = p.Notifier
	} else {
		notifierCtyVal, moreDiags := attr.Value(evalContext)
		if moreDiags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, moreDiags)
		}

		notifier, err := ctyValueToPipelineStepNotifierValueMap(notifierCtyVal)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse notifier attribute: " + err.Error())
		}
		results[schema.AttributeTypeNotifier] = notifier
	}

	return results, nil
}

func (p *PipelineStepInput) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeType:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}
			if val != cty.NilVal {
				t, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeType + " attribute to string",
						Subject:  &attr.Range,
					})
					continue
				}
				p.InputType = t
			}
		case schema.AttributeTypePrompt:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}
			if val != cty.NilVal {
				prompt, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypePrompt + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Prompt = &prompt
			}
		case schema.AttributeTypeOptions:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				opts, ctyErr := CtyValueToPipelineStepInputOptionList(val)
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeOptions + " attribute to InputOption slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.OptionList = append(p.OptionList, opts...)
			}
		case schema.AttributeTypeNotifier:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				var err error
				p.Notifier, err = ctyValueToPipelineStepNotifierValueMap(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeNotifier + " attribute to InputNotifier",
						Detail:   err.Error(),
						Subject:  &attr.Range,
					})
				}
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Input Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepInput) SetBlockConfig(blocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	hasAttrOptions := len(p.OptionList) > 0 || p.UnresolvedAttributes["options"] != nil
	optionIndex := 0
	for _, b := range blocks {
		switch b.Type {
		case schema.BlockTypeOption:
			opt := PipelineStepInputOption{}
			if hasAttrOptions {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Option blocks and options attribute are mutually exclusive",
					Subject:  &b.DefRange,
				})
				continue
			}
			moreDiags := gohcl.DecodeBody(b.Body, evalContext, &opt)
			if len(moreDiags) > 0 {
				moreDiags = p.PipelineStepBase.HandleDecodeBodyDiags(moreDiags, fmt.Sprintf("%s:%d", schema.BlockTypeOption, optionIndex), b.Body)
				if len(moreDiags) > 0 {
					diags = append(diags, moreDiags...)
					continue
				}
			}
			if helpers.IsNil(opt.Value) {
				opt.Value = &b.Labels[0]
			}
			p.OptionList = append(p.OptionList, opt)
			optionIndex++
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported block type for Input Step: " + b.Type,
				Subject:  &b.DefRange,
			})
		}
	}

	return diags
}

func (p *PipelineStepInput) Validate() hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// validate type
	if !constants.IsValidInputType(p.InputType) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attribute " + schema.AttributeTypeType + " specified with invalid value " + p.InputType,
		})
	}

	// check for and validate style on options
	for _, o := range p.OptionList {
		if !helpers.IsNil(o.Style) && !constants.IsValidInputStyleType(*o.Style) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute " + schema.AttributeTypeStyle + " specified with invalid value " + *o.Style,
			})
		}
	}

	return diags
}

func ctyValueToPipelineStepNotifierValueMap(value cty.Value) (NotifierImpl, error) {
	notifier := NotifierImpl{}

	valueMap := value.AsValueMap()
	notifiesCty := valueMap[schema.AttributeTypeNotifies]

	if notifiesCty == cty.NilVal {
		return notifier, nil
	}

	notifiesCtySlice := notifiesCty.AsValueSlice()

	for _, notifyCty := range notifiesCtySlice {
		n, err := ctyValueToNotify(notifyCty)
		if err != nil {
			return notifier, err
		}
		notifier.Notifies = append(notifier.Notifies, n)
	}

	return notifier, nil
}

func ctyValueToNotify(val cty.Value) (Notify, error) {

	n := Notify{}

	if val.IsNull() {
		return n, nil
	}

	valMap := val.AsValueMap()

	cc := valMap["cc"]
	if cc != cty.NilVal {
		ccSlice := cc.AsValueSlice()
		for _, c := range ccSlice {
			n.Cc = append(n.Cc, c.AsString())
		}
	}

	bcc := valMap["bcc"]
	if bcc != cty.NilVal {
		bccSlice := bcc.AsValueSlice()
		for _, b := range bccSlice {
			n.Bcc = append(n.Bcc, b.AsString())
		}
	}

	channel := valMap["channel"]
	if channel != cty.NilVal {
		channel := channel.AsString()
		n.Channel = &channel
	}

	description := valMap["description"]
	if description != cty.NilVal {
		description := description.AsString()
		n.Description = &description
	}

	subject := valMap["subject"]
	if subject != cty.NilVal {
		subject := subject.AsString()
		n.Subject = &subject
	}

	title := valMap["title"]
	if title != cty.NilVal {
		title := title.AsString()
		n.Title = &title
	}

	to := valMap["to"]
	if to != cty.NilVal {
		toSlice := to.AsValueSlice()
		for _, t := range toSlice {
			n.To = append(n.To, t.AsString())
		}
	}

	integration := valMap["integration"]

	if integration != cty.NilVal {
		integration, err := integrationFromCtyValue(integration)
		if err != nil {
			return n, err
		}
		n.Integration = integration
	}

	return n, nil
}

type PipelineStepInputOption struct {
	Label    *string `json:"label" hcl:"label,optional"`
	Value    *string `json:"value" hcl:"value,optional"`
	Selected *bool   `json:"selected,omitempty" hcl:"selected,optional"`
	Style    *string `json:"style,omitempty" hcl:"style,optional"`
}

func CtyValueToPipelineStepInputOptionList(value cty.Value) ([]PipelineStepInputOption, error) {
	var output []PipelineStepInputOption

	opts := value.AsValueSlice()

	for _, opt := range opts {
		valueMap := opt.AsValueMap()

		isValid := false
		option := PipelineStepInputOption{}
		for k, v := range valueMap {
			switch k {
			case schema.AttributeTypeValue:
				if !v.IsNull() {
					isValid = true
					val := v.AsString()
					option.Value = &val
				}
			case schema.AttributeTypeLabel:
				if !v.IsNull() {
					label := v.AsString()
					option.Label = &label
				}
			case schema.AttributeTypeSelected:
				if !v.IsNull() && v.Type() == cty.Bool {
					isSelected := v.True()
					option.Selected = &isSelected
				}
			case schema.AttributeTypeStyle:
				if !v.IsNull() {
					s := v.AsString()
					option.Style = &s
				}
			default:
				return nil, perr.BadRequestWithMessage(k + " is not a valid attribute for input options")
			}
		}

		if isValid {
			output = append(output, option)
		} else {
			return nil, perr.BadRequestWithMessage("input options must declare a value")
		}
	}

	return output, nil
}

func (p *PipelineStepInputOption) Validate() hcl.Diagnostics {
	var diags hcl.Diagnostics

	// TODO: Figure out validation(s)

	return diags
}
