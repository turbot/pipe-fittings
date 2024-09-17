package parse

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
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

	step := modconfig.NewPipelineStep(stepType, stepName, pipelineHcl)
	if step == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid pipeline step type " + stepType,
			Subject:  &block.DefRange,
		}}
	}
	step.SetPipelineName(pipelineHcl.FullName)
	step.SetRange(block.DefRange.Ptr())

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
						} else if parts[0] == schema.BlockTypeCredential {

							if len(parts) == 2 {
								// dynamic references:
								// step "transform" "aws" {
								// 	value   = credential.aws[param.cred].env
								// }
								dependsOn := parts[1] + ".<dynamic>"
								step.AppendCredentialDependsOn(dependsOn)
							} else {
								dependsOn := parts[1] + "." + parts[2]
								step.AppendCredentialDependsOn(dependsOn)
							}
						} else if parts[0] == schema.BlockTypeConnection {
							if len(parts) == 2 {
								// dynamic references:
								// step "transform" "aws" {
								// 	value   = credential.aws[param.cred].env
								// }
								dependsOn := parts[1] + ".<dynamic>"
								step.AppendConnectionDependsOn(dependsOn)
							} else {
								dependsOn := parts[1] + "." + parts[2]
								step.AppendConnectionDependsOn(dependsOn)
							}
						} else {
							dependsOn := parts[0]
							step.AppendConnectionDependsOn(dependsOn)
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

var ConnectionCtyType = cty.Capsule("ConnectionCtyType", reflect.TypeOf(&modconfig.ConnectionImpl{}))
var NotifierCtyType = cty.Capsule("NotifierCtyType", reflect.TypeOf(&modconfig.NotifierImpl{}))

func handleScopeTraversalExpr(expr *hclsyntax.ScopeTraversalExpr) (cty.Type, bool) {
	dottedString := hclhelpers.TraversalAsString(expr.Traversal)
	parts := strings.Split(dottedString, ".")
	if len(parts) == 2 {
		switch parts[0] {
		case schema.BlockTypeConnection:
			ty := modconfig.ConnectionCtyType(parts[1])
			return ty, ty == cty.NilType
		case schema.BlockTypeNotifier:
			return NotifierCtyType, false
		}
	} else if len(parts) == 1 {
		switch parts[0] {
		case schema.BlockTypeConnection:
			return ConnectionCtyType, false
		case schema.BlockTypeNotifier:
			return NotifierCtyType, false
		}
	}
	return cty.NilType, true
}

func handleFunctionCallExpr(fCallExpr *hclsyntax.FunctionCallExpr) (cty.Type, bool) {
	if fCallExpr.Name == "list" && len(fCallExpr.Args) == 1 {
		dottedString := hclhelpers.TraversalAsString(fCallExpr.Args[0].Variables()[0])
		parts := strings.Split(dottedString, ".")
		if len(parts) == 2 {
			switch parts[0] {
			case schema.BlockTypeConnection:
				innerTy := modconfig.ConnectionCtyType(parts[1])
				return cty.List(innerTy), innerTy == cty.NilType
			case schema.BlockTypeNotifier:
				return cty.List(NotifierCtyType), false
			}
		} else if len(parts) == 1 {
			switch parts[0] {
			case schema.BlockTypeConnection:
				return cty.List(ConnectionCtyType), false
			case schema.BlockTypeNotifier:
				return cty.List(NotifierCtyType), false
			}
		}
	}
	return cty.NilType, true
}

func extractExpressionString(expr hcl.Expression, src string) string {
	rng := expr.Range()
	return src[rng.Start.Byte:rng.End.Byte]
}

// Checks if the given type is in the allowed list
func containsType(allowedTypes []cty.Type, typ cty.Type) bool {
	for _, t := range allowedTypes {
		if t == typ {
			return true
		}
	}
	return false
}

// Creates an HCL error diagnostic
func createErrorDiagnostic(summary string, subject *hcl.Range) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  summary,
		Subject:  subject,
	}
}

