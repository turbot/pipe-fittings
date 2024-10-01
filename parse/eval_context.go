package parse

import (
	"log/slog"
	"strings"

	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func BuildNotifierMapForEvalContext(notifiers map[string]modconfig.Notifier) (map[string]cty.Value, error) {

	varValueNotifierMap := make(map[string]cty.Value)

	for k, i := range notifiers {
		var err error
		varValueNotifierMap[k], err = i.CtyValue()
		if err != nil {
			slog.Warn("failed to convert notifier to cty value", "notifier", i.Name(), "error", err)
		}
	}

	return varValueNotifierMap, nil
}

// **WARNING** this function has a specific use case do not use
//
// The key word is "temporary"
func BuildTemporaryConnectionMapForEvalContext(allConnections map[string]connection.PipelingConnection) map[string]cty.Value {
	connectionMap := map[string]cty.Value{}

	for _, c := range allConnections {
		parts := strings.Split(c.Name(), ".")
		if len(parts) != 2 {
			// this should never happen as Name() should always return a string in the format "type.name"
			slog.Warn("connection name is not in the correct format", "connection", c.Name())
			continue
		}

		tempMap := map[string]cty.Value{
			"name":          cty.StringVal(c.GetShortName()),
			"type":          cty.StringVal(parts[0]),
			"resource_type": cty.StringVal("connection"),
			"temporary":     cty.BoolVal(true),
		}

		pCty := cty.ObjectVal(tempMap)

		connectionType := parts[0]

		if pCty != cty.NilVal {
			// Check if the connection type already exists in the map
			if existing, ok := connectionMap[connectionType]; ok {
				// If it exists, merge the new object with the existing one
				existingMap := existing.AsValueMap()
				existingMap[parts[1]] = pCty
				connectionMap[connectionType] = cty.ObjectVal(existingMap)
			} else {
				// If it doesn't exist, create a new entry
				connectionMap[connectionType] = cty.ObjectVal(map[string]cty.Value{
					parts[1]: pCty,
				})
			}
		}
	}

	return connectionMap
}

// connectionNamesValueFromVarValue takes the cty value of a variable, and if the variable contains one or more
// temporary connections, it builds a list of the connection names and returns as a cty value
func connectionNamesValueFromVarValue(v cty.Value) (cty.Value, bool) {
	var connectionNames []cty.Value
	ty := v.Type()
	if ty.IsObjectType() {
		resourceName, ok := ConnectionNameFromTemporaryConnectionMap(v.AsValueMap())
		if ok {
			connectionNames = append(connectionNames, cty.StringVal(resourceName))
		}
	} else if ty.IsListType() || ty.IsTupleType() {
		for _, val := range v.AsValueSlice() {
			ty := val.Type()
			if ty.IsObjectType() {
				resourceName, ok := ConnectionNameFromTemporaryConnectionMap(val.AsValueMap())
				if ok {
					connectionNames = append(connectionNames, cty.StringVal(resourceName))
				}
			}
		}
	}
	return cty.ListVal(connectionNames), true
}

func ConnectionNameFromTemporaryConnectionMap(valueMap map[string]cty.Value) (string, bool) {
	var resourceType, ty, name string
	var ok bool
	resourceType, ok = StringValueFromCtyMap(valueMap, "resource_type")
	if !ok || resourceType != "connection" {
		return "", false
	}
	ty, ok = StringValueFromCtyMap(valueMap, "type")
	if !ok {
		return "", false
	}
	name, ok = StringValueFromCtyMap(valueMap, "name")
	if !ok {
		return "", false

	}
	return ty + "." + name, true
}

func StringValueFromCtyMap(valueMap map[string]cty.Value, key string) (string, bool) {
	if valueMap[key] == cty.NilVal ||
		valueMap[key].IsNull() ||
		valueMap[key].Type() != cty.String {
		return "", false
	}

	return valueMap[key].AsString(), true
}
