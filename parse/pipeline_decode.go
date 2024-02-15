package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func decodeStep(mod *modconfig.Mod, block *hcl.Block, parseCtx *ModParseContext, pipelineHcl *modconfig.Pipeline) (modconfig.PipelineStep, hcl.Diagnostics) {

	stepType := block.Labels[0]
	stepName := block.Labels[1]

	step := modconfig.NewPipelineStep(stepType, stepName)
	if step == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid pipeline step type " + stepType,
			Subject:  &block.DefRange,
		}}
	}
	step.SetPipelineName(pipelineHcl.FullName)

	pipelineStepBlockSchema := GetPipelineStepBlockSchema(stepType)
	if pipelineStepBlockSchema == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Pipeline step block schema not found for step " + stepType,
			Subject:  &block.DefRange,
		}}
	}

	stepOptions, diags := block.Body.Content(pipelineStepBlockSchema)

	if diags.HasErrors() {
		return step, diags
	}

	moreDiags := step.SetAttributes(stepOptions.Attributes, parseCtx.EvalCtx)
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	moreDiags = step.SetBlockConfig(stepOptions.Blocks, parseCtx.EvalCtx)
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	stepOutput := map[string]*modconfig.PipelineOutput{}

	outputBlocks := stepOptions.Blocks.ByType()[schema.BlockTypePipelineOutput]
	for _, outputBlock := range outputBlocks {
		attributes, moreDiags := outputBlock.Body.JustAttributes()
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
			continue
		}

		if attr, exists := attributes[schema.AttributeTypeValue]; exists {

			o := &modconfig.PipelineOutput{
				Name: outputBlock.Labels[0],
			}

			expr := attr.Expr
			if len(expr.Variables()) > 0 {
				traversals := expr.Variables()
				for _, traversal := range traversals {
					parts := hclhelpers.TraversalAsStringSlice(traversal)
					if len(parts) > 0 {
						if parts[0] == schema.BlockTypePipelineStep {
							dependsOn := parts[1] + "." + parts[2]
							step.AppendDependsOn(dependsOn)
						}
					}
				}
				o.UnresolvedValue = attr.Expr
			} else {
				ctyVal, _ := attr.Expr.Value(nil)
				val, _ := hclhelpers.CtyToGo(ctyVal)
				o.Value = val
			}

			stepOutput[o.Name] = o
		} else {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Missing value attribute",
				Subject:  &block.DefRange,
			})
		}

	}
	step.SetOutputConfig(stepOutput)

	if len(diags) == 0 {
		moreDiags := step.Validate()
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
		}
	}

	return step, diags
}

func decodePipelineParam(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.PipelineParam, hcl.Diagnostics) {

	o := &modconfig.PipelineParam{
		Name: block.Labels[0],
	}

	paramOptions, diags := block.Body.Content(modconfig.PipelineParamBlockSchema)

	if diags.HasErrors() {
		return o, diags
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeType]; exists {
		expr := attr.Expr
		// First we'll deal with some shorthand forms that the HCL-level type
		// expression parser doesn't include. These both emulate pre-0.12 behavior
		// of allowing a list or map of any element type as long as all of the
		// elements are consistent. This is the same as list(any) or map(any).
		switch hcl.ExprAsKeyword(expr) {
		case "list":
			o.Type = cty.List(cty.DynamicPseudoType)
		case "map":
			o.Type = cty.Map(cty.DynamicPseudoType)
		case "set":
			o.Type = cty.Set(cty.DynamicPseudoType)
		default:
			ty, moreDiags := typeexpr.TypeConstraint(expr)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				return o, diags
			}

			o.Type = ty
		}
	} else {
		o.Type = cty.DynamicPseudoType
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeOptional]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Optional)
		diags = append(diags, valDiags...)
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeDefault]; exists {
		ctyVal, moreDiags := attr.Expr.Value(parseCtx.EvalCtx)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			return o, diags
		}

		o.Default = ctyVal
	} else if o.Optional {
		o.Default = cty.NullVal(o.Type)
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeDescription]; exists {
		ctyVal, moreDiags := attr.Expr.Value(parseCtx.EvalCtx)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			return o, diags
		}

		o.Description = ctyVal.AsString()
	}

	return o, diags
}

