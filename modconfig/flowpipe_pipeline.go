package modconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func NewPipeline(mod *Mod, block *hcl.Block) *Pipeline {

	pipelineFullName := block.Labels[0]

	// TODO: rethink this area, we need to be able to handle pipelines that are not in a mod
	// TODO: we're trying to integrate the pipeline & trigger functionality into the mod system, so it will look
	// TODO: like a clutch for now
	if mod != nil {
		modName := mod.Name()
		if strings.HasPrefix(modName, "mod") {
			modName = strings.TrimPrefix(modName, "mod.")
		}
		pipelineFullName = modName + ".pipeline." + pipelineFullName
	} else {
		pipelineFullName = "local.pipeline." + pipelineFullName
	}

	pipeline := &Pipeline{
		HclResourceImpl: HclResourceImpl{
			// The FullName is the full name of the resource, including the mod name
			FullName:        pipelineFullName,
			UnqualifiedName: "pipeline." + block.Labels[0],
			DeclRange:       block.DefRange,
			blockType:       block.Type,
		},
		// TODO: hack to serialise pipeline name because HclResourceImpl is not serialised
		PipelineName: pipelineFullName,
		Params:       []PipelineParam{},
		mod:          mod,
	}

	return pipeline
}

type ResourceWithParam interface {
	GetParam(paramName string) *PipelineParam
	GetParams() []PipelineParam
}

// Pipeline represents a "pipeline" block in an flowpipe HCL (*.fp) file
//
// Note that this Pipeline definition is different that the pipeline that is running. This definition
// contains unresolved expressions (mostly in steps), how to handle errors etc but not the actual Pipeline
// execution data.
type Pipeline struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	mod *Mod

	// TODO: hack to serialise pipeline name because HclResourceImpl is not serialised
	PipelineName string `json:"pipeline_name"`

	// Unparsed HCL body, needed so we can de-code the step HCL into the correct struct
	RawBody hcl.Body `json:"-" hcl:",remain"`

	// Unparsed JSON raw message, needed so we can unmarshall the step JSON into the correct struct
	StepsRawJson json.RawMessage `json:"-"`

	Steps           []PipelineStep   `json:"steps,omitempty"`
	OutputConfig    []PipelineOutput `json:"outputs,omitempty"`
	Params          []PipelineParam  `json:"params,omitempty"`
	FileName        string           `json:"file_name"`
	StartLineNumber int              `json:"start_line_number"`
	EndLineNumber   int              `json:"end_line_number"`
}

func (p *Pipeline) GetParams() []PipelineParam {
	return p.Params
}

func (p *Pipeline) GetParam(paramName string) *PipelineParam {
	for _, param := range p.Params {
		if param.Name == paramName {
			return &param
		}
	}
	return nil
}

func (p *Pipeline) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	p.FileName = fileName
	p.StartLineNumber = startLineNumber
	p.EndLineNumber = endLineNumber
}

func ValidateParams(p ResourceWithParam, params map[string]interface{}) []error {
	errors := []error{}

	// Lists out all the pipeline params that don't have a default value
	pipelineParamsWithNoDefaultValue := map[string]bool{}
	for _, v := range p.GetParams() {
		if v.Default.IsNull() && !v.Optional {
			pipelineParamsWithNoDefaultValue[v.Name] = true
		}
	}

	for k, v := range params {

		param := p.GetParam(k)
		if param == nil {
			errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("unknown parameter specified '%s'", k)))
			continue
		}

		if !hclhelpers.GoTypeMatchesCtyType(v, param.Type) {
			wanted := param.Type.FriendlyName()
			typeOfInterface := reflect.TypeOf(v)
			if typeOfInterface == nil {
				errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted %s but received null", k, wanted)))
			} else {
				received := typeOfInterface.String()
				errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted %s but received %s", k, wanted, received)))
			}
		} else {
			delete(pipelineParamsWithNoDefaultValue, k)
		}
	}

	var missingParams []string
	for k := range pipelineParamsWithNoDefaultValue {
		missingParams = append(missingParams, k)
	}

	// Return error if there is no arguments provided for the pipeline params that don't have a default value
	if len(missingParams) > 0 {
		errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("missing parameter: %s", strings.Join(missingParams, ", "))))
	}

	return errors
}

