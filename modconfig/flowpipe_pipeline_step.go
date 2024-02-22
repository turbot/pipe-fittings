package modconfig

import (
	"fmt"
	"reflect"
	"slices"
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
	"github.com/turbot/terraform-components/addrs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

const (
	HttpMethodGet    = "get"
	HttpMethodPost   = "post"
	HttpMethodPut    = "put"
	HttpMethodDelete = "delete"
	HttpMethodPatch  = "patch"
)

var ValidHttpMethods = []string{
	HttpMethodGet,
	HttpMethodPost,
	HttpMethodPut,
	HttpMethodDelete,
	HttpMethodPatch,
}

type StepForEach struct {
	ForEachStep bool                 `json:"for_each_step"`
	Key         string               `json:"key"  binding:"required"`
	Output      *Output              `json:"output,omitempty"`
	TotalCount  int                  `json:"total_count" binding:"required"`
	Each        json.SimpleJSONValue `json:"each" swaggerignore:"true"`
}

type StepLoop struct {
	Index         int    `json:"index" binding:"required"`
	Input         *Input `json:"input,omitempty"`
	LoopCompleted bool   `json:"loop_completed"`
}

type StepRetry struct {
	Count          int    `json:"count" binding:"required"`
	Input          *Input `json:"input,omitempty"`
	RetryCompleted bool   `json:"retry_completed"`
}

// Input to the step or pipeline execution
type Input map[string]interface{}

func (i *Input) AsCtyMap() (map[string]cty.Value, error) {
	if i == nil {
		return map[string]cty.Value{}, nil
	}

	variables := make(map[string]cty.Value)

	for key, value := range *i {
		if value == nil || key == "step_name" {
			continue
		}

		ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(value)
		if err != nil {
			return nil, err
		}

		variables[key] = ctyVal
	}

	return variables, nil
}

// Output is the output from a step execution.
type Output struct {
	Status      string      `json:"status,omitempty"`
	FailureMode string      `json:"failure_mode,omitempty"`
	Data        OutputData  `json:"data,omitempty"`
	Errors      []StepError `json:"errors,omitempty"`
}

type OutputData map[string]interface{}

func (o *Output) Get(key string) interface{} {
	if o == nil {
		return nil
	}
	return o.Data[key]
}

func (o *Output) Set(key string, value interface{}) {
	if o == nil {
		return
	}
	o.Data[key] = value
}

func (o *Output) HasErrors() bool {
	if o == nil {
		return false
	}

	return o.Errors != nil && len(o.Errors) > 0
}

func (o *Output) AsCtyMap() (map[string]cty.Value, error) {
	if o == nil {
		return map[string]cty.Value{}, nil
	}

	variables := make(map[string]cty.Value)

	for key, value := range o.Data {
		if value == nil {
			continue
		}

		ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(value)
		if err != nil {
			return nil, err
		}

		variables[key] = ctyVal
	}

	if o.Errors != nil {
		errList := []cty.Value{}
		for _, stepErr := range o.Errors {
			ctyMap := map[string]cty.Value{}
			var err error
			errorAttributes := map[string]cty.Type{
				"instance": cty.String,
				"detail":   cty.String,
				"type":     cty.String,
				"title":    cty.String,
				"status":   cty.Number,
			}

			errorObject := map[string]interface{}{
				"instance": stepErr.Error.Instance,
				"detail":   stepErr.Error.Detail,
				"type":     stepErr.Error.Type,
				"title":    stepErr.Error.Title,
				"status":   stepErr.Error.Status,
			}
			ctyMap["error"], err = gocty.ToCtyValue(errorObject, cty.Object(errorAttributes))
			if err != nil {
				return nil, err
			}
			ctyMap["pipeline_execution_id"], err = gocty.ToCtyValue(stepErr.PipelineExecutionID, cty.String)
			if err != nil {
				return nil, err
			}
			ctyMap["step_execution_id"], err = gocty.ToCtyValue(stepErr.StepExecutionID, cty.String)
			if err != nil {
				return nil, err
			}
			ctyMap["pipeline"], err = gocty.ToCtyValue(stepErr.Pipeline, cty.String)
			if err != nil {
				return nil, err
			}
			ctyMap["step"], err = gocty.ToCtyValue(stepErr.Step, cty.String)
			if err != nil {
				return nil, err
			}
			errList = append(errList, cty.ObjectVal(ctyMap))
		}
		variables["errors"] = cty.ListVal(errList)
	}
	return variables, nil
}

type StepError struct {
	PipelineExecutionID string          `json:"pipeline_execution_id"`
	StepExecutionID     string          `json:"step_execution_id"`
	Pipeline            string          `json:"pipeline"`
	Step                string          `json:"step"`
	Error               perr.ErrorModel `json:"error"`
}

type NextStepAction string

const (
	// Default Next Step action which is just to start them, note that
	// the step may yet be "skipped" if the IF clause is preventing the step
	// to actually start, but at the very least we can "start" the step.
	NextStepActionStart NextStepAction = "start"

	// This happens if the step can't be started because one of it's dependency as failed
	//
	// Q: So why would step failure does not mean pipeline fail straight away?
	// A: We can't raise the pipeline fail command if there's "ignore error" directive on the step.
	//    If there are steps that depend on the failed step, these steps becomes "inaccessible", they can't start
	//    because the dependend step has failed.
	//
	NextStepActionInaccessible NextStepAction = "inaccessible"

	NextStepActionSkip NextStepAction = "skip"
)

type NextStep struct {
	StepName    string         `json:"step_name"`
	Action      NextStepAction `json:"action"`
	StepForEach *StepForEach   `json:"step_for_each,omitempty"`
	StepLoop    *StepLoop      `json:"step_loop,omitempty"`
	Input       Input          `json:"input"`
}

func NewPipelineStep(stepType, stepName string) PipelineStep {
	var step PipelineStep
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		step = &PipelineStepHttp{}
	case schema.BlockTypePipelineStepSleep:
		step = &PipelineStepSleep{}
	case schema.BlockTypePipelineStepEmail:
		step = &PipelineStepEmail{}
	case schema.BlockTypePipelineStepTransform:
		step = &PipelineStepTransform{}
	case schema.BlockTypePipelineStepQuery:
		step = &PipelineStepQuery{}
	case schema.BlockTypePipelineStepPipeline:
		step = &PipelineStepPipeline{}
	case schema.BlockTypePipelineStepFunction:
		step = &PipelineStepFunction{}
	case schema.BlockTypePipelineStepContainer:
		step = &PipelineStepContainer{}
	case schema.BlockTypePipelineStepInput:
		step = &PipelineStepInput{}
	default:
		return nil
	}

	step.Initialize()
	step.SetName(stepName)
	step.SetType(stepType)

	return step
}

