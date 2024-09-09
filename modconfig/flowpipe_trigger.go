package modconfig

import (
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"

	"github.com/robfig/cron/v3"
)

// The definition of a single Flowpipe Trigger
type Trigger struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	mod *Mod

	FileName        string          `json:"file_name"`
	StartLineNumber int             `json:"start_line_number"`
	EndLineNumber   int             `json:"end_line_number"`
	Params          []PipelineParam `json:"params,omitempty"`

	// 27/09/23 - Args is a combination of both parse time and runtime arguments. "var" should be resolved
	// at parse time, the vars all should be supplied when we start the system.
	//
	// However, args can also contain runtime variable, i.e. self.request_body, self.rows
	//
	// Args are not currently validated at parse time. To validate Args at parse time we need to:
	// - identity which args are parse time (var and param)
	// - validate those parse time args
	//
	ArgsRaw hcl.Expression `json:"-"`

	// TODO: 2024/01/09 - change of direction with Trigger schema, pipeline no longer "common" because Query Trigger no longer has a single pipeline for
	// all its events, similarly for HTTP trigger the pipeline is being moved down to the "method" block.
	Pipeline cty.Value     `json:"-"`
	RawBody  hcl.Body      `json:"-" hcl:",remain"`
	Config   TriggerConfig `json:"-"`
	Enabled  *bool         `json:"-"`
}

// Implements the ModTreeItem interface
func (t *Trigger) GetMod() *Mod {
	return t.mod
}

func (t *Trigger) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	t.FileName = fileName
	t.StartLineNumber = startLineNumber
	t.EndLineNumber = endLineNumber
}

func (t *Trigger) GetParam(paramName string) *PipelineParam {
	for _, param := range t.Params {
		if param.Name == paramName {
			return &param
		}
	}
	return nil
}

func (t *Trigger) GetParams() []PipelineParam {
	return t.Params
}

func (t *Trigger) ValidateTriggerParam(params map[string]interface{}, evalCtx *hcl.EvalContext) []error {
	return ValidateParams(t, params, evalCtx)
}

func (p *Trigger) CoerceTriggerParams(params map[string]string, evalCtx *hcl.EvalContext) (map[string]interface{}, []error) {
	return CoerceParams(p, params, evalCtx)
}

func (t *Trigger) Equals(other *Trigger) bool {
	if t == nil && other == nil {
		return true
	}

	if t == nil && other != nil || t != nil && other == nil {
		return false
	}

	baseEqual := t.HclResourceImpl.Equals(&t.HclResourceImpl)
	if !baseEqual {
		return false
	}

	// Order of params does not matter, but the value does
	if len(t.Params) != len(other.Params) {
		return false
	}

	// Compare param values
	for _, v := range t.Params {
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
		pParam := t.GetParam(v.Name)
		if pParam == nil {
			return false
		}
	}

	if !utils.BoolPtrEqual(t.Enabled, other.Enabled) {
		return false
	}

	if t.Pipeline.Equals(other.Pipeline).False() {
		return false
	}

	if !reflect.DeepEqual(t.ArgsRaw, other.ArgsRaw) {
		return false
	}

	if t.Config == nil && !helpers.IsNil(other.Config) || t.Config != nil && helpers.IsNil(other.Config) {
		return false
	}

	if !t.Config.Equals(other.Config) {
		return false
	}

	return t.FullName == other.FullName &&
		t.GetMetadata().ModFullName == other.GetMetadata().ModFullName
}

func (t *Trigger) GetPipeline() cty.Value {
	return t.Pipeline
}

func (t *Trigger) GetArgs(evalContext *hcl.EvalContext) (Input, hcl.Diagnostics) {

	if t.ArgsRaw == nil {
		return Input{}, hcl.Diagnostics{}
	}

	value, diags := t.ArgsRaw.Value(evalContext)

	if diags.HasErrors() {
		return Input{}, diags
	}

	retVal, err := hclhelpers.CtyToGoMapInterface(value)
	if err != nil {
		return Input{}, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + schema.AttributeTypeArgs + " Trigger attribute to Go values",
		}}
	}
	return retVal, diags
}