// This is inefficient because we are coercing the value from string -> Go using Cty (because that's how the pipeline is defined)
// and again we convert from Go -> Cty when we're executing the pipeline to build EvalContext when we're evaluating
// data are not resolved during parse time.
func CoerceParams(p ResourceWithParam, params map[string]string) (map[string]interface{}, []error) {
	errors := []error{}

	// Lists out all the pipeline params that don't have a default value
	pipelineParamsWithNoDefaultValue := map[string]bool{}
	for _, p := range p.GetParams() {
		if p.Default.IsNull() && !p.Optional {
			pipelineParamsWithNoDefaultValue[p.Name] = true
		}
	}

	res := map[string]interface{}{}

	for k, v := range params {
		param := p.GetParam(k)
		if param == nil {
			errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("unknown parameter specified '%s'", k)))
			continue
		}

		val, moreErr := hclhelpers.CoerceStringToGoBasedOnCtyType(v, param.Type)
		if moreErr != nil {
			errors = append(errors, moreErr)
			continue
		}
		res[k] = val

		delete(pipelineParamsWithNoDefaultValue, k)
	}

	var missingParams []string
	for k := range pipelineParamsWithNoDefaultValue {
		missingParams = append(missingParams, k)
	}

	// Return error if there is no arguments provided for the pipeline params that don't have a default value
	if len(missingParams) > 0 {
		errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("missing parameter: %s", strings.Join(missingParams, ", "))))
	}

	return res, errors
}

func (p *Pipeline) ValidatePipelineParam(params map[string]interface{}) []error {
	return ValidateParams(p, params)
}

func (p *Pipeline) CoercePipelineParams(params map[string]string) (map[string]interface{}, []error) {
	return CoerceParams(p, params)
}

func (p *Pipeline) GetMod() *Mod {
	return p.mod
}

// Pipeline functions
func (p *Pipeline) GetStep(stepFullyQualifiedName string) PipelineStep {
	for i := 0; i < len(p.Steps); i++ {
		if p.Steps[i].GetFullyQualifiedName() == stepFullyQualifiedName {
			return p.Steps[i]
		}
	}
	return nil
}