func decodePipelineParam(src string, block *hcl.Block, parseCtx *ModParseContext) (*modconfig.PipelineParam, hcl.Diagnostics) {

	o := &modconfig.PipelineParam{
		Name: block.Labels[0],
	}

	evalCtx := parseCtx.EvalCtx

	// because we want to use late binding for temp creds *and* the ability for pipeline param to define custom type,
	// we do the validation where with a list of temporary connections
	if len(parseCtx.PipelingConnections) > 0 {
		connMap, err := BuildTemporaryConnectionMapForEvalContext(context.TODO(), parseCtx.PipelingConnections)
		if err != nil {
			slog.Warn("failed to build temporary connection map for eval context", "error", err)
			return o, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "failed to build temporary connection map for eval context",
					Subject:  &block.DefRange,
				},
			}
		}

		vars := evalCtx.Variables
		vars[schema.BlockTypeConnection] = cty.ObjectVal(connMap)
		evalCtx.Variables = vars

		defer func() {
			vars := evalCtx.Variables
			delete(vars, schema.BlockTypeConnection)
			evalCtx.Variables = vars
		}()
	}

	if len(parseCtx.Notifiers) > 0 {
		notifierMap, err := BuildNotifierMapForEvalContext(parseCtx.Notifiers)
		if err != nil {
			slog.Warn("failed to build notifier map for eval context", "error", err)
			return o, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "failed to build notifier map for eval context",
					Subject:  &block.DefRange,
				},
			}
		}

		vars := evalCtx.Variables
		vars[schema.BlockTypeNotifier] = cty.ObjectVal(notifierMap)
		evalCtx.Variables = vars
	}

	paramOptions, diags := block.Body.Content(modconfig.PipelineParamBlockSchema)

	if diags.HasErrors() {
		return o, diags
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeType]; exists {
		expr := attr.Expr
		ty, moreDiags := typeexpr.TypeConstraint(expr)

		typeErr := moreDiags.HasErrors()

		if typeErr {
			// Handle shorthand forms for list, map, and set

			switch hcl.ExprAsKeyword(expr) {
			case "list":
				ty = cty.List(cty.DynamicPseudoType)
				typeErr = false
			case "map":
				ty = cty.Map(cty.DynamicPseudoType)
				typeErr = false
			case "set":
				ty = cty.Set(cty.DynamicPseudoType)
				typeErr = false
			default:
				if scopeTraversalExpr, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
					ty, typeErr = handleScopeTraversalExpr(scopeTraversalExpr)
				} else if fCallExpr, ok := expr.(*hclsyntax.FunctionCallExpr); ok {
					ty, typeErr = handleFunctionCallExpr(fCallExpr)
				}
			}

			if typeErr {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "A type specification is either a primitive type keyword (bool, number, string), complex type constructor call or Turbot custom type (connection, notifier)",
					Subject:  &attr.Range,
				})
				return o, diags
			}
		}

		o.Type = ty
		o.TypeString = extractExpressionString(expr, src)
	} else {
		o.Type = cty.DynamicPseudoType
		o.TypeString = "any"
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

		// Does the default value matches the specified type?
		if o.Type != cty.DynamicPseudoType && ctyVal.Type() != o.Type {
			if o.IsCustomType() {
				ctdiags := modconfig.CustomTypeValidation(attr, ctyVal, o.Type)
				if len(ctdiags) > 0 {
					diags = append(diags, ctdiags...)
					return o, diags
				}
				o.Default = ctyVal
			} else {
				isCompatible := hclhelpers.IsValueCompatibleWithType(o.Type, ctyVal)
				if !isCompatible {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("default value type mismatched - expected %s, got %s", o.Type.FriendlyName(), ctyVal.Type().FriendlyName()),
						Subject:  &attr.Range,
					})
				} else {
					o.Default = ctyVal
				}
			}
		} else {
			o.Default = ctyVal
		}
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

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeEnum]; exists {
		allowedTypes := []cty.Type{
			cty.String, cty.Bool, cty.Number,
			cty.List(cty.String), cty.List(cty.Bool), cty.List(cty.Number),
		}

		if !containsType(allowedTypes, o.Type) {
			return o, append(diags, createErrorDiagnostic("enum is only supported for string, bool, number, list of string, list of bool, list of number types", &attr.Range))
		}

		ctyVal, moreDiags := attr.Expr.Value(parseCtx.EvalCtx)
		if len(moreDiags) > 0 {
			return o, append(diags, moreDiags...)
		}

		if !hclhelpers.IsCollectionOrTuple(ctyVal.Type()) {
			return o, append(diags, createErrorDiagnostic("enum values must be a list", &attr.Range))
		}

		if !hclhelpers.IsEnumValueCompatibleWithType(o.Type, ctyVal) {
			return o, append(diags, createErrorDiagnostic("enum values type mismatched", &attr.Range))
		}

		if o.Default != cty.NilVal {
			if !hclhelpers.IsEnumValueCompatibleWithType(o.Default.Type(), ctyVal) {
				return o, append(diags, createErrorDiagnostic("param default value type mismatched with enum", &attr.Range))
			}
			if valid, err := hclhelpers.ValidateSettingWithEnum(o.Default, ctyVal); err != nil || !valid {
				return o, append(diags, createErrorDiagnostic("default value not in enum or error validating", &attr.Range))
			}
		}

		o.Enum = ctyVal

		enumGo, err := hclhelpers.CtyToGo(o.Enum)
		if err != nil {
			return o, append(diags, createErrorDiagnostic("error converting enum to go", &attr.Range))
		}

		enumGoSlice, ok := enumGo.([]any)
		if !ok {
			return o, append(diags, createErrorDiagnostic("enum is not a slice", &attr.Range))
		}

		o.EnumGo = enumGoSlice
	}

	if _, exists := paramOptions.Attributes[schema.AttributeTypeTags]; exists {
		valDiags := decodeProperty(paramOptions, "tags", &o.Tags, parseCtx.EvalCtx)
		diags = append(diags, valDiags...)
	}

	return o, diags
}