var ValidBaseTriggerAttributes = []string{
	schema.AttributeTypeDescription,
	schema.AttributeTypePipeline,
	schema.AttributeTypeArgs,
	schema.AttributeTypeTitle,
	schema.AttributeTypeDocumentation,
	schema.AttributeTypeTags,
	schema.AttributeTypeEnabled,
}

func (t *Trigger) IsBaseAttribute(name string) bool {
	return slices.Contains[[]string, string](ValidBaseTriggerAttributes, name)
}

func (t *Trigger) SetBaseAttributes(mod *Mod, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	var diags hcl.Diagnostics

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		desc, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Description = desc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTitle]; exists {
		title, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Title = title
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeDocumentation]; exists {
		doc, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Documentation = doc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTags]; exists {
		tags, moreDiags := hclhelpers.AttributeToMap(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			resultMap := make(map[string]string)
			for key, value := range tags {
				resultMap[key] = value.(string)
			}
			t.Tags = resultMap
		}
	}

	// TODO: this is now only relevant for Schedule Trigger, move it to the Schedule Trigger
	attr := hclAttributes[schema.AttributeTypePipeline]
	if attr != nil {
		expr := attr.Expr

		val, err := expr.Value(evalContext)
		if err != nil {
			// For Trigger's Pipeline reference, all it needs is the pipeline. It can't possibly use the output of a pipeline
			// so if the Pipeline is not parsed (yet) then the error message is:
			// Summary: "Unknown variable"
			// Detail: "There is no variable named \"pipeline\"."
			//
			// Do not unpack the error and create a new "Diagnostic", leave the original error message in
			// and let the "Mod processing" determine if there's an unresolved block

			// Don't error out, it's fine to unable to find the reference, we will try again later
			diags = append(diags, err...)
		} else {
			t.Pipeline = val
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeEnabled]; exists {
		triggerEnabled, moreDiags := hclhelpers.AttributeToBool(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Enabled = triggerEnabled
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
		if attr.Expr != nil {
			t.ArgsRaw = attr.Expr
		}
	}

	return diags
}

type TriggerConfig interface {
	SetAttributes(*Mod, *Trigger, hcl.Attributes, *hcl.EvalContext) hcl.Diagnostics
	SetBlocks(*Mod, *Trigger, hcl.Blocks, *hcl.EvalContext) hcl.Diagnostics
	Equals(other TriggerConfig) bool
	GetType() string
}

type TriggerSchedule struct {
	Schedule string `json:"schedule"`
}

func (t *TriggerSchedule) GetType() string {
	return schema.TriggerTypeSchedule
}

func (t *TriggerSchedule) Equals(other TriggerConfig) bool {
	otherTrigger, ok := other.(*TriggerSchedule)
	if !ok {
		return false
	}

	if t == nil && !helpers.IsNil(otherTrigger) || t != nil && helpers.IsNil(otherTrigger) {
		return false
	}

	if t == nil && helpers.IsNil(otherTrigger) {
		return true
	}

	return t.Schedule == otherTrigger.Schedule
}

func (t *TriggerSchedule) SetAttributes(mod *Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSchedule:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			if val.Type() != cty.String {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "The given schedule is not a string",
					Detail:   "The given schedule is not a string",
					Subject:  &attr.Range,
				})
				continue
			}

			t.Schedule = val.AsString()

			if helpers.StringSliceContains(validIntervals, strings.ToLower(t.Schedule)) {
				continue
			}

			// if it's not an interval, assume it's a cron and attempt to validate the cron expression
			_, err := cron.ParseStandard(t.Schedule)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid cron expression: " + t.Schedule + ". Specify valid intervals hourly, daily, weekly, monthly or valid cron expression",
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
			}
		default:
			if !trigger.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Schedule: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

func (t *TriggerSchedule) SetBlocks(mod *Mod, trigger *Trigger, hclBlocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}

var validIntervals = []string{"hourly", "daily", "weekly", "5m", "10m", "15m", "30m", "60m", "1h", "2h", "4h", "6h", "12h", "24h"}

type TriggerQuery struct {
	Sql        string                          `json:"sql"`
	Schedule   string                          `json:"schedule"`
	Database   string                          `json:"database"`
	PrimaryKey string                          `json:"primary_key"`
	Captures   map[string]*TriggerQueryCapture `json:"captures"`
}

func (t *TriggerQuery) GetType() string {
	return schema.TriggerTypeQuery
}