func (p *Pipeline) CtyValue() (cty.Value, error) {
	baseCtyValue, err := p.HclResourceImpl.CtyValue()
	if err != nil {
		return cty.NilVal, err
	}

	pipelineVars := baseCtyValue.AsValueMap()
	pipelineVars[schema.LabelName] = cty.StringVal(p.Name())

	if p.Description != nil {
		pipelineVars[schema.AttributeTypeDescription] = cty.StringVal(*p.Description)
	}

	return cty.ObjectVal(pipelineVars), nil
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *Pipeline) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {

	var diags hcl.Diagnostics
	switch o := opts.(type) {
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}

func (ph *Pipeline) UnmarshalJSON(data []byte) error {
	// Define an auxiliary type to decode the JSON and capture the value of the 'ISteps' field
	type Aux struct {
		PipelineName string          `json:"pipeline_name"`
		Description  *string         `json:"description,omitempty"`
		Output       *string         `json:"output,omitempty"`
		Raw          json.RawMessage `json:"-"`
		ISteps       json.RawMessage `json:"steps"`
	}

	aux := Aux{ISteps: json.RawMessage([]byte("null"))} // Provide a default value for 'ISteps' field
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Assign values to the fields of the main struct

	ph.FullName = aux.PipelineName
	ph.PipelineName = aux.PipelineName
	ph.Description = aux.Description
	ph.StepsRawJson = []byte(aux.Raw)

	// Determine the concrete type of 'ISteps' based on the data present in the JSON
	if aux.ISteps != nil && string(aux.ISteps) != "null" {
		// Replace the JSON array of 'ISteps' with the desired concrete type
		var stepSlice []json.RawMessage
		if err := json.Unmarshal(aux.ISteps, &stepSlice); err != nil {
			return err
		}

		// Iterate over the stepSlice and determine the concrete type of each step
		for _, stepData := range stepSlice {
			// Extract the 'step_type' field from the stepData
			var stepType struct {
				StepType string `json:"step_type"`
			}
			if err := json.Unmarshal(stepData, &stepType); err != nil {
				return err
			}

			switch stepType.StepType {

			case schema.BlockTypePipelineStepHttp:
				var step PipelineStepHttp
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)

			case schema.BlockTypePipelineStepSleep:
				var step PipelineStepSleep
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)

			case schema.BlockTypePipelineStepEmail:
				var step PipelineStepEmail
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)

			case schema.BlockTypePipelineStepTransform:
				var step PipelineStepTransform
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)

			case schema.BlockTypePipelineStepQuery:
				var step PipelineStepQuery
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)

			case schema.BlockTypePipelineStepPipeline:
				var step PipelineStepPipeline
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)

			case schema.BlockTypePipelineStepFunction:
				var step PipelineStepFunction
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			case schema.BlockTypePipelineStepContainer:
				var step PipelineStepContainer
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			case schema.BlockTypePipelineStepInput:
				var step PipelineStepInput
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			case schema.BlockTypePipelineStepMessage:
				var step PipelineStepMessage
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			default:
				// Handle unrecognized step types or return an error
				return perr.BadRequestWithMessage(fmt.Sprintf("unrecognized step type '%s'", stepType.StepType))

			}
		}
	}

	return nil
}

func (p *Pipeline) Equals(other *Pipeline) bool {

	if p == nil && other == nil {
		return true
	}

	if p == nil && other != nil || p != nil && other == nil {
		return false
	}

	baseEqual := p.HclResourceImpl.Equals(&other.HclResourceImpl)
	if !baseEqual {
		return false
	}

	// Order of params does not matter, but the value does
	if len(p.Params) != len(other.Params) {
		return false
	}

	// Compare param values
	for _, v := range p.Params {
		otherParam := other.GetParam(v.Name)
		if otherParam == nil {
			return false
		}

		if !v.Equals(otherParam) {
			return false
		}
	}

	// catch name change of the other param
	for _, v := range other.Params {
		pParam := p.GetParam(v.Name)
		if pParam == nil {
			return false
		}
	}

	if len(p.Steps) != len(other.Steps) {
		return false
	}

	for i := 0; i < len(p.Steps); i++ {
		if !p.Steps[i].Equals(other.Steps[i]) {
			return false
		}
	}

	if len(p.OutputConfig) != len(other.OutputConfig) {
		return false
	}

	// build map for output so it's easier to lookup
	myOutput := map[string]*PipelineOutput{}
	for i, o := range p.OutputConfig {
		myOutput[o.Name] = &p.OutputConfig[i]
	}

	otherOutput := map[string]*PipelineOutput{}
	for i, o := range other.OutputConfig {
		otherOutput[o.Name] = &other.OutputConfig[i]
	}

	for k, v := range myOutput {
		if _, ok := otherOutput[k]; !ok {
			return false
		} else if !v.Equals(otherOutput[k]) {
			return false
		}
	}

	// check name changes on the other output
	for k := range otherOutput {
		if _, ok := myOutput[k]; !ok {
			return false
		}
	}

	return p.FullName == other.FullName &&
		p.GetMetadata().ModFullName == other.GetMetadata().ModFullName
}