func decodeOutput(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.PipelineOutput, hcl.Diagnostics) {

	o := &modconfig.PipelineOutput{
		Name:  block.Labels[0],
		Range: block.DefRange.Ptr(),
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
					} else if parts[0] == schema.BlockTypeCredential {

						if len(parts) == 2 {
							// dynamic references:
							// step "transform" "aws" {
							// 	value   = credential.aws[param.cred].env
							// }
							dependsOn := parts[1] + ".<dynamic>"
							o.AppendCredentialDependsOn(dependsOn)
						} else {
							dependsOn := parts[1] + "." + parts[2]
							o.AppendCredentialDependsOn(dependsOn)
						}
					} else if parts[0] == schema.BlockTypeConnection {
						dependsOn := parts[1] + "." + parts[2]
						o.AppendConnectionDependsOn(dependsOn)
					} else {
						dependsOn := parts[0]
						o.AppendConnectionDependsOn(dependsOn)
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

	res := NewDecodeResult()

	if len(block.Labels) != 2 {
		res.HandleDecodeDiags(hcl.Diagnostics{
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
		res.HandleDecodeDiags(hcl.Diagnostics{
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
		res.HandleDecodeDiags(diags)
		return nil, res
	}

	if triggerHcl == nil {
		res.HandleDecodeDiags(hcl.Diagnostics{
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
		res.HandleDecodeDiags(diags)
		return triggerHcl, res
	}

	diags = triggerHcl.Config.SetBlocks(mod, triggerHcl, triggerOptions.Blocks, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
		return triggerHcl, res
	}

	// Read the entire file content as bytes
	content, err := os.ReadFile(block.DefRange.Filename)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("error reading file %s", block.DefRange.Filename),
			Subject:  &block.DefRange,
		}
		res.HandleDecodeDiags(hcl.Diagnostics{diag})
		return triggerHcl, res
	}
	src := string(content)

	var triggerParams []modconfig.PipelineParam
	for _, block := range triggerOptions.Blocks {
		if block.Type == schema.BlockTypeParam {
			param, diags := decodePipelineParam(src, block, parseCtx)
			if len(diags) > 0 {
				res.HandleDecodeDiags(diags)
				return triggerHcl, res
			}
			triggerParams = append(triggerParams, *param)
		}
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
	res.AddDiags(moreDiags)

	triggerHcl.Params = triggerParams

	return triggerHcl, res
}

// TODO: validation - if you specify invalid depends_on it doesn't error out
// TODO: validation - invalid name?
func decodePipeline(mod *modconfig.Mod, block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Pipeline, *DecodeResult) {
	res := NewDecodeResult()

	// get shell pipelineHcl
	pipelineHcl := modconfig.NewPipeline(mod, block)

	pipelineOptions, diags := block.Body.Content(modconfig.PipelineBlockSchema)
	if diags.HasErrors() {
		res.HandleDecodeDiags(diags)
		return pipelineHcl, res
	}

	diags = pipelineHcl.SetAttributes(pipelineOptions.Attributes, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
		return pipelineHcl, res
	}

	// Read the entire file content as bytes
	content, err := os.ReadFile(block.DefRange.Filename)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("error reading file %s", block.DefRange.Filename),
			Subject:  &block.DefRange,
		}
		res.HandleDecodeDiags(hcl.Diagnostics{diag})
		return pipelineHcl, res
	}
	src := string(content)

	// use a map keyed by a string for fast lookup
	// we use an empty struct as the value type, so that
	// we don't use up unnecessary memory
	// foundOptions := map[string]struct{}{}
	for _, block := range pipelineOptions.Blocks {
		switch block.Type {
		case schema.BlockTypePipelineStep:
			step, diags := decodeStep(mod, block, parseCtx, pipelineHcl)
			if diags.HasErrors() {
				res.HandleDecodeDiags(diags)

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
				res.HandleDecodeDiags(diags)
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
							res.HandleDecodeDiags(diags)
							return pipelineHcl, res
						}
					}
				}

				pipelineHcl.OutputConfig = append(pipelineHcl.OutputConfig, *output)
			}

		case schema.BlockTypeParam:
			pipelineParam, moreDiags := decodePipelineParam(src, block, parseCtx)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				res.HandleDecodeDiags(diags)
				return pipelineHcl, res
			}

			if pipelineParam != nil {

				// check for duplicate pipeline parameters names
				if len(pipelineHcl.Params) > 0 {
					p := pipelineHcl.GetParam(pipelineParam.Name)

					if p != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  fmt.Sprintf("duplicate pipeline parameter name '%s' - parameter names must be unique", pipelineParam.Name),
							Subject:  &block.DefRange,
						})
						res.HandleDecodeDiags(diags)
						return pipelineHcl, res
					}
				}

				pipelineHcl.Params = append(pipelineHcl.Params, *pipelineParam)
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

	diags = validatePipelineSteps(pipelineHcl)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)

		return pipelineHcl, res
	}

	handlePipelineDecodeResult(pipelineHcl, res, block, parseCtx)
	diags = validatePipelineDependencies(pipelineHcl, parseCtx.Credentials, parseCtx.PipelingConnections)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)

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

