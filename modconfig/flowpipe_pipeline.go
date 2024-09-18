package modconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
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

func ValidateParams(p ResourceWithParam, inputParams map[string]interface{}, evalCtx *hcl.EvalContext) []error {
	errors := []error{}

	// Lists out all the pipeline params that don't have a default value
	pipelineParamsWithNoDefaultValue := map[string]bool{}
	for _, v := range p.GetParams() {
		if v.Default.IsNull() && !v.Optional {
			pipelineParamsWithNoDefaultValue[v.Name] = true
		}
	}

	for k, v := range inputParams {
		param := p.GetParam(k)
		if param == nil {
			errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("unknown parameter specified '%s'", k)))
			continue
		}

		errorExist := false

		if !hclhelpers.GoTypeMatchesCtyType(v, param.Type) {
			wanted := param.Type.FriendlyName()
			typeOfInterface := reflect.TypeOf(v)
			if typeOfInterface == nil {
				errorExist = true
				errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted %s but received null", k, wanted)))
			} else {
				received := typeOfInterface.String()
				errorExist = true
				errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted %s but received %s", k, wanted, received)))
			}
		} else {
			delete(pipelineParamsWithNoDefaultValue, k)
		}

		if !errorExist {
			errValidation := validateParam(param, v, evalCtx)
			if errValidation != nil {
				errors = append(errors, errValidation)
			}
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
func CoerceParams(p ResourceWithParam, inputParams map[string]string, evalCtx *hcl.EvalContext) (map[string]interface{}, []error) {
	errors := []error{}

	// Lists out all the pipeline params that don't have a default value
	pipelineParamsWithNoDefaultValue := map[string]bool{}
	for _, p := range p.GetParams() {
		if p.Default.IsNull() && !p.Optional {
			pipelineParamsWithNoDefaultValue[p.Name] = true
		}
	}

	res := map[string]interface{}{}

	for k, v := range inputParams {
		param := p.GetParam(k)
		if param == nil {
			errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("unknown parameter specified '%s'", k)))
			continue
		}

		var val interface{}
		if param.IsCustomType() {
			dottedStringParts := strings.Split(v, ".")
			if len(dottedStringParts) != 3 {
				errors = append(errors, perr.BadRequestWithMessage("invalid connection string format"))
				continue
			}

			val = map[string]interface{}{
				"name":          dottedStringParts[2],
				"type":          dottedStringParts[1],
				"resource_type": "connection",
				"temporary":     true,
			}

		} else if param.IsNotifierType() {
			return nil, []error{perr.BadRequestWithMessage("notifier type is not supported")}
		} else {
			var moreErr error
			val, moreErr = hclhelpers.CoerceStringToGoBasedOnCtyType(v, param.Type)
			if moreErr != nil {
				errors = append(errors, moreErr)
				continue
			}
		}
		res[k] = val

		delete(pipelineParamsWithNoDefaultValue, k)

		errValidation := validateParam(param, val, evalCtx)
		if errValidation != nil {
			errors = append(errors, errValidation)
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

	return res, errors
}

func validateParam(param *PipelineParam, inputParam interface{}, evalCtx *hcl.EvalContext) error {
	var valToValidate cty.Value
	var err error
	if !param.Type.HasDynamicTypes() && !param.IsCustomType() && !param.IsNotifierType() {
		valToValidate, err = gocty.ToCtyValue(inputParam, param.Type)
		if err != nil {
			return err
		}
	} else {
		// we'll do our best here
		valToValidate, err = hclhelpers.ConvertInterfaceToCtyValue(inputParam)
		if err != nil {
			return err
		}
	}
	validParam, diags, err := param.ValidateSetting(valToValidate, evalCtx)
	if err != nil {
		return err
	} else if !validParam {
		if len(diags) > 0 {
			return error_helpers.BetterHclDiagsToError(param.Name, diags)
		}
		return perr.BadRequestWithMessage("invalid value for param " + param.Name)
	}
	return nil
}

func (p *Pipeline) ValidatePipelineParam(params map[string]interface{}, evalCtx *hcl.EvalContext) []error {
	return ValidateParams(p, params, evalCtx)
}

func (p *Pipeline) CoercePipelineParams(params map[string]string, evalCtx *hcl.EvalContext) (map[string]interface{}, []error) {
	return CoerceParams(p, params, evalCtx)
}

// Implements ModItem interface
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

// SetOptions sets the options on the pipeline (not supported)
func (p *Pipeline) SetOptions(_ options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "options are not supported on pipelines",
		Subject:  &block.DefRange,
	}}
}

