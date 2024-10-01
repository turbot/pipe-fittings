package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

// ResourceNamesFromLateBindingVarValueError checks if the error is due to a late binding variable
// (late binding variables are not added to the eval context as they are evaluated at run time)
func resourceNamesFromLateBindingVarValueError(e *hcl.Diagnostic, evalContext *hcl.EvalContext) []string {
	if e.Summary == "Unsupported attribute" {
		var resourceNames []string

		v := e.Expression.Variables()
		fmt.Println(v)

		for _, traversal := range v {
			if traversal.RootName() == schema.AttributeVar {
				// is there an entry for theivariable in the late binding vars map
				if lateBindingVars, ok := evalContext.Variables[constants.LateBindingVarsKey]; ok {
					// retrieve the list of resource names the late binding variable depends on
					varShortName := VarShortNameFromTraversal(traversal)
					moreResourceNames := ResourceNamesFromLateBindingVarValue(lateBindingVars, varShortName)
					resourceNames = append(resourceNames, moreResourceNames...)
				}
			}
		}
	}
	return nil
}

func VarShortNameFromTraversal(traversal hcl.Traversal) string {
	varName := hclhelpers.TraversalAsString(traversal)
	parts := strings.Split(varName, ".")
	return parts[len(parts)-1]
}

// ResourceNamesFromLateBindingVarValue checks if the variable value is a single or list of late binding resources
// (specifically - connections) and if so returns the resource names
func ResourceNamesFromLateBindingVarValue(valValue cty.Value, varShortName string) []string {
	var resourceNames []string
	if valValue.Type().IsObjectType() {
		resourceNames = GetLateBindingResourceNamesFromObject(valValue, varShortName)
	} else if valValue.Type().IsListType() || valValue.Type().IsTupleType() {
		lateBindingVars := valValue.AsValueSlice()
		for _, varValue := range lateBindingVars {
			if varValue.Type().IsObjectType() {
				moreNames := GetLateBindingResourceNamesFromObject(varValue, varShortName)
				resourceNames = append(resourceNames, moreNames...)
			}
		}
	}
	return resourceNames
}

// GetLateBindingResourceNamesFromObject checks if the variable is late binding
// and if so returns the resource (connection) names which the variable depends on
func GetLateBindingResourceNamesFromObject(val cty.Value, varShortName string) []string {
	var resourceNames []string
	lateBindingVars := val.AsValueMap()

	if lateBindingResourceNames, ok := lateBindingVars[varShortName]; ok {

		if lateBindingResourceNames.Type().IsListType() {
			for _, name := range lateBindingResourceNames.AsValueSlice() {
				resourceNames = append(resourceNames, name.AsString())
			}
		} else {
			resourceNames = append(resourceNames, lateBindingResourceNames.AsString())
		}
	}
	return resourceNames
}