func (t *TriggerQuery) Equals(other TriggerConfig) bool {
	otherTrigger, ok := other.(*TriggerQuery)
	if !ok {
		return false
	}

	if t == nil && !helpers.IsNil(otherTrigger) || t != nil && helpers.IsNil(otherTrigger) {
		return false
	}

	if t == nil && helpers.IsNil(otherTrigger) {
		return true
	}

	if t.Sql != otherTrigger.Sql {
		return false
	}

	if t.Schedule != otherTrigger.Schedule {
		return false
	}

	if t.Database != otherTrigger.Database {
		return false
	}

	if t.PrimaryKey != otherTrigger.PrimaryKey {
		return false
	}

	if len(t.Captures) != len(otherTrigger.Captures) {
		return false
	}

	for key, value := range t.Captures {
		otherValue, exists := otherTrigger.Captures[key]
		if !exists {
			return false
		}

		if !value.Equals(otherValue) {
			return false
		}
	}

	return true
}

type TriggerQueryCapture struct {
	Type     string
	Pipeline cty.Value
	ArgsRaw  hcl.Expression
}

func (c *TriggerQueryCapture) Equals(other *TriggerQueryCapture) bool {
	if c == nil && other == nil {
		return true
	}

	if c == nil && other != nil || c != nil && other == nil {
		return false
	}

	if c.Type != other.Type {
		return false
	}

	if c.Pipeline.Equals(other.Pipeline).False() {
		return false
	}

	if !reflect.DeepEqual(c.ArgsRaw, other.ArgsRaw) {
		return false
	}

	return true
}

func (t *TriggerQuery) SetAttributes(mod *Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSchedule:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			t.Schedule = val.AsString()

			if slices.Contains(validIntervals, t.Schedule) {
				continue
			}

			// assume it's a cron expression
			_, err := cron.ParseStandard(t.Schedule)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid cron expression: " + t.Schedule,
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
			}
		case schema.AttributeTypeSql:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			t.Sql = val.AsString()
		case schema.AttributeTypeDatabase:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			t.Database = val.AsString()
		case schema.AttributeTypePrimaryKey:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			t.PrimaryKey = val.AsString()

		default:
			if !trigger.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Query: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

var validCaptureBlockTypes = []string{"insert", "update", "delete"}

func (t *TriggerQuery) SetBlocks(mod *Mod, trigger *Trigger, hclBlocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	t.Captures = make(map[string]*TriggerQueryCapture)

	for _, captureBlock := range hclBlocks {

		if len(captureBlock.Labels) != 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid capture block",
				Detail:   "Capture block must have a single label",
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		captureBlockType := captureBlock.Labels[0]
		if !slices.Contains(validCaptureBlockTypes, captureBlockType) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid capture block type",
				Detail:   "Capture block type must be one of: " + strings.Join(validCaptureBlockTypes, ","),
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		hclAttributes, moreDiags := captureBlock.Body.JustAttributes()
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		attr := hclAttributes[schema.AttributeTypePipeline]
		if attr == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Pipeline attribute is required for capture block",
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		if t.Captures[captureBlockType] != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate capture block",
				Detail:   "Duplicate capture block for type: " + captureBlockType,
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		triggerCapture := &TriggerQueryCapture{
			Type: captureBlockType,
		}

		expr := attr.Expr

		val, err := expr.Value(evalContext)
		if err != nil {
			// For Trigger's Pipeline reference, all it needs is the pipeline. It can't possibly use the output of a pipeline
			// so if the Pipeline is not parsed (yet) then the error message is:
			// Summary: "Unknown variable"
			// Detail: "There is no variable named \"pipeline\"."
			//
			// Do not unpack the error and create a new "Diagnostic", leave the original error message in
			// and let the "Mod processing" determine if there's an unresolved block

			// Don't error out, it's fine to unable to find the reference, we will try again later
			diags = append(diags, err...)
		} else {
			triggerCapture.Pipeline = val
		}

		if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
			if attr.Expr != nil {
				triggerCapture.ArgsRaw = attr.Expr
			}
		}

		t.Captures[captureBlockType] = triggerCapture
	}

	return diags
}