func (p *Pipeline) UnmarshalJSON(data []byte) error {
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

	p.FullName = aux.PipelineName
	p.PipelineName = aux.PipelineName
	p.Description = aux.Description
	p.StepsRawJson = []byte(aux.Raw)

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
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepSleep:
				var step PipelineStepSleep
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepEmail:
				var step PipelineStepEmail
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepTransform:
				var step PipelineStepTransform
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepQuery:
				var step PipelineStepQuery
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepPipeline:
				var step PipelineStepPipeline
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

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
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Optional    bool              `json:"optional,omitempty"`
	Default     cty.Value         `json:"-"`
	Enum        cty.Value         `json:"-"`
	EnumGo      []any             `json:"enum"`
	Type        cty.Type          `json:"-"`
	TypeString  string            `json:"type_string"`
	Tags        map[string]string `json:"tags,omitempty"`
}

func (p *PipelineParam) IsCustomType() bool {
	if !p.Type.IsCapsuleType() && !(p.Type.IsListType() && p.Type.ElementType().IsCapsuleType()) {
		return false
	}

	var encapsulatedGoType reflect.Type
	if p.Type.IsListType() {
		encapsulatedGoType = p.Type.ElementType().EncapsulatedType()
	} else {
		encapsulatedGoType = p.Type.EncapsulatedType()
	}

	encapulatedInstanceNew := reflect.New(encapsulatedGoType)
	if _, ok := encapulatedInstanceNew.Interface().(PipelingConnection); ok {
		return true
	}

	if encapsulatedGoType.String() == "*modconfig.ConnectionImpl" {
		return true
	} else if encapsulatedGoType.String() == "*modconfig.NotifierImpl" {
		return true
	}

	return false
}

func (p *PipelineParam) IsNotifierType() bool {
	return false
}

func (p *PipelineParam) Equals(other *PipelineParam) bool {
	if p == nil && other == nil {
		return true
	}

	if p == nil && other != nil || p != nil && other == nil {
		return false
	}

	if p.Default.Equals(other.Default) == cty.False {
		return false
	}

	if p.Enum.Equals(other.Enum) == cty.False {
		return false
	}

	return p.Name == other.Name &&
		p.Description == other.Description &&
		p.Optional == other.Optional &&
		p.Type.Equals(other.Type)
}

func (p *PipelineParam) ValidateSetting(setting cty.Value, evalCtx *hcl.EvalContext) (bool, hcl.Diagnostics, error) {
	if setting.IsNull() {
		return true, hcl.Diagnostics{}, nil
	}

	// Helper function to perform capsule type and list type validations
	validateCustomType := func() (bool, hcl.Diagnostics) {
		ctdiags := CustomTypeValidation(nil, setting, p.Type)
		if len(ctdiags) > 0 {
			return false, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid type for param " + p.Name,
				Detail:   "The param type is not compatible with the given value",
			}}
		}

		ctdiags = p.PipelineParamCustomValueValidation(setting, evalCtx)
		return len(ctdiags) == 0, ctdiags
	}

	// Check for capsule type or list of capsule types
	if p.Type.IsCapsuleType() || (p.Type.IsListType() && p.Type.ListElementType().IsCapsuleType()) {
		valid, diags := validateCustomType()
		return valid, diags, nil
	} else if !hclhelpers.IsValueCompatibleWithType(p.Type, setting) {
		// This is non-capsule type compatibility check
		return false, hcl.Diagnostics{}, nil
	}

	// Enum-based validation
	valid, err := hclhelpers.ValidateSettingWithEnum(setting, p.Enum)
	return valid, hcl.Diagnostics{}, err
}