// A common interface that all pipeline steps must implement
type PipelineStep interface {
	PipelineStepBaseInterface

	Initialize()
	GetFullyQualifiedName() string
	GetName() string
	SetName(string)
	GetType() string
	SetType(string)
	SetPipelineName(string)
	GetPipelineName() string
	IsResolved() bool
	GetUnresolvedAttributes() map[string]hcl.Expression
	AddUnresolvedBody(string, hcl.Body)
	GetUnresolvedBodies() map[string]hcl.Body
	GetInputs(*hcl.EvalContext) (map[string]interface{}, error)
	GetDependsOn() []string
	GetCredentialDependsOn() []string
	GetForEach() hcl.Expression
	SetAttributes(hcl.Attributes, *hcl.EvalContext) hcl.Diagnostics
	SetBlockConfig(hcl.Blocks, *hcl.EvalContext) hcl.Diagnostics
	SetErrorConfig(*ErrorConfig)
	GetErrorConfig(*hcl.EvalContext, bool) (*ErrorConfig, hcl.Diagnostics)
	GetRetryConfig(*hcl.EvalContext, bool) (*RetryConfig, hcl.Diagnostics)
	GetThrowConfig() []ThrowConfig
	SetOutputConfig(map[string]*PipelineOutput)
	GetOutputConfig() map[string]*PipelineOutput
	Equals(other PipelineStep) bool
	Validate() hcl.Diagnostics
	SetFileReference(fileName string, startLineNumber int, endLineNumber int)
	GetMaxConcurrency() *int
}

type PipelineStepBaseInterface interface {
	AppendDependsOn(...string)
	AppendCredentialDependsOn(...string)
	AddUnresolvedAttribute(string, hcl.Expression)
}

type ErrorConfig struct {
	// This means that invalid attributes must be validated "manually"
	If hcl.Body `json:"-" hcl:",remain"`

	Ignore *bool `json:"ignore,omitempty" hcl:"ignore,optional" cty:"ignore"`
}

func (e *ErrorConfig) Validate() bool {
	return true
}