func (c *TriggerQueryCapture) GetArgs(evalContext *hcl.EvalContext) (Input, hcl.Diagnostics) {

	if c.ArgsRaw == nil {
		return Input{}, hcl.Diagnostics{}
	}

	value, diags := c.ArgsRaw.Value(evalContext)

	if diags.HasErrors() {
		return Input{}, diags
	}

	retVal, err := hclhelpers.CtyToGoMapInterface(value)
	if err != nil {
		return Input{}, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + schema.AttributeTypeArgs + " Trigger attribute to Go values",
		}}
	}
	return retVal, diags
}

type TriggerHttp struct {
	Url           string                        `json:"url"`
	ExecutionMode string                        `json:"execution_mode"`
	Methods       map[string]*TriggerHTTPMethod `json:"methods"`
}

func (t *TriggerHttp) GetType() string {
	return schema.TriggerTypeHttp
}

func (t *TriggerHttp) Equals(other TriggerConfig) bool {
	otherTrigger, ok := other.(*TriggerHttp)
	if !ok {
		return false
	}

	if t == nil && !helpers.IsNil(otherTrigger) || t != nil && helpers.IsNil(otherTrigger) {
		return false
	}

	if t == nil && helpers.IsNil(otherTrigger) {
		return true
	}

	if t.Url != otherTrigger.Url {
		return false
	}

	if t.ExecutionMode != otherTrigger.ExecutionMode {
		return false
	}

	if len(t.Methods) != len(otherTrigger.Methods) {
		return false
	}

	for key, value := range t.Methods {
		otherValue, exists := otherTrigger.Methods[key]
		if !exists {
			return false
		}

		if !value.Equals(otherValue) {
			return false
		}
	}

	return true
}

type TriggerHTTPMethod struct {
	Type          string
	ExecutionMode string
	Pipeline      cty.Value
	ArgsRaw       hcl.Expression
}

func (c *TriggerHTTPMethod) Equals(other *TriggerHTTPMethod) bool {
	if c == nil && other == nil {
		return true
	}

	if c == nil && other != nil || c != nil && other == nil {
		return false
	}

	if c.Type != other.Type || c.ExecutionMode != other.ExecutionMode {
		return false
	}

	if c.Pipeline.Equals(other.Pipeline).False() {
		return false
	}

	if !reflect.DeepEqual(c.ArgsRaw, other.ArgsRaw) {
		return false
	}

	return true
}

var validExecutionMode = []string{"synchronous", "asynchronous"}
var validMethodBlockTypes = []string{"post", "get"}