func (p *PipelineParam) PipelineParamCustomValueListValidation(setting cty.Value, evalCtx *hcl.EvalContext) hcl.Diagnostics {

	if !hclhelpers.IsCollectionOrTuple(setting.Type()) {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid value for param " + p.Name,
			Detail:   "The value for param must be a list",
		}
		return hcl.Diagnostics{diag}
	}

	var diags hcl.Diagnostics
	for it := setting.ElementIterator(); it.Next(); {
		_, element := it.Element()
		diags = append(diags, p.PipelineParamCustomValueValidation(element, evalCtx)...)
	}

	return diags
}

func (p *PipelineParam) PipelineParamCustomValueValidation(setting cty.Value, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	// this time we check if the given setting, i.e.
	// name = "example
	// type = "aws"

	// for connection actually exists in the eval context

	if hclhelpers.IsCollectionOrTuple(setting.Type()) {
		return p.PipelineParamCustomValueListValidation(setting, evalCtx)
	}

	if !setting.Type().IsObjectType() && !setting.Type().IsMapType() {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "The value for param must be an object",
		}
		return hcl.Diagnostics{diag}
	}

	settingValueMap := setting.AsValueMap()

	if settingValueMap["resource_type"].IsNull() || settingValueMap["name"].IsNull() {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "The value for param must have a 'resource_type', and 'name' key",
		}
		return hcl.Diagnostics{diag}
	}

	resourceType := settingValueMap["resource_type"].AsString()
	if resourceType == schema.BlockTypeConnection {
		if settingValueMap["type"].IsNull() {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "The value for param must have a 'type' key",
			}
			return hcl.Diagnostics{diag}
		}

		// check if the connection actually exists in the eval context
		allConnections := evalCtx.Variables[schema.BlockTypeConnection]
		if allConnections == cty.NilVal {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "No connection found",
			}
			return hcl.Diagnostics{diag}
		}

		connectionType := settingValueMap["type"].AsString()
		connectionName := settingValueMap["name"].AsString()

		if allConnections.Type().IsMapType() || allConnections.Type().IsObjectType() {
			allConnectionsMap := allConnections.AsValueMap()
			if allConnectionsMap[connectionType].IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No connection found for the given connection type",
				}
				return hcl.Diagnostics{diag}
			}

			connectionTypeMap := allConnectionsMap[connectionType].AsValueMap()
			if connectionTypeMap[connectionName].IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No connection found for the given connection name",
				}
				return hcl.Diagnostics{diag}
			} else {
				// TRUE
				return hcl.Diagnostics{}
			}
		}
	} else if resourceType == schema.BlockTypeNotifier {
		// check if the connection actually exists in the eval context
		allNotifiers := evalCtx.Variables[schema.BlockTypeNotifier]
		if allNotifiers == cty.NilVal {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "No notifier found",
			}
			return hcl.Diagnostics{diag}
		}

		notifierName := settingValueMap["name"].AsString()

		if allNotifiers.Type().IsMapType() || allNotifiers.Type().IsObjectType() {
			allNotifiersMap := allNotifiers.AsValueMap()

			if allNotifiersMap[notifierName].IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No noitifier found for the given notifier name",
				}
				return hcl.Diagnostics{diag}
			} else {
				// TRUE
				return hcl.Diagnostics{}
			}
		}
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid value for param " + p.Name,
		Detail:   "Invalid value for param " + p.Name,
	}
	return hcl.Diagnostics{diag}

}

type PipelineOutput struct {
	Name                string         `json:"name"`
	Description         string         `json:"description,omitempty"`
	DependsOn           []string       `json:"depends_on,omitempty"`
	CredentialDependsOn []string       `json:"credential_depends_on,omitempty"`
	ConnectionDependsOn []string       `json:"connection_depends_on,omitempty"`
	Resolved            bool           `json:"resolved,omitempty"`
	Value               interface{}    `json:"value,omitempty"`
	UnresolvedValue     hcl.Expression `json:"-"`
	Range               *hcl.Range     `json:"Range"`
}

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

func (o *PipelineOutput) AppendConnectionDependsOn(connectionDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range o.ConnectionDependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range connectionDependsOn {
		if !existingDeps[dep] {
			o.ConnectionDependsOn = append(o.ConnectionDependsOn, dep)
			existingDeps[dep] = true
		}
	}
}