func (p *Pipeline) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDescription:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				p.Description = &valString
			}
		case schema.AttributeTypeTitle:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				p.Title = &valString
			}
		case schema.AttributeTypeDocumentation:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				p.Documentation = &valString
			}
		case schema.AttributeTypeTags:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsValueMap()
				resultMap := make(map[string]string)
				for key, value := range valString {
					resultMap[key] = value.AsString()
				}
				p.Tags = resultMap
			}

		case schema.AttributeTypeMaxConcurrency:
			maxConcurrency, moreDiags := hclhelpers.AttributeToInt(attr, nil, false)
			if moreDiags != nil && moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			} else {
				mcInt := int(*maxConcurrency)
				p.MaxConcurrency = &mcInt
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for pipeline: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}
	return diags
}

// end pipeline functions

// Pipeline HclResource interface functions

func (p *Pipeline) OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *Pipeline) setBaseProperties() {
}

// end Pipeline Hclresource interface functions

type PipelineParam struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Optional    bool      `json:"optional,omitempty"`
	Default     cty.Value `json:"-"`
	Type        cty.Type  `json:"type"`
	TypeString  string    `json:"type_string"`
}

func (p *PipelineParam) Equals(other *PipelineParam) bool {
	if p == nil && other == nil {
		return true
	}

	if p == nil && p != nil || p != nil && p == nil {
		return false
	}

	if p.Default.Equals(other.Default) == cty.False {
		return false
	}

	return p.Name == other.Name &&
		p.Description == other.Description &&
		p.Optional == other.Optional &&
		p.Type.Equals(other.Type)
}

type PipelineOutput struct {
	Name                string         `json:"name"`
	Description         string         `json:"description,omitempty"`
	DependsOn           []string       `json:"depends_on,omitempty"`
	CredentialDependsOn []string       `json:"credential_depends_on,omitempty"`
	Resolved            bool           `json:"resolved,omitempty"`
	Value               interface{}    `json:"value,omitempty"`
	UnresolvedValue     hcl.Expression `json:"-"`
	Range               *hcl.Range     `json:"Range"`
}

// GetShowData implements the Showable interface
//func (o PipelineOutput) GetShowData() *printers.RowData {
//	return printers.NewRowData(
//		printers.NewFieldValue("Name", o.Name, printers.WithListKeyRender(o.renderName)),
//		printers.NewFieldValue("Description", o.Description),
//		printers.NewFieldValue("Type", "any"))
//}

//func (o *PipelineOutput) renderName(opts sanitize.RenderOptions) string {
//	au := aurora.NewAurora(opts.ColorEnabled)
//	return fmt.Sprintf("%s:", au.Cyan(o.Name))
//}

func (o *PipelineOutput) Equals(other *PipelineOutput) bool {
	// If both pointers are nil, they are considered equal
	if o == nil && other == nil {
		return true
	}

	// If one of the pointers is nil while the other is not, they are not equal
	if (o == nil && other != nil) || (o != nil && other == nil) {
		return false
	}

	// Compare Name field
	if o.Name != other.Name {
		return false
	}

	if !helpers.StringSliceEqualIgnoreOrder(o.DependsOn, other.DependsOn) {
		return false
	}

	// Compare Resolved field
	if o.Resolved != other.Resolved {
		return false
	}

	// Compare Value field using deep equality
	if !reflect.DeepEqual(o.Value, other.Value) {
		return false
	}

	// Compare UnresolvedValue field using deep equality
	if !hclhelpers.ExpressionsEqual(o.UnresolvedValue, other.UnresolvedValue) {
		return false
	}

	// All fields are equal
	return true
}

func (o *PipelineOutput) AppendDependsOn(dependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range o.DependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range dependsOn {
		if !existingDeps[dep] {
			o.DependsOn = append(o.DependsOn, dep)
			existingDeps[dep] = true
		}
	}
}

func (o *PipelineOutput) AppendCredentialDependsOn(credentialDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range o.CredentialDependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range credentialDependsOn {
		if !existingDeps[dep] {
			o.CredentialDependsOn = append(o.CredentialDependsOn, dep)
			existingDeps[dep] = true
		}
	}
}