func (t *TriggerHttp) SetAttributes(mod *Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeExecutionMode:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			if val.Type() != cty.String {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "The given execution mode is not a string",
					Detail:   "The given execution mode is not a string",
					Subject:  &attr.Range,
				})
				continue
			}

			t.ExecutionMode = val.AsString()

			if !slices.Contains(validExecutionMode, t.ExecutionMode) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid execution mode",
					Detail:   "The execution mode must be one of: " + strings.Join(validExecutionMode, ","),
					Subject:  &attr.Range,
				})
			}

		default:
			if !trigger.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Interval: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (t *TriggerHttp) SetBlocks(mod *Mod, trigger *Trigger, hclBlocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	t.Methods = make(map[string]*TriggerHTTPMethod)

	// If no method blocks appear, only 'post' is supported, and the top-level `pipeline`, `args` and `execution_mode` will be applied
	if len(hclBlocks) == 0 {
		triggerMethod := &TriggerHTTPMethod{
			Type: HttpMethodPost,
		}

		// Get the top-level pipeline
		pipeline := trigger.GetPipeline()
		if pipeline == cty.NilVal {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Bad Request",
				Detail:   "Missing required attribute 'pipeline'",
			})
			return diags
		}
		triggerMethod.Pipeline = pipeline

		// Get the top-level args
		pipelineArgs := trigger.ArgsRaw
		if pipelineArgs != nil {
			triggerMethod.ArgsRaw = pipelineArgs
		}

		// Get the top-level execution_mode
		if t.ExecutionMode != "" {
			triggerMethod.ExecutionMode = t.ExecutionMode
		}

		t.Methods[HttpMethodPost] = triggerMethod

		return diags
	}

	// If the method blocks provided, we will consider the configuration provided in the method block
	for _, methodBlock := range hclBlocks {

		if len(methodBlock.Labels) != 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid method block",
				Detail:   "Method block must have a single label",
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		methodBlockType := methodBlock.Labels[0]
		if !slices.Contains(validMethodBlockTypes, methodBlockType) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid method block type",
				Detail:   "Method block type must be one of: " + strings.Join(validMethodBlockTypes, ","),
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		hclAttributes, moreDiags := methodBlock.Body.JustAttributes()
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		attr := hclAttributes[schema.AttributeTypePipeline]
		if attr == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Pipeline attribute is required for method block",
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		if t.Methods[methodBlockType] != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate method block",
				Detail:   "Duplicate method block for type: " + methodBlockType,
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		triggerMethod := &TriggerHTTPMethod{
			Type: methodBlockType,
		}

		expr := attr.Expr

		val, err := expr.Value(evalContext)
		if err != nil {
			// For Trigger's Pipeline reference, all it needs is the pipeline. It can't possibly use the output of a pipeline
			// so if the Pipeline is not parsed (yet) then the error message is:
			// Summary: "Unknown variable"
			// Detail: "There is no variable named \"pipeline\"."
			//
			// Do not unpack the error and create a new "Diagnostic", leave the original error message in
			// and let the "Mod processing" determine if there's an unresolved block

			// Don't error out, it's fine to unable to find the reference, we will try again later
			diags = append(diags, err...)
		} else {
			triggerMethod.Pipeline = val
		}

		if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
			if attr.Expr != nil {
				triggerMethod.ArgsRaw = attr.Expr
			}
		}

		if attr, exists := hclAttributes[schema.AttributeTypeExecutionMode]; exists {
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
				}

				executionMode, ctyErr := hclhelpers.CtyToString(val)
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeExecutionMode + " attribute to string",
						Subject:  &attr.Range,
					})
				}

				if !slices.Contains(validExecutionMode, executionMode) {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid execution mode",
						Detail:   "The execution mode must be one of: " + strings.Join(validExecutionMode, ","),
						Subject:  &attr.Range,
					})
				}

				triggerMethod.ExecutionMode = executionMode
			}
		}

		t.Methods[methodBlockType] = triggerMethod
	}

	return diags
}

func (c *TriggerHTTPMethod) GetArgs(evalContext *hcl.EvalContext) (Input, hcl.Diagnostics) {

	if c.ArgsRaw == nil {
		return Input{}, hcl.Diagnostics{}
	}

	value, diags := c.ArgsRaw.Value(evalContext)

	if diags.HasErrors() {
		return Input{}, diags
	}

	retVal, err := hclhelpers.CtyToGoMapInterface(value)
	if err != nil {
		return Input{}, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + schema.AttributeTypeArgs + " Trigger attribute to Go values",
		}}
	}
	return retVal, diags
}

func NewTrigger(block *hcl.Block, mod *Mod, triggerType, triggerName string) *Trigger {

	triggerFullName := triggerType + "." + triggerName

	if mod != nil {
		modName := mod.Name()
		if strings.HasPrefix(modName, "mod") {
			modName = strings.TrimPrefix(modName, "mod.")
		}
		triggerFullName = modName + ".trigger." + triggerFullName
	} else {
		triggerFullName = "local.trigger." + triggerFullName
	}

	trigger := &Trigger{
		HclResourceImpl: HclResourceImpl{
			FullName:        triggerFullName,
			UnqualifiedName: "trigger." + triggerName,
			DeclRange:       block.DefRange,
			blockType:       block.Type,
		},
		mod: mod,
	}

	switch triggerType {
	case schema.TriggerTypeSchedule:
		trigger.Config = &TriggerSchedule{}
	case schema.TriggerTypeQuery:
		trigger.Config = &TriggerQuery{}
	case schema.TriggerTypeHttp:
		trigger.Config = &TriggerHttp{}
	default:
		return nil
	}

	return trigger
}

// GetTriggerTypeFromTriggerConfig returns the type of the trigger from the trigger config
func GetTriggerTypeFromTriggerConfig(config TriggerConfig) string {
	switch config.(type) {
	case *TriggerSchedule:
		return schema.TriggerTypeSchedule
	case *TriggerQuery:
		return schema.TriggerTypeQuery
	case *TriggerHttp:
		return schema.TriggerTypeHttp
	}

	return ""
}