func (ec *ErrorConfig) Equals(other *ErrorConfig) bool {
	if ec == nil || other == nil {
		return false
	}

	// Compare Ignore
	if ec.Ignore != other.Ignore {
		return false
	}

	return true
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`

	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
}

func (b *BasicAuthConfig) GetInputs(evalContext *hcl.EvalContext, unresolvedAttributes map[string]hcl.Expression) (*BasicAuthConfig, hcl.Diagnostics) {
	var username, password string
	if unresolvedAttributes[schema.AttributeTypeUsername] != nil {
		diags := gohcl.DecodeExpression(unresolvedAttributes[schema.AttributeTypeUsername], evalContext, &username)
		if diags.HasErrors() {
			return nil, diags
		}
		b.Username = username
	}
	if unresolvedAttributes[schema.AttributeTypePassword] != nil {
		diags := gohcl.DecodeExpression(unresolvedAttributes[schema.AttributeTypePassword], evalContext, &password)
		if diags.HasErrors() {
			return nil, diags
		}
		b.Password = password
	}
	return b, nil
}

func ctyValueToPipelineStepInputNotifyValueMap(value cty.Value) (NotifierImpl, error) {
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

// A common base struct that all pipeline steps must embed
type PipelineStepBase struct {
	Title               *string                    `json:"title,omitempty"`
	Description         *string                    `json:"description,omitempty"`
	Name                string                     `json:"name"`
	Type                string                     `json:"step_type"`
	PipelineName        string                     `json:"pipeline_name,omitempty"`
	Timeout             interface{}                `json:"timeout,omitempty"`
	DependsOn           []string                   `json:"depends_on,omitempty"`
	CredentialDependsOn []string                   `json:"credential_depends_on,omitempty"`
	Resolved            bool                       `json:"resolved,omitempty"`
	ErrorConfig         *ErrorConfig               `json:"-"`
	RetryConfig         *RetryConfig               `json:"retry,omitempty"`
	ThrowConfig         []ThrowConfig              `json:"throw,omitempty"`
	OutputConfig        map[string]*PipelineOutput `json:"-"`
	FileName            string                     `json:"file_name"`
	StartLineNumber     int                        `json:"start_line_number"`
	EndLineNumber       int                        `json:"end_line_number"`
	MaxConcurrency      *int                       `json:"max_concurrency,omitempty"`

	// This cant' be serialised
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	UnresolvedBodies     map[string]hcl.Body       `json:"-"`
	ForEach              hcl.Expression            `json:"-"`
}

func (p *PipelineStepBase) Initialize() {
	p.UnresolvedAttributes = make(map[string]hcl.Expression)
	p.UnresolvedBodies = make(map[string]hcl.Body)
}

func (p *PipelineStepBase) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	p.FileName = fileName
	p.StartLineNumber = startLineNumber
	p.EndLineNumber = endLineNumber
}

func (p *PipelineStepBase) GetRetryConfig(evalContext *hcl.EvalContext, ifResolution bool) (*RetryConfig, hcl.Diagnostics) {
	retryBlock := p.UnresolvedBodies[schema.BlockTypeRetry]

	if retryBlock == nil {
		return p.RetryConfig, hcl.Diagnostics{}
	}

	retryConfig := NewRetryConfig()

	if ifResolution {
		retryBlockAttribs, diags := retryBlock.JustAttributes()
		if len(diags) > 0 {
			return nil, diags
		}

		ifAttribute := retryBlockAttribs[schema.AttributeTypeIf]
		if ifAttribute != nil && ifAttribute.Expr != nil {
			// check if we should run this retry
			ifValue, diags := ifAttribute.Expr.Value(evalContext)
			if len(diags) > 0 {
				return nil, diags
			}

			// If the `if` attribute returns "false" then we return nil for the retry config, thus we won't be retrying it
			if !ifValue.True() {
				return nil, hcl.Diagnostics{}
			}
		}
	}

	diags := gohcl.DecodeBody(p.UnresolvedBodies[schema.BlockTypeRetry], nil, retryConfig)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = append(diags, retryConfig.Validate()...)
	if len(diags) > 0 {
		return nil, diags
	}

	return retryConfig, hcl.Diagnostics{}

}

func (p *PipelineStepBase) GetThrowConfig() []ThrowConfig {
	return p.ThrowConfig
}

func (p *PipelineStepBase) SetBlockConfig(blocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	stepType := p.GetType()

	loopBlocks := blocks.ByType()[schema.BlockTypeLoop]
	if len(loopBlocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only one loop block is allowed per step",
			Subject:  &blocks.ByType()[schema.BlockTypeLoop][0].DefRange,
		})
	}

	if len(loopBlocks) == 1 {
		loopBlock := loopBlocks[0]

		loopDefn := GetLoopDefn(stepType)
		if loopDefn == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Loop block is not supported for step type %s", stepType),
				Subject:  &loopBlock.DefRange,
			})
		} else {
			moreDiags := gohcl.DecodeBody(loopBlock.Body, evalContext, loopDefn)

			// Loop should always be unresolved
			if len(moreDiags) > 0 {
				moreDiags = p.HandleDecodeBodyDiags(moreDiags, schema.BlockTypeLoop, loopBlock.Body)
				if len(moreDiags) > 0 {
					diags = append(diags, moreDiags...)
				}
			} else {
				// still need to add the loop to "unresolved" body even if it's fully resolved. Consider this scenario:
				//     loop {
				// 			until = try(result.response_body.next, null) == null
				// 			url   = try(result.response_body.next, "")
				//		}
				//
				// The above fragment will not result in diagnostics error.
				p.AddUnresolvedBody(schema.BlockTypeLoop, loopBlock.Body)
			}
		}
	}

	errorBlocks := blocks.ByType()[schema.BlockTypeError]
	if len(errorBlocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only one error block is allowed per step",
			Subject:  &blocks.ByType()[schema.BlockTypeError][0].DefRange,
		})
	}

	if len(errorBlocks) == 1 {
		errorBlock := errorBlocks[0]
		errorDefn := ErrorConfig{}

		var errorBlockAttributes hcl.Attributes
		errorBlockAttributes, diags = errorBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return diags
		}

		validAttributes := map[string]bool{
			"if":     true,
			"ignore": true,
		}

		attributesOK := true

		for _, a := range errorBlockAttributes {
			if !validAttributes[a.Name] {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Unsupported attribute '%s' in error block", a.Name),
					Subject:  &a.NameRange,
				})
				attributesOK = false
			}
		}

		if attributesOK {
			moreDiags := gohcl.DecodeBody(errorBlock.Body, evalContext, &errorDefn)
			if len(moreDiags) > 0 {
				moreDiags = p.HandleDecodeBodyDiags(moreDiags, schema.BlockTypeError, errorBlock.Body)
				if len(moreDiags) > 0 {
					diags = append(diags, moreDiags...)
				}
			} else if errorBlockAttributes != nil && errorBlockAttributes[schema.AttributeTypeIf] != nil {
				p.AddUnresolvedBody(schema.BlockTypeError, errorBlock.Body)
			} else {
				// fully resolved error block
				p.ErrorConfig = &errorDefn
			}
		}
	}

	retryBlocks := blocks.ByType()[schema.BlockTypeRetry]
	if len(retryBlocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only one retry block is allowed per step",
			Subject:  &blocks.ByType()[schema.BlockTypeRetry][0].DefRange,
		})
	}

	if len(retryBlocks) == 1 {
		retryBlock := retryBlocks[0]
		retryConfig := NewRetryConfig()

		var retryBlockAttributes hcl.Attributes
		retryBlockAttributes, diags = retryBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return diags
		}

		moreDiags := gohcl.DecodeBody(retryBlock.Body, evalContext, retryConfig)

		if len(moreDiags) > 0 {
			moreDiags = p.HandleDecodeBodyDiags(moreDiags, schema.BlockTypeRetry, retryBlock.Body)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
			}
		} else if retryBlockAttributes != nil && retryBlockAttributes[schema.AttributeTypeIf] != nil {
			p.AddUnresolvedBody(schema.BlockTypeRetry, retryBlock.Body)
		} else {
			// fully resolved retry block
			p.RetryConfig = retryConfig

			moreDiags := p.RetryConfig.Validate()
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
			}
		}

		for _, a := range retryBlockAttributes {
			if !helpers.StringSliceContains([]string{
				schema.AttributeTypeIf,
				"max_attempts",
				"strategy",
				"min_interval",
				"max_interval",
			}, a.Name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Unsupported attribute %s in retry block", a.Name),
					Subject:  &a.NameRange,
				})
			}
		}
	}

	throwBlocks := blocks.ByType()[schema.BlockTypeThrow]

	for _, throwBlock := range throwBlocks {
		throwConfig := ThrowConfig{}

		// Decode the loop block
		moreDiags := gohcl.DecodeBody(throwBlock.Body, evalContext, &throwConfig)

		if len(moreDiags) > 0 {
			moreDiags := p.HandleDecodeBodyDiags(moreDiags, schema.BlockTypeThrow, throwBlock.Body)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				// Fill in the blank if we have an error so it's easier for debugging purpose
			} else {
				throwConfig.UnresolvedBody = throwBlock.Body
				throwConfig.Unresolved = true
				p.ThrowConfig = append(p.ThrowConfig, throwConfig)
			}
		} else {
			p.ThrowConfig = append(p.ThrowConfig, throwConfig)
		}
	}

	return diags
}

func (p *PipelineStepBase) AddUnresolvedBody(name string, body hcl.Body) {
	p.UnresolvedBodies[name] = body
}

func (p *PipelineStepBase) GetUnresolvedBodies() map[string]hcl.Body {
	return p.UnresolvedBodies
}

func (p *PipelineStepBase) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (p *PipelineStepBase) Equals(otherBase *PipelineStepBase) bool {
	if p == nil || otherBase == nil {
		return false
	}

	// Compare Title
	if !reflect.DeepEqual(p.Title, otherBase.Title) {
		return false
	}

	// Compare Description
	if !reflect.DeepEqual(p.Description, otherBase.Description) {
		return false
	}

	// Compare Name
	if p.Name != otherBase.Name {
		return false
	}

	// Compare Type
	if p.Type != otherBase.Type {
		return false
	}

	// Compare Timeout
	if p.Timeout != otherBase.Timeout {
		return false
	}

	// Compare DependsOn slices
	if len(p.DependsOn) != len(otherBase.DependsOn) {
		return false
	}
	for i, dep := range p.DependsOn {
		if dep != otherBase.DependsOn[i] {
			return false
		}
	}

	// Compare Resolved
	if p.Resolved != otherBase.Resolved {
		return false
	}

	// Compare ErrorConfig (if not nil)
	if (p.ErrorConfig == nil && otherBase.ErrorConfig != nil) || (p.ErrorConfig != nil && otherBase.ErrorConfig == nil) {
		return false
	}
	if p.ErrorConfig != nil && otherBase.ErrorConfig != nil && !p.ErrorConfig.Equals(otherBase.ErrorConfig) {
		return false
	}

	// Compare UnresolvedAttributes (map comparison)
	if len(p.UnresolvedAttributes) != len(otherBase.UnresolvedAttributes) {
		return false
	}
	for key, expr := range p.UnresolvedAttributes {
		otherExpr, ok := otherBase.UnresolvedAttributes[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}

		// haven't found a good way to test check equality for two hcl expressions
	}

	// Compare ForEach (if not nil)
	if (p.ForEach == nil && otherBase.ForEach != nil) || (p.ForEach != nil && otherBase.ForEach == nil) {
		return false
	}
	if p.ForEach != nil && otherBase.ForEach != nil && !hclhelpers.ExpressionsEqual(p.ForEach, otherBase.ForEach) {
		return false
	}

	return true
}

func (p *PipelineStepBase) SetPipelineName(pipelineName string) {
	p.PipelineName = pipelineName
}

func (p *PipelineStepBase) GetPipelineName() string {
	return p.PipelineName
}

func (p *PipelineStepBase) SetErrorConfig(errorConfig *ErrorConfig) {
	p.ErrorConfig = errorConfig
}

func (p *PipelineStepBase) GetErrorConfig(evalContext *hcl.EvalContext, ifResolution bool) (*ErrorConfig, hcl.Diagnostics) {
	errorBlock := p.UnresolvedBodies[schema.BlockTypeError]

	if errorBlock == nil {
		return p.ErrorConfig, hcl.Diagnostics{}
	}

	errorConfig := &ErrorConfig{}

	if ifResolution {
		errorBlockAttribs, diags := errorBlock.JustAttributes()
		if len(diags) > 0 {
			return nil, diags
		}

		ifAttribute := errorBlockAttribs[schema.AttributeTypeIf]
		if ifAttribute != nil && ifAttribute.Expr != nil {
			// check if we should run this retry
			ifValue, diags := ifAttribute.Expr.Value(evalContext)
			if len(diags) > 0 {
				return nil, diags
			}

			// If the `if` attribute returns "false" then we return nil for the error config since it doesn't apply
			if !ifValue.True() {
				return nil, hcl.Diagnostics{}
			}
		}
	}

	diags := gohcl.DecodeBody(p.UnresolvedBodies[schema.BlockTypeError], nil, errorConfig)
	if len(diags) > 0 {
		return nil, diags
	}

	return errorConfig, hcl.Diagnostics{}
}

func (p *PipelineStepBase) SetOutputConfig(output map[string]*PipelineOutput) {
	p.OutputConfig = output
}

func (p *PipelineStepBase) GetOutputConfig() map[string]*PipelineOutput {
	return p.OutputConfig
}

func (p *PipelineStepBase) GetForEach() hcl.Expression {
	return p.ForEach
}

func (p *PipelineStepBase) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	p.UnresolvedAttributes[name] = expr
}

func (p *PipelineStepBase) GetUnresolvedAttributes() map[string]hcl.Expression {
	return p.UnresolvedAttributes
}

func (p *PipelineStepBase) SetName(name string) {
	p.Name = name
}

func (p *PipelineStepBase) GetName() string {
	return p.Name
}

func (p *PipelineStepBase) SetType(stepType string) {
	p.Type = stepType
}

func (p *PipelineStepBase) GetType() string {
	return p.Type
}

func (p *PipelineStepBase) GetDependsOn() []string {
	return p.DependsOn
}

func (p *PipelineStepBase) GetCredentialDependsOn() []string {
	return p.CredentialDependsOn
}

func (p *PipelineStepBase) IsResolved() bool {
	return len(p.UnresolvedAttributes) == 0
}

func (p *PipelineStepBase) SetResolved(resolved bool) {
	p.Resolved = resolved
}

func (p *PipelineStepBase) GetFullyQualifiedName() string {
	return p.Type + "." + p.Name
}

func (p *PipelineStepBase) AppendDependsOn(dependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range p.DependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range dependsOn {
		if !existingDeps[dep] {
			p.DependsOn = append(p.DependsOn, dep)
			existingDeps[dep] = true
		}
	}
}

func (p *PipelineStepBase) AppendCredentialDependsOn(credentialDependsOn ...string) {
	existingDeps := make(map[string]bool)
	for _, dep := range p.CredentialDependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range credentialDependsOn {
		if !existingDeps[dep] {
			p.CredentialDependsOn = append(p.CredentialDependsOn, dep)
			existingDeps[dep] = true
		}
	}
}

// Direct copy from Terraform source code
func decodeDependsOn(attr *hcl.Attribute) ([]hcl.Traversal, hcl.Diagnostics) {
	var ret []hcl.Traversal
	exprs, diags := hcl.ExprList(attr.Expr)

	for _, expr := range exprs {
		// expr, shimDiags := shimTraversalInString(expr, false)
		// diags = append(diags, shimDiags...)

		// TODO: should we support legacy "expression in string" syntax here?
		// TODO: terraform supports it by calling shimTraversalInString

		traversal, travDiags := hcl.AbsTraversalForExpr(expr)
		diags = append(diags, travDiags...)
		if len(traversal) != 0 {
			ret = append(ret, traversal)
		}
	}

	return ret, diags
}

func (p *PipelineStepBase) GetMaxConcurrency() *int {
	return p.MaxConcurrency
}

func (p *PipelineStepBase) SetBaseAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	var hclDependsOn []hcl.Traversal
	if attr, exists := hclAttributes[schema.AttributeTypeDependsOn]; exists {
		deps, depsDiags := decodeDependsOn(attr)
		diags = append(diags, depsDiags...)
		hclDependsOn = append(hclDependsOn, deps...)
	}

	if len(diags) > 0 {
		return diags
	}

	var dependsOn []string
	for _, traversal := range hclDependsOn {
		_, addrDiags := addrs.ParseRef(traversal)
		if addrDiags.HasErrors() {
			// We ignore this here, because this isn't a suitable place to return
			// errors. This situation should be caught and rejected during
			// validation.
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  constants.BadDependsOn,
				Detail:   fmt.Sprintf("The depends_on argument must be a reference to another step, but the given value %q is not a valid reference.", traversal),
				Subject:  traversal.SourceRange().Ptr(),
			})
		}
		parts := hclhelpers.TraversalAsStringSlice(traversal)
		if len(parts) < 3 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  constants.BadDependsOn,
				Detail:   "Invalid depends_on format " + strings.Join(parts, "."),
				Subject:  traversal.SourceRange().Ptr(),
			})
			continue
		}

		dependsOn = append(dependsOn, parts[1]+"."+parts[2])
	}

	if attr, exists := hclAttributes[schema.AttributeTypeForEach]; exists {
		p.ForEach = attr.Expr

		do, dgs := hclhelpers.ExpressionToDepends(attr.Expr, ValidDependsOnTypes)
		diags = append(diags, dgs...)
		dependsOn = append(dependsOn, do...)
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTitle]; exists {
		title, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			p.Title = title
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		description, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			p.Description = description
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeMaxConcurrency]; exists {
		maxConcurrency, moreDiags := hclhelpers.AttributeToInt(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			mcInt := int(*maxConcurrency)
			p.MaxConcurrency = &mcInt
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTimeout]; exists {
		val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
		if stepDiags.HasErrors() {
			diags = append(diags, stepDiags...)
		} else if val != cty.NilVal {
			duration, err := hclhelpers.CtyToGo(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse '" + schema.AttributeTypeTimeout + "' attribute to interface",
					Subject:  &attr.Range,
				})
			}
			p.Timeout = duration
		}
	}

	// if attribute is always unresolved, or at least we treat it to be unresolved. Most of the
	// usage will be testing the value that can only be had during the pipeline execution
	if attr, exists := hclAttributes[schema.AttributeTypeIf]; exists {
		// If is always treated as an unresolved attribute
		p.AddUnresolvedAttribute(schema.AttributeTypeIf, attr.Expr)

		do, dgs := hclhelpers.ExpressionToDepends(attr.Expr, ValidDependsOnTypes)
		diags = append(diags, dgs...)
		dependsOn = append(dependsOn, do...)
	}

	p.AppendDependsOn(dependsOn...)

	return diags
}

func (p *PipelineStepBase) GetBaseInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	inputs := map[string]interface{}{}

	if p.UnresolvedAttributes[schema.AttributeTypeTimeout] == nil && p.Timeout != nil {
		inputs[schema.AttributeTypeTimeout] = p.Timeout
	} else if p.UnresolvedAttributes[schema.AttributeTypeTimeout] != nil {

		var timeoutDurationCtyValue cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeTimeout], evalContext, &timeoutDurationCtyValue)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		goVal, err := hclhelpers.CtyToGo(timeoutDurationCtyValue)
		if err != nil {
			return nil, err
		}
		inputs[schema.AttributeTypeTimeout] = goVal
	}

	return inputs, nil
}

func (p *PipelineStepBase) ValidateBaseAttributes() hcl.Diagnostics {

	diags := hcl.Diagnostics{}

	if p.Timeout != nil {
		switch p.Timeout.(type) {
		case string, int:
			// valid duration
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Value of the attribute '" + schema.AttributeTypeTimeout + "' must be a string or a whole number: " + p.GetFullyQualifiedName(),
			})
		}
	}

	return diags
}

func (p *PipelineStepBase) HandleDecodeBodyDiags(diags hcl.Diagnostics, attributeName string, body hcl.Body) hcl.Diagnostics {
	resolvedDiags := 0

	unresolvedDiags := hcl.Diagnostics{}

	for _, e := range diags {
		if e.Severity == hcl.DiagError {
			if e.Detail == `There is no variable named "step".` || e.Detail == `There is no variable named "credential".` {
				traversals := e.Expression.Variables()
				dependsOnAdded := false
				for _, traversal := range traversals {
					parts := hclhelpers.TraversalAsStringSlice(traversal)
					if len(parts) > 0 {
						// When the expression/traversal is referencing an index, the index is also included in the parts
						// for example: []string len: 5, cap: 5, ["step","sleep","sleep_1","0","duration"]
						if parts[0] == schema.BlockTypePipelineStep {
							if len(parts) < 3 {
								return diags
							}
							dependsOn := parts[1] + "." + parts[2]
							p.AppendDependsOn(dependsOn)
							dependsOnAdded = true
						} else if parts[0] == schema.BlockTypeCredential {
							if len(parts) < 2 {
								return diags
							}

							if len(parts) == 2 {
								// dynamic references:
								// step "transform" "aws" {
								// 	value   = credential.aws[param.cred].env
								// }
								dependsOn := parts[1] + ".<dynamic>"
								p.AppendCredentialDependsOn(dependsOn)
								dependsOnAdded = true
							} else {
								dependsOn := parts[1] + "." + parts[2]
								p.AppendCredentialDependsOn(dependsOn)
								dependsOnAdded = true
							}
						}
					}
				}
				if dependsOnAdded {
					resolvedDiags++
				}
			} else if e.Detail == `There is no variable named "result".` && (attributeName == schema.BlockTypeLoop || attributeName == schema.BlockTypeRetry || attributeName == schema.BlockTypeThrow) {
				// result is a reference to the output of the step after it was run, however it should only apply to the loop type block or retry type block
				resolvedDiags++
			} else if e.Detail == `There is no variable named "each".` || e.Detail == `There is no variable named "param".` || e.Detail == "Unsuitable value: value must be known" || e.Detail == `There is no variable named "loop".` || e.Detail == `There is no variable named "retry".` {
				// hcl.decodeBody returns 2 error messages:
				// 1. There's no variable named "param", AND
				// 2. Unsuitable value: value must be known
				resolvedDiags++
			} else {
				unresolvedDiags = append(unresolvedDiags, e)
			}
		}
	}

	// check if all diags have been resolved
	if resolvedDiags == len(diags) {
		if attributeName == schema.BlockTypeThrow {
			return hcl.Diagnostics{}
		} else {
			// * Don't forget to add this, if you change the logic ensure that the code flow still
			// * calls AddUnresolvedBody
			p.AddUnresolvedBody(attributeName, body)
			return hcl.Diagnostics{}
		}
	}

	// There's an error here
	return unresolvedDiags

}

var ValidBaseStepAttributes = []string{
	schema.AttributeTypeTitle,
	schema.AttributeTypeDescription,
	schema.AttributeTypeDependsOn,
	schema.AttributeTypeForEach,
	schema.AttributeTypeIf,
	schema.AttributeTypeTimeout,
	schema.AttributeTypeMaxConcurrency,
}

var ValidDependsOnTypes = []string{
	schema.BlockTypePipelineStep,
}

func (p *PipelineStepBase) IsBaseAttribute(name string) bool {
	return slices.Contains[[]string, string](ValidBaseStepAttributes, name)
}

type PipelineStepHttp struct {
	PipelineStepBase

	Url             *string                `json:"url" binding:"required"`
	Method          *string                `json:"method,omitempty"`
	CaCertPem       *string                `json:"ca_cert_pem,omitempty"`
	Insecure        *bool                  `json:"insecure,omitempty"`
	RequestBody     *string                `json:"request_body,omitempty"`
	RequestHeaders  map[string]interface{} `json:"request_headers,omitempty"`
	BasicAuthConfig *BasicAuthConfig       `json:"basic_auth_config,omitempty"`
}

func (p *PipelineStepHttp) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepHttp)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Compare Url field
	if reflect.DeepEqual(p.Url, other.Url) {
		return false
	}

	// Compare Method field
	if reflect.DeepEqual(p.Method, other.Method) {
		return false
	}

	// Compare Insecure field
	if reflect.DeepEqual(p.Insecure, other.Insecure) {
		return false
	}

	// Compare RequestBody field
	if reflect.DeepEqual(p.RequestBody, other.RequestBody) {
		return false
	}

	// Compare RequestHeaders field using deep equality
	if !reflect.DeepEqual(p.RequestHeaders, other.RequestHeaders) {
		return false
	}

	// All fields are equal
	return true

}

func (p *PipelineStepHttp) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	inputs, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, err
	}

	var urlInput string
	if p.UnresolvedAttributes[schema.AttributeTypeUrl] == nil {
		if p.Url == nil {
			return nil, perr.InternalWithMessage("Url must be supplied")
		}
		urlInput = *p.Url
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeUrl], evalContext, &urlInput)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}
	inputs[schema.AttributeTypeUrl] = urlInput

	if p.UnresolvedAttributes[schema.AttributeTypeMethod] == nil {
		if p.Method != nil {
			inputs[schema.AttributeTypeMethod] = *p.Method
		}
	} else {
		var method string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeMethod], evalContext, &method)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeMethod] = strings.ToLower(method)
	}

	if p.UnresolvedAttributes[schema.AttributeTypeCaCertPem] == nil {
		if p.CaCertPem != nil {
			inputs[schema.AttributeTypeCaCertPem] = *p.CaCertPem
		}
	} else {
		var caCertPem string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCaCertPem], evalContext, &caCertPem)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeCaCertPem] = caCertPem
	}

	if p.UnresolvedAttributes[schema.AttributeTypeInsecure] == nil {
		if p.Insecure != nil {
			inputs[schema.AttributeTypeInsecure] = *p.Insecure
		}
	} else {
		var insecure bool
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeInsecure], evalContext, &insecure)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeInsecure] = insecure
	}

	if p.UnresolvedAttributes[schema.AttributeTypeRequestBody] == nil {
		if p.RequestBody != nil {
			inputs[schema.AttributeTypeRequestBody] = *p.RequestBody
		}
	} else {
		var requestBody string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRequestBody], evalContext, &requestBody)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeRequestBody] = requestBody
	}

	if p.UnresolvedAttributes[schema.AttributeTypeRequestHeaders] == nil {
		if p.RequestHeaders != nil {
			inputs[schema.AttributeTypeRequestHeaders] = p.RequestHeaders
		}
	} else {
		var requestHeaders map[string]string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRequestHeaders], evalContext, &requestHeaders)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeRequestHeaders] = requestHeaders
	}

	if p.BasicAuthConfig != nil {
		basicAuth, diags := p.BasicAuthConfig.GetInputs(evalContext, p.UnresolvedAttributes)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
		basicAuthMap := make(map[string]interface{})
		basicAuthMap["Username"] = basicAuth.Username
		basicAuthMap["Password"] = basicAuth.Password
		inputs[schema.BlockTypePipelineBasicAuth] = basicAuthMap
	}
	inputs[schema.AttributeTypeStepName] = p.Name

	return inputs, nil
}

func (p *PipelineStepHttp) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeUrl:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				urlString, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeUrl + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Url = &urlString
			}

		case schema.AttributeTypeMethod:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				method, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMethod + " attribute to string",
						Subject:  &attr.Range,
					})
				}

				if method != "" {
					if !helpers.StringSliceContains(ValidHttpMethods, strings.ToLower(method)) {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Invalid HTTP method: " + method,
							Subject:  &attr.Range,
						})
						continue
					}
					p.Method = &method
				}
			}
		case schema.AttributeTypeCaCertPem:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				caCertPem, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeCaCertPem + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.CaCertPem = &caCertPem
			}
		case schema.AttributeTypeInsecure:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				if val.Type() != cty.Bool {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid value for insecure attribute",
						Subject:  &attr.Range,
					})
					continue
				}
				insecure := val.True()
				p.Insecure = &insecure
			}

		case schema.AttributeTypeRequestBody:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				requestBody, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeRequestBody + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.RequestBody = &requestBody
			}

		case schema.AttributeTypeRequestHeaders:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)

			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				var err error
				p.RequestHeaders, err = hclhelpers.CtyToGoMapInterface(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse request_headers attribute",
						Subject:  &attr.Range,
					})
					continue
				}
			}
		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for HTTP Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepHttp) SetBlockConfig(blocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.PipelineStepBase.SetBlockConfig(blocks, evalContext)

	basicAuthConfig := &BasicAuthConfig{}

	if basicAuthBlocks := blocks.ByType()[schema.BlockTypePipelineBasicAuth]; len(basicAuthBlocks) > 0 {
		if len(basicAuthBlocks) > 1 {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Multiple basic_auth blocks found for step http",
			}}
		}
		basicAuthBlock := basicAuthBlocks[0]

		var attributes hcl.Attributes

		attributes, diags = basicAuthBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return diags
		}

		if attr, exists := attributes[schema.AttributeTypeUsername]; exists {
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				username, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeUsername + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				basicAuthConfig.Username = username
			}

		}

		if attr, exists := attributes[schema.AttributeTypePassword]; exists {
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				password, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypePassword + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				basicAuthConfig.Password = password
			}

		}
		p.BasicAuthConfig = basicAuthConfig
	}

	return diags
}

func (p *PipelineStepHttp) Validate() hcl.Diagnostics {
	diags := p.ValidateBaseAttributes()
	return diags
}

type PipelineStepSleep struct {
	PipelineStepBase
	Duration interface{} `json:"duration"`
}

func (p *PipelineStepSleep) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepSleep)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	return p.Duration == other.Duration
}

func (p *PipelineStepSleep) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var durationInput interface{}

	if p.UnresolvedAttributes[schema.AttributeTypeDuration] == nil {
		durationInput = p.Duration
	} else {

		var sleepDurationCtyValue cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeDuration], evalContext, &sleepDurationCtyValue)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		goVal, err := hclhelpers.CtyToGo(sleepDurationCtyValue)
		if err != nil {
			return nil, err
		}
		durationInput = goVal
	}

	return map[string]interface{}{
		schema.AttributeTypeDuration: durationInput,
	}, nil
}

func (p *PipelineStepSleep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDuration:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				duration, err := hclhelpers.CtyToGo(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeDuration + "' attribute to interface",
						Subject:  &attr.Range,
					})
				}
				p.Duration = duration
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for " + schema.BlockTypePipelineStepSleep + " Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepSleep) Validate() hcl.Diagnostics {

	diags := hcl.Diagnostics{}

	if p.Duration != nil {
		switch p.Duration.(type) {
		case string, int:
			// valid duration
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Value of the attribute '" + schema.AttributeTypeDuration + "' must be a string or a whole number: " + p.GetFullyQualifiedName(),
			})
		}
	}

	return diags
}

type PipelineStepEmail struct {
	PipelineStepBase
	To           []string `json:"to"`
	From         *string  `json:"from"`
	SmtpPassword *string  `json:"smtp_password"`
	SmtpUsername *string  `json:"smtp_username"`
	Host         *string  `json:"host"`
	Port         *int64   `json:"port"`
	SenderName   *string  `json:"sender_name"`
	Cc           []string `json:"cc"`
	Bcc          []string `json:"bcc"`
	Body         *string  `json:"body"`
	ContentType  *string  `json:"content_type"`
	Subject      *string  `json:"subject"`
}

func (p *PipelineStepEmail) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepEmail)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Use reflect.DeepEqual to compare slices and pointers
	return reflect.DeepEqual(p.To, other.To) &&
		reflect.DeepEqual(p.From, other.From) &&
		reflect.DeepEqual(p.SmtpUsername, other.SmtpUsername) &&
		reflect.DeepEqual(p.SmtpPassword, other.SmtpPassword) &&
		reflect.DeepEqual(p.Host, other.Host) &&
		reflect.DeepEqual(p.Port, other.Port) &&
		reflect.DeepEqual(p.SenderName, other.SenderName) &&
		reflect.DeepEqual(p.Cc, other.Cc) &&
		reflect.DeepEqual(p.Bcc, other.Bcc) &&
		reflect.DeepEqual(p.Body, other.Body) &&
		reflect.DeepEqual(p.ContentType, other.ContentType) &&
		reflect.DeepEqual(p.Subject, other.Subject)

}

func (p *PipelineStepEmail) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var to []string
	if p.UnresolvedAttributes[schema.AttributeTypeTo] == nil {
		to = p.To
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeTo], evalContext, &to)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var from *string
	if p.UnresolvedAttributes[schema.AttributeTypeFrom] == nil {
		from = p.From
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeFrom], evalContext, &from)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var smtpUsername *string
	if p.UnresolvedAttributes[schema.AttributeTypeSmtpUsername] == nil {
		smtpUsername = p.SmtpUsername
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSmtpUsername], evalContext, &smtpUsername)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var smtpPassword *string
	if p.UnresolvedAttributes[schema.AttributeTypeSmtpPassword] == nil {
		smtpPassword = p.SmtpPassword
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSmtpPassword], evalContext, &smtpPassword)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var host *string
	if p.UnresolvedAttributes[schema.AttributeTypeHost] == nil {
		host = p.Host
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeHost], evalContext, &host)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var port *int64
	if p.UnresolvedAttributes[schema.AttributeTypePort] == nil {
		port = p.Port
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypePort], evalContext, &port)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var senderName *string
	if p.UnresolvedAttributes[schema.AttributeTypeSenderName] == nil {
		senderName = p.SenderName
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSenderName], evalContext, &senderName)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var body *string
	if p.UnresolvedAttributes[schema.AttributeTypeBody] == nil {
		body = p.Body
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeBody], evalContext, &body)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var subject *string
	if p.UnresolvedAttributes[schema.AttributeTypeSubject] == nil {
		subject = p.Subject
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSubject], evalContext, &subject)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var contentType *string
	if p.UnresolvedAttributes[schema.AttributeTypeContentType] == nil {
		contentType = p.ContentType
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeContentType], evalContext, &contentType)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var cc []string
	if p.UnresolvedAttributes[schema.AttributeTypeCc] == nil {
		cc = p.Cc
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCc], evalContext, &cc)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var bcc []string
	if p.UnresolvedAttributes[schema.AttributeTypeBcc] == nil {
		bcc = p.Bcc
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeBcc], evalContext, &bcc)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	results := map[string]interface{}{}

	if to != nil {
		results[schema.AttributeTypeTo] = to
	}

	if from != nil {
		results[schema.AttributeTypeFrom] = *from
	}

	if smtpUsername != nil {
		results[schema.AttributeTypeSmtpUsername] = *smtpUsername
	}

	if smtpPassword != nil {
		results[schema.AttributeTypeSmtpPassword] = *smtpPassword
	}

	if host != nil {
		results[schema.AttributeTypeHost] = *host
	}

	if port != nil {
		results[schema.AttributeTypePort] = *port
	}

	if senderName != nil {
		results[schema.AttributeTypeSenderName] = *senderName
	}

	if cc != nil {
		results[schema.AttributeTypeCc] = cc
	}

	if bcc != nil {
		results[schema.AttributeTypeBcc] = bcc
	}

	if body != nil {
		results[schema.AttributeTypeBody] = *body
	}

	if contentType != nil {
		results[schema.AttributeTypeContentType] = *contentType
	}

	if subject != nil {
		results[schema.AttributeTypeSubject] = *subject
	}

	return results, nil
}

func (p *PipelineStepEmail) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeTo:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				emailRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeTo + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.To = emailRecipients
			}

		case schema.AttributeTypeFrom:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				from, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeFrom + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.From = &from
			}

		case schema.AttributeTypeSmtpUsername:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				smtpUsername, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSmtpUsername + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SmtpUsername = &smtpUsername
			}

		case schema.AttributeTypeSmtpPassword:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				smtpPassword, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSmtpPassword + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SmtpPassword = &smtpPassword
			}

		case schema.AttributeTypeHost:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				host, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeHost + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Host = &host
			}

		case schema.AttributeTypePort:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				port, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to convert port into integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Port = port
			}

		case schema.AttributeTypeSenderName:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				senderName, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSenderName + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SenderName = &senderName
			}

		case schema.AttributeTypeCc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				ccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeCc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Cc = ccRecipients
			}

		case schema.AttributeTypeBcc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				bccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBcc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Bcc = bccRecipients
			}

		case schema.AttributeTypeBody:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				body, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBody + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Body = &body
			}

		case schema.AttributeTypeContentType:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				contentType, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeContentType + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.ContentType = &contentType
			}

		case schema.AttributeTypeSubject:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				subject, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSubject + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Subject = &subject
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Email Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

func dependsOnFromExpressions(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface) (cty.Value, hcl.Diagnostics) {
	expr := attr.Expr

	// If there is a param in the expression, then we must assume that we can't resolve it at this stage.
	// If the param has a default, it will be fully resolved and when we change the param, Flowpipe doesn't know that the
	// attribute needs to be recalculated
	for _, traversals := range expr.Variables() {
		if traversals.RootName() == "param" {
			p.AddUnresolvedAttribute(attr.Name, expr)
			// Don't return here because there may be other dependencies to be created below
		}
	}

	// resolve it first if we can
	val, stepDiags := expr.Value(evalContext)
	if stepDiags != nil && stepDiags.HasErrors() {
		resolvedDiags := 0
		for _, e := range stepDiags {
			if e.Severity == hcl.DiagError {
				if e.Detail == `There is no variable named "step".` || e.Detail == `There is no variable named "credential".` {
					traversals := expr.Variables()
					dependsOnAdded := false
					for _, traversal := range traversals {
						parts := hclhelpers.TraversalAsStringSlice(traversal)
						if len(parts) > 0 {
							// When the expression/traversal is referencing an index, the index is also included in the parts
							// for example: []string len: 5, cap: 5, ["step","sleep","sleep_1","0","duration"]
							if parts[0] == schema.BlockTypePipelineStep {
								if len(parts) < 3 {
									return cty.NilVal, stepDiags
								}
								dependsOn := parts[1] + "." + parts[2]
								p.AppendDependsOn(dependsOn)
								dependsOnAdded = true
							} else if parts[0] == schema.BlockTypeCredential {
								if len(parts) < 2 {
									return cty.NilVal, stepDiags
								}

								if len(parts) == 2 {
									// dynamic references:
									// step "transform" "aws" {
									// 	value   = credential.aws[param.cred].env
									// }
									dependsOn := parts[1] + ".<dynamic>"
									p.AppendCredentialDependsOn(dependsOn)
									dependsOnAdded = true
								} else {
									dependsOn := parts[1] + "." + parts[2]
									p.AppendCredentialDependsOn(dependsOn)
									dependsOnAdded = true
								}
							}
						}
					}
					if dependsOnAdded {
						resolvedDiags++
					}
				} else if e.Detail == `There is no variable named "each".` || e.Detail == `There is no variable named "param".` || e.Detail == `There is no variable named "loop".` {
					resolvedDiags++
				} else {
					return cty.NilVal, stepDiags
				}
			}
		}

		// check if all diags have been resolved
		if resolvedDiags == len(stepDiags) {

			// * Don't forget to add this, if you change the logic ensure that the code flow still
			// * calls AddUnresolvedAttribute
			p.AddUnresolvedAttribute(attr.Name, expr)
			return cty.NilVal, hcl.Diagnostics{}
		} else {
			// There's an error here
			return cty.NilVal, stepDiags
		}
	}

	return val, hcl.Diagnostics{}
}

type PipelineStepTransform struct {
	PipelineStepBase
	Value any `json:"value"`
}

func (p *PipelineStepTransform) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepTransform)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	return true
}

func (p *PipelineStepTransform) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var value any

	if p.UnresolvedAttributes[schema.AttributeTypeValue] == nil {
		value = p.Value
	} else {
		var transformValueCtyValue cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeValue], evalContext, &transformValueCtyValue)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		goVal, err := hclhelpers.CtyToGo(transformValueCtyValue)
		if err != nil {
			return nil, err
		}
		value = goVal
	}

	// if test, ok := value.([]any); ok {
	// 	if len(test) == 0 {
	// 		test = make([]any, 0)
	// 		return map[string]interface{}{
	// 			schema.AttributeTypeValue: test,
	// 		}, nil
	// 	}
	// }

	return map[string]interface{}{
		schema.AttributeTypeValue: value,
	}, nil
}

func (p *PipelineStepTransform) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeValue:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				goVal, err := hclhelpers.CtyToGo(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeValue + " attribute to interface",
						Subject:  &attr.Range,
					})
				}

				p.Value = goVal
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Transform Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

type PipelineStepQuery struct {
	PipelineStepBase
	ConnnectionString *string       `json:"connection_string"`
	Sql               *string       `json:"sql"`
	Args              []interface{} `json:"args"`
}

func (p *PipelineStepQuery) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepQuery)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	if len(p.Args) != len(other.Args) {
		return false
	}
	for i := range p.Args {
		if p.Args[i] != other.Args[i] {
			return false
		}
	}

	return reflect.DeepEqual(p.ConnnectionString, other.ConnnectionString) &&
		reflect.DeepEqual(p.Sql, other.Sql)
}

func (p *PipelineStepQuery) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	results, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, err
	}

	var sql *string
	if p.UnresolvedAttributes[schema.AttributeTypeSql] == nil {
		if p.Sql == nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": sql must be supplied")
		}
		sql = p.Sql
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSql], evalContext, &sql)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var connectionString *string
	if p.UnresolvedAttributes[schema.AttributeTypeConnectionString] == nil {
		if p.ConnnectionString == nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": connection string must be supplied")
		}
		connectionString = p.ConnnectionString
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeConnectionString], evalContext, &connectionString)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	if sql != nil {
		results[schema.AttributeTypeSql] = *sql
	}

	if connectionString != nil {
		results[schema.AttributeTypeConnectionString] = *connectionString
	}

	if p.UnresolvedAttributes[schema.AttributeTypeArgs] != nil {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeArgs], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
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

func (p *PipelineStepQuery) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSql:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				sql, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSql + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Sql = &sql
			}
		case schema.AttributeTypeConnectionString:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				connectionString := val.AsString()
				p.ConnnectionString = &connectionString
			}
		case schema.AttributeTypeArgs:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				goVals, err2 := hclhelpers.CtyToGoInterfaceSlice(val)
				if err2 != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeArgs + "' attribute to Go values",
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
					Summary:  "Unsupported attribute for Query Step '" + attr.Name + "'",
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepQuery) Validate() hcl.Diagnostics {
	// validate the base attributes
	diags := p.ValidateBaseAttributes()
	return diags
}

type PipelineStepPipeline struct {
	PipelineStepBase

	Pipeline cty.Value `json:"-"`
	Args     Input     `json:"args"`
}

func (p *PipelineStepPipeline) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
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
		value2, ok := other.Args[key]
		if !ok {
			return false
		}

		// Use reflect.DeepEqual to compare the values
		if !reflect.DeepEqual(value1, value2) {
			return false
		}
	}

	// TODO: more here, can't just compare the name
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
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
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
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
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
			expr := attr.Expr
			if attr.Expr != nil {
				val, err := expr.Value(evalContext)
				if err != nil {
					// For Step's Pipeline reference, all it needs is the pipeline. It can't possibly use the output of a pipeline
					// so if the Pipeline is not parsed (yet) then the error message is:
					// Summary: "Unknown variable"
					// Detail: "There is no variable named \"pipeline\"."
					//
					// Do not unpack the error and create a new "Diagnostic", leave the original error message in
					// and let the "Mod processing" determine if there's an unresolved block
					//
					// There's no "depends_on" from the step to the pipeline, the Flowpipe ES engine does not require it
					diags = append(diags, err...)

					return diags
				}
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

		notifier, err := ctyValueToPipelineStepInputNotifyValueMap(notifierCtyVal)
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
				p.Notifier, err = ctyValueToPipelineStepInputNotifyValueMap(val)
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

	return diags
}

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

	var image *string
	if p.UnresolvedAttributes[schema.AttributeTypeImage] == nil {
		image = p.Image
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeImage], evalContext, &image)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var source *string
	if p.UnresolvedAttributes[schema.AttributeTypeSource] == nil {
		source = p.Source
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSource], evalContext, &source)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var cmd []string
	if p.UnresolvedAttributes[schema.AttributeTypeCmd] == nil {
		cmd = p.Cmd
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCmd], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		var err error
		cmd, err = hclhelpers.CtyToGoStringSlice(args, args.Type())
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse cmd attribute to []string: " + err.Error())
		}
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

	var entryPoint []string
	if p.UnresolvedAttributes[schema.AttributeTypeEntryPoint] == nil {
		entryPoint = p.EntryPoint
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeEntryPoint], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}

		var err error
		entryPoint, err = hclhelpers.CtyToGoStringSlice(args, args.Type())
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse entrypoint attribute to []string: " + err.Error())
		}
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
	results[schema.AttributeTypeCmd] = cmd
	results[schema.AttributeTypeEnv] = env
	results[schema.AttributeTypeEntryPoint] = entryPoint

	if image != nil {
		results[schema.AttributeTypeImage] = *image
	}

	if source != nil {
		results[schema.AttributeTypeSource] = *source
	}

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
