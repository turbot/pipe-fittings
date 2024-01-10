package modconfig

import (
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"

	"github.com/robfig/cron/v3"
)

// The definition of a single Flowpipe Trigger
type Trigger struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	FileName        string `json:"file_name"`
	StartLineNumber int    `json:"start_line_number"`
	EndLineNumber   int    `json:"end_line_number"`

	// 27/09/23 - Args is introduces combination of both parse time and runtime arguments. "var" should be resolved
	// at parse time, the vars all should be supplied when we start the system. However, args can also contain
	// runtime variable, i.e. self.request_body, self.rows
	//
	ArgsRaw hcl.Expression `json:"-"`

	// TODO: 2024/01/09 - change of direction with Trigger schema, pipeline no longer "common" because Query Trigger no longer has a single pipeline for
	// all its events, similarly for HTTP trigger the pipeline is being moved down to the "method" block.
	Pipeline cty.Value     `json:"-"`
	RawBody  hcl.Body      `json:"-" hcl:",remain"`
	Config   TriggerConfig `json:"-"`
}

func (t *Trigger) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	t.FileName = fileName
	t.StartLineNumber = startLineNumber
	t.EndLineNumber = endLineNumber
}

func (t *Trigger) Equals(other *Trigger) bool {
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
}

func (t *Trigger) IsBaseAttribute(name string) bool {
	return slices.Contains[[]string, string](ValidBaseTriggerAttributes, name)
}

func (t *Trigger) SetBaseAttributes(mod *Mod, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	var diags hcl.Diagnostics

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		desc, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Description = desc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTitle]; exists {
		title, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Title = title
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeDocumentation]; exists {
		doc, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Documentation = doc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTags]; exists {
		tags, moreDiags := hclhelpers.AttributeToMap(attr, nil, false)
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
}

type TriggerSchedule struct {
	Schedule string `json:"schedule"`
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

var validIntervals = []string{"hourly", "daily", "weekly", "monthly"}

type TriggerQuery struct {
	Sql              string                          `json:"sql"`
	Schedule         string                          `json:"schedule"`
	ConnectionString string                          `json:"connection_string"`
	PrimaryKey       string                          `json:"primary_key"`
	Captures         map[string]*TriggerQueryCapture `json:"captures"`
}

type TriggerQueryCapture struct {
	Type     string
	Pipeline cty.Value
	ArgsRaw  hcl.Expression
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
		case schema.AttributeTypeConnectionString:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			t.ConnectionString = val.AsString()
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
					Summary:  "Unsupported attribute for Trigger Interval: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	if t.PrimaryKey == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "[development] Primary key is required for Trigger Query",
			Subject:  &hclAttributes[schema.AttributeTypePrimaryKey].Range,
		})
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
	Url           string `json:"url"`
	ExecutionMode string `json:"execution_mode"`
}

var validExecutionMode = []string{"synchronous", "asynchronous"}

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
	return diags
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