func decodeOutput(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.PipelineOutput, hcl.Diagnostics) {

	o := &modconfig.PipelineOutput{
		Name: block.Labels[0],
	}

	outputOptions, diags := block.Body.Content(modconfig.PipelineOutputBlockSchema)

	if diags.HasErrors() {
		return o, diags
	}

	if attr, exists := outputOptions.Attributes[schema.AttributeTypeDescription]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Description)
		diags = append(diags, valDiags...)
	}

	if attr, exists := outputOptions.Attributes[schema.AttributeTypeValue]; exists {
		expr := attr.Expr
		if len(expr.Variables()) > 0 {
			traversals := expr.Variables()
			for _, traversal := range traversals {
				parts := hclhelpers.TraversalAsStringSlice(traversal)
				if len(parts) > 0 {
					if parts[0] == schema.BlockTypePipelineStep {
						dependsOn := parts[1] + "." + parts[2]
						o.AppendDependsOn(dependsOn)
					}
				}
			}
		}
		o.UnresolvedValue = attr.Expr

	} else {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing value attribute",
			Subject:  &block.DefRange,
		})
	}

	return o, diags
}

func decodeTrigger(mod *modconfig.Mod, block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Trigger, *DecodeResult) {

	res := newDecodeResult()

	if len(block.Labels) != 2 {
		res.handleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid trigger block - expected 2 labels, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		})
		return nil, res
	}

	triggerType := block.Labels[0]
	triggerName := block.Labels[1]

	triggerHcl := modconfig.NewTrigger(block, mod, triggerType, triggerName)

	triggerSchema := GetTriggerBlockSchema(triggerType)
	if triggerSchema == nil {
		res.handleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "invalid trigger type: " + triggerType,
				Subject:  &block.DefRange,
			},
		})
		return triggerHcl, res
	}

	triggerOptions, diags := block.Body.Content(triggerSchema)

	if diags.HasErrors() {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	if triggerHcl == nil {
		res.handleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid trigger type '%s'", triggerType),
				Subject:  &block.DefRange,
			},
		})
		return nil, res
	}

	diags = triggerHcl.Config.SetAttributes(mod, triggerHcl, triggerOptions.Attributes, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
		return triggerHcl, res
	}

	diags = triggerHcl.Config.SetBlocks(mod, triggerHcl, triggerOptions.Blocks, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
		return triggerHcl, res
	}

	body, ok := block.Body.(*hclsyntax.Body)
	if ok {
		triggerHcl.SetFileReference(block.DefRange.Filename, body.SrcRange.Start.Line, body.EndRange.Start.Line)
	} else {
		// This shouldn't happen, but if it does, try our best effort to set the file reference. It will get the start line correctly
		// but not the end line
		triggerHcl.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)
	}

	moreDiags := parseCtx.AddTrigger(triggerHcl)
	res.addDiags(moreDiags)

	return triggerHcl, res
}

