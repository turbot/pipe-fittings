package modconfig

import (
	"fmt"
	"reflect"
	"slices"
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
	case schema.BlockTypePipelineStepMessage:
		step = &PipelineStepMessage{}
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
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
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

func stringSliceInputFromAttribute(p PipelineStep, results map[string]interface{}, evalContext *hcl.EvalContext, attributeName, fieldName string) (map[string]interface{}, hcl.Diagnostics) {
	var tempValue []string

	unresolvedAttrib := p.GetUnresolvedAttributes()[attributeName]

	if unresolvedAttrib == nil {
		val := reflect.ValueOf(p)
		if val.Kind() == reflect.Ptr {
			val = val.Elem() // If a pointer to a struct is passed, get the struct
		}

		field := val.FieldByName(fieldName)

		if !field.IsValid() {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No such field: " + fieldName + " in obj for " + p.GetFullyQualifiedName(),
				},
			}
		}

		if !helpers.IsNil(field.Interface()) {
			tempValue = field.Interface().([]string)
		}
	} else {
		var args cty.Value

		diags := gohcl.DecodeExpression(unresolvedAttrib, evalContext, &args)
		if diags.HasErrors() {
			return nil, diags
		}

		var err error
		tempValue, err = hclhelpers.CtyToGoStringSlice(args, args.Type())
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + attributeName + " attribute to string",
					Subject:  unresolvedAttrib.Range().Ptr(),
				},
			}
		}
	}

	if tempValue != nil {
		results[attributeName] = tempValue
	}

	return results, hcl.Diagnostics{}
}

func simpleTypeInputFromAttribute[T any](p PipelineStep, results map[string]interface{}, evalContext *hcl.EvalContext, attributeName string, fieldValue T) (map[string]interface{}, hcl.Diagnostics) {
	var tempValue T

	if p.GetUnresolvedAttributes()[attributeName] == nil {
		if !helpers.IsNil(fieldValue) {
			tempValue = fieldValue
		}
	} else {
		diags := gohcl.DecodeExpression(p.GetUnresolvedAttributes()[attributeName], evalContext, &tempValue)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	if !helpers.IsNil(tempValue) {
		if utils.IsPointer(tempValue) {
			// Reflect on tempValue to get its underlying value if it's a pointer
			valueOfTempValue := reflect.ValueOf(tempValue)
			if valueOfTempValue.Kind() == reflect.Ptr && !valueOfTempValue.IsNil() {
				// Dereference the pointer and set the result in the map
				results[attributeName] = valueOfTempValue.Elem().Interface()
			} else {
				results[attributeName] = tempValue
			}
		} else {
			results[attributeName] = tempValue
		}
	}

	return results, hcl.Diagnostics{}
}

// setField sets the field of a struct pointed to by v to the given value.
// v must be a pointer to a struct, fieldName must be the name of a field in the struct,
// and value must be assignable to the field.
func setField(v interface{}, fieldName string, value interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return perr.BadRequestWithMessage("v must be a pointer to a struct")
	}

	rv = rv.Elem() // Dereference the pointer to get the struct

	field := rv.FieldByName(fieldName)
	if !field.IsValid() {
		return perr.BadRequestWithMessage(fmt.Sprintf("no such field: %s in obj", fieldName))
	}

	if !field.CanSet() {
		return perr.BadRequestWithMessage(fmt.Sprintf("cannot set field %s", fieldName))
	}

	fieldValue := reflect.ValueOf(value)
	if field.Type() != fieldValue.Type() {
		return perr.BadRequestWithMessage("provided value type does not match field type")
	}

	field.Set(fieldValue)
	return nil
}

func setStringSliceAttribute(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool) hcl.Diagnostics {
	val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
	if stepDiags.HasErrors() {
		return stepDiags
	}

	if val == cty.NilVal {
		return hcl.Diagnostics{}
	}

	t, err := hclhelpers.CtyToGoStringSlice(val, val.Type())
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + " attribute to string",
				Subject:  &attr.Range,
			},
		}
	}

	if isPtr {
		err = setField(p, fieldName, &t)
	} else {
		err = setField(p, fieldName, t)
	}

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to set " + attr.Name + " attribute to struct",
				Subject:  &attr.Range,
			},
		}
	}

	return hcl.Diagnostics{}
}

func setStringAttribute(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool) hcl.Diagnostics {
	val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
	if stepDiags.HasErrors() {
		return stepDiags
	}

	if val == cty.NilVal {
		return hcl.Diagnostics{}
	}

	t, err := hclhelpers.CtyToString(val)
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + " attribute to string",
				Subject:  &attr.Range,
			},
		}
	}

	if isPtr {
		err = setField(p, fieldName, &t)
	} else {
		err = setField(p, fieldName, t)
	}

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to set " + attr.Name + " attribute to struct",
				Subject:  &attr.Range,
			},
		}
	}

	return hcl.Diagnostics{}
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

			// special handling if the attribute name is "pipeline"
			//
			// this is to handle the pipeline step:
			/**

			step "pipeline" "run_pipeline {
				pipeline = pipeline[param.name]
			}

			we short circuit it straight away and return. It will be resolved at runtime. We can't do that for other attributes because we
			may do something like:

			value = "${param.foo} and ${step.transform.name.value}"

			so the above has dependency on param.foo AND the step.transform.name. We *need* to add `step.transform.name` to the depends_on list
			so it can't return here

			pipeline attribute is special that it can only reference another pipeline
			*/

			if attr.Name == "pipeline" {
				return cty.NilVal, hcl.Diagnostics{}
			}
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
