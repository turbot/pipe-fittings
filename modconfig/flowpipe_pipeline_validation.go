package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func CustomValueValidation(name string, setting cty.Value, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	// this time we check if the given setting, i.e.
	// name = "example
	// type = "aws"

	// for connection actually exists in the eval context

	if hclhelpers.IsListLike(setting.Type()) {
		return pipelineParamCustomValueListValidation(name, setting, evalCtx)
	}

	if !hclhelpers.IsMapLike(setting.Type()) {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "The value for param must be an object",
		}
		return hcl.Diagnostics{diag}
	}

	settingValueMap := setting.AsValueMap()

	resourceType := ""
	if !settingValueMap["resource_type"].IsNull() {
		resourceType = settingValueMap["resource_type"].AsString()
	}

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
	} else if len(settingValueMap) > 0 {
		diags := hcl.Diagnostics{}
		for _, v := range settingValueMap {
			if v.IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "The value for param must not have a null value",
				}
				return hcl.Diagnostics{diag}
			}

			if !hclhelpers.IsComplexType(v.Type()) {
				// this test is meant for custom value validation, there's no need to test if it's not these type, i.e. connection or notifier
				continue
			}

			// this test is meant for custom value validation, there's no need to test if it's not these type, i.e. connection or notifier
			nestedDiags := CustomValueValidation(name, v, evalCtx)
			diags = append(diags, nestedDiags...)
		}

		return diags
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid value for param " + name,
		Detail:   "Invalid value for param " + name,
	}
	return hcl.Diagnostics{diag}
}

func pipelineParamCustomValueListValidation(name string, setting cty.Value, evalCtx *hcl.EvalContext) hcl.Diagnostics {

	if !hclhelpers.IsListLike(setting.Type()) {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid value for param " + name,
			Detail:   "The value for param must be a list",
		}
		return hcl.Diagnostics{diag}
	}

	var diags hcl.Diagnostics
	for it := setting.ElementIterator(); it.Next(); {
		_, element := it.Element()
		diags = append(diags, CustomValueValidation(name, element, evalCtx)...)
	}

	return diags
}