// TODO: validation - if you specify invalid depends_on it doesn't error out
// TODO: validation - invalid name?
func decodePipeline(mod *modconfig.Mod, block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Pipeline, *DecodeResult) {
	res := newDecodeResult()

	// get shell pipelineHcl
	pipelineHcl := modconfig.NewPipeline(mod, block)

	pipelineOptions, diags := block.Body.Content(modconfig.PipelineBlockSchema)
	if diags.HasErrors() {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	diags = pipelineHcl.SetAttributes(pipelineOptions.Attributes, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	// use a map keyed by a string for fast lookup
	// we use an empty struct as the value type, so that
	// we don't use up unnecessary memory
	// foundOptions := map[string]struct{}{}
	for _, block := range pipelineOptions.Blocks {
		switch block.Type {
		case schema.BlockTypePipelineStep:
			step, diags := decodeStep(mod, block, parseCtx, pipelineHcl)
			if diags.HasErrors() {
				res.handleDecodeDiags(diags)

				// Must also return the pipelineHcl even if it failed parsing, because later on the handling of "unresolved blocks" expect
				// the resource to be there
				return pipelineHcl, res
			}

			body, ok := block.Body.(*hclsyntax.Body)
			if ok {
				step.SetFileReference(block.DefRange.Filename, body.SrcRange.Start.Line, body.EndRange.Start.Line)
			} else {
				// This shouldn't happen, but if it does, try our best effort to set the file reference. It will get the start line correctly
				// but not the end line
				step.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)
			}
			pipelineHcl.Steps = append(pipelineHcl.Steps, step)

		case schema.BlockTypePipelineOutput:
			output, cfgDiags := decodeOutput(block, parseCtx)
			diags = append(diags, cfgDiags...)
			if len(diags) > 0 {
				res.handleDecodeDiags(diags)
				return pipelineHcl, res
			}

			if output != nil {

				// check for duplicate output names
				if len(pipelineHcl.OutputConfig) > 0 {
					for _, o := range pipelineHcl.OutputConfig {
						if o.Name == output.Name {
							diags = append(diags, &hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  fmt.Sprintf("duplicate output name '%s' - output names must be unique", output.Name),
								Subject:  &block.DefRange,
							})
							res.handleDecodeDiags(diags)
							return pipelineHcl, res
						}
					}
				}

				pipelineHcl.OutputConfig = append(pipelineHcl.OutputConfig, *output)
			}

		case schema.BlockTypeParam:
			pipelineParam, moreDiags := decodePipelineParam(block, parseCtx)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				res.handleDecodeDiags(diags)
				return pipelineHcl, res
			}

			if pipelineParam != nil {

				// check for duplicate pipeline parameters names
				if len(pipelineHcl.Params) > 0 {
					if _, ok := pipelineHcl.Params[pipelineParam.Name]; ok {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  fmt.Sprintf("duplicate pipeline parameter name '%s' - parameter names must be unique", pipelineParam.Name),
							Subject:  &block.DefRange,
						})
						res.handleDecodeDiags(diags)
						return pipelineHcl, res
					}
				}

				pipelineHcl.Params[pipelineParam.Name] = pipelineParam
			}

		default:
			// this should never happen
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for workspace profiles", block.Type),
				Subject:  &block.DefRange,
			})
		}
	}

	handlePipelineDecodeResult(pipelineHcl, res, block, parseCtx)
	diags = validatePipelineDependencies(pipelineHcl, parseCtx.Credentials)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)

		// Must also return the pipelineHcl even if it failed parsing, because later on the handling of "unresolved blocks" expect
		// the resource to be there
		return pipelineHcl, res
	}

	body, ok := block.Body.(*hclsyntax.Body)
	if ok {
		pipelineHcl.SetFileReference(block.DefRange.Filename, body.SrcRange.Start.Line, body.EndRange.Start.Line)
	} else {
		// This shouldn't happen, but if it does, try our best effort to set the file reference. It will get the start line correctly
		// but not the end line
		pipelineHcl.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)
	}

	return pipelineHcl, res
}