func validatePipelineSteps(pipelineHcl *modconfig.Pipeline) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	stepMap := map[string]bool{}

	for _, step := range pipelineHcl.Steps {

		if _, ok := stepMap[step.GetFullyQualifiedName()]; ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("duplicate step name '%s' - step names must be unique", step.GetFullyQualifiedName()),
				Subject:  step.GetRange(),
			})
			continue
		}

		stepMap[step.GetFullyQualifiedName()] = true

		moreDiags := step.Validate()
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
			continue
		}
	}

	return diags
}

func validatePipelineDependencies(pipelineHcl *modconfig.Pipeline, credentials map[string]credential.Credential, connections map[string]modconfig.PipelingConnection) hcl.Diagnostics {
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

	var connectionRegisters []string
	availableConnectionTypes := map[string]bool{}
	for k := range connections {
		parts := strings.Split(k, ".")
		if len(parts) != 2 {
			continue
		}

		// Add the connection to the register
		connectionRegisters = append(connectionRegisters, k)

		// List out the supported connection types
		availableConnectionTypes[parts[0]] = true
	}

	var connectionTypes []string
	for k := range availableConnectionTypes {
		connectionTypes = append(connectionTypes, k)
	}

	for _, step := range pipelineHcl.Steps {
		dependsOn := step.GetDependsOn()

		for _, dep := range dependsOn {
			if !helpers.StringSliceContains(stepRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s', step '%s' does not exist in pipeline %s", dep, dep, pipelineHcl.Name()),
					Detail:   fmt.Sprintf("valid steps are: %s", strings.Join(stepRegisters, ", ")),
					Subject:  step.GetRange(),
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
						Summary:  fmt.Sprintf("invalid depends_on '%s', credential type '%s' not supported in pipeline %s", dep, parts[0], pipelineHcl.Name()),
						Detail:   fmt.Sprintf("valid credential types are: %s", strings.Join(credentialTypes, ", ")),
						Subject:  step.GetRange(),
					})
				}
				continue
			}

			if !helpers.StringSliceContains(credentialRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s', credential does not exist in pipeline %s", dep, pipelineHcl.Name()),
					Detail:   fmt.Sprintf("valid credentials are: %s", strings.Join(credentialRegisters, ", ")),
					Subject:  step.GetRange(),
				})
			}
		}

		connectionDependsOn := step.GetConnectionDependsOn()
		for _, dep := range connectionDependsOn {
			// Check if the credential type is supported, if <dynamic>
			parts := strings.Split(dep, ".")
			if len(parts) != 2 {
				continue
			}

			if parts[1] == "<dynamic>" {
				if !availableConnectionTypes[parts[0]] {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("invalid depends_on '%s', connection type '%s' not supported in pipeline %s", dep, parts[0], pipelineHcl.Name()),
						Detail:   fmt.Sprintf("valid connection types are: %s", strings.Join(connectionTypes, ", ")),
						Subject:  step.GetRange(),
					})
				}
				continue
			}

			if !helpers.StringSliceContains(connectionRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s', connection does not exist in pipeline %s", dep, pipelineHcl.Name()),
					Subject:  step.GetRange(),
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
					Summary:  fmt.Sprintf("invalid depends_on '%s' in output block, '%s' does not exist in pipeline %s", dep, dep, pipelineHcl.Name()),
					Subject:  outputConfig.Range,
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
		res.AddDiags(moreDiags)

		moreDiags = parseCtx.AddPipeline(resource)
		res.AddDiags(moreDiags)
		return
	}

	// failure :(
	if len(res.Depends) > 0 {
		moreDiags := parseCtx.AddDependencies(block, resource.Name(), res.Depends)
		res.AddDiags(moreDiags)
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
	case schema.BlockTypePipelineStepMessage:
		return modconfig.PipelineStepMessageBlockSchema
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
	case schema.IntegrationTypeMsTeams:
		return modconfig.IntegrationTeamsBlockSchema
	default:
		return nil
	}
}