func validatePipelineDependencies(pipelineHcl *modconfig.Pipeline, credentials map[string]credential.Credential) hcl.Diagnostics {
	var diags hcl.Diagnostics

	var stepRegisters []string
	for _, step := range pipelineHcl.Steps {
		stepRegisters = append(stepRegisters, step.GetFullyQualifiedName())
	}

	var credentialRegisters []string
	availableCredentialTypes := map[string]bool{}
	for k := range credentials {
		parts := strings.Split(k, ".")
		if len(parts) != 2 {
			continue
		}

		// Add the credential to the register
		credentialRegisters = append(credentialRegisters, k)

		// List out the supported credential types
		availableCredentialTypes[parts[0]] = true
	}

	var credentialTypes []string
	for k := range availableCredentialTypes {
		credentialTypes = append(credentialTypes, k)
	}

	for _, step := range pipelineHcl.Steps {
		dependsOn := step.GetDependsOn()

		for _, dep := range dependsOn {
			if !helpers.StringSliceContains(stepRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s' - step '%s' does not exist for pipeline %s", dep, step.GetFullyQualifiedName(), pipelineHcl.Name()),
					Detail:   fmt.Sprintf("valid steps are: %s", strings.Join(stepRegisters, ", ")),
				})
			}
		}

		credentialDependsOn := step.GetCredentialDependsOn()
		for _, dep := range credentialDependsOn {
			// Check if the credential type is supported, if <dynamic>
			parts := strings.Split(dep, ".")
			if len(parts) != 2 {
				continue
			}

			if parts[1] == "<dynamic>" {
				if !availableCredentialTypes[parts[0]] {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("invalid depends_on '%s' - credential type '%s' not supported for pipeline %s", dep, parts[0], pipelineHcl.Name()),
						Detail:   fmt.Sprintf("valid credential types are: %s", strings.Join(credentialTypes, ", ")),
					})
				}
				continue
			}

			if !helpers.StringSliceContains(credentialRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s' - credential does not exist for pipeline %s", dep, pipelineHcl.Name()),
					Detail:   fmt.Sprintf("valid credentials are: %s", strings.Join(credentialRegisters, ", ")),
				})
			}
		}
	}

	for _, outputConfig := range pipelineHcl.OutputConfig {
		dependsOn := outputConfig.DependsOn

		for _, dep := range dependsOn {
			if !helpers.StringSliceContains(stepRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s' - does not exist for pipeline %s", dep, pipelineHcl.Name()),
				})
			}
		}
	}

	return diags
}

func handlePipelineDecodeResult(resource *modconfig.Pipeline, res *DecodeResult, block *hcl.Block, parseCtx *ModParseContext) {
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded(block, parseCtx)
		res.addDiags(moreDiags)

		moreDiags = parseCtx.AddPipeline(resource)
		res.addDiags(moreDiags)
		return
	}

	// failure :(
	if len(res.Depends) > 0 {
		moreDiags := parseCtx.AddDependencies(block, resource.Name(), res.Depends)
		res.addDiags(moreDiags)
	}
}

func GetPipelineStepBlockSchema(stepType string) *hcl.BodySchema {
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		return modconfig.PipelineStepHttpBlockSchema
	case schema.BlockTypePipelineStepSleep:
		return modconfig.PipelineStepSleepBlockSchema
	case schema.BlockTypePipelineStepEmail:
		return modconfig.PipelineStepEmailBlockSchema
	case schema.BlockTypePipelineStepTransform:
		return modconfig.PipelineStepTransformBlockSchema
	case schema.BlockTypePipelineStepQuery:
		return modconfig.PipelineStepQueryBlockSchema
	case schema.BlockTypePipelineStepPipeline:
		return modconfig.PipelineStepPipelineBlockSchema
	case schema.BlockTypePipelineStepFunction:
		return modconfig.PipelineStepFunctionBlockSchema
	case schema.BlockTypePipelineStepContainer:
		return modconfig.PipelineStepContainerBlockSchema
	case schema.BlockTypePipelineStepInput:
		return modconfig.PipelineStepInputBlockSchema
	default:
		return nil
	}
}

func GetTriggerBlockSchema(triggerType string) *hcl.BodySchema {
	switch triggerType {
	case schema.TriggerTypeSchedule:
		return modconfig.TriggerScheduleBlockSchema
	case schema.TriggerTypeQuery:
		return modconfig.TriggerQueryBlockSchema
	case schema.TriggerTypeHttp:
		return modconfig.TriggerHttpBlockSchema
	default:
		return nil
	}
}

func GetIntegrationBlockSchema(integrationType string) *hcl.BodySchema {
	switch integrationType {
	case schema.IntegrationTypeSlack:
		return modconfig.IntegrationSlackBlockSchema
	case schema.IntegrationTypeEmail:
		return modconfig.IntegrationEmailBlockSchema
	case schema.IntegrationTypeTeams:
		return modconfig.IntegrationTeamsBlockSchema
	default:
		return nil
	}
}
