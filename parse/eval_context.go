package parse

import (
	"github.com/turbot/pipe-fittings/cty_helpers"
	"log/slog"
	"strings"

	"github.com/turbot/pipe-fittings/connection"
	"github.com/zclconf/go-cty/cty"
)

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
			"name":          cty.StringVal(c.Name()),
			"short_name":    cty.StringVal(c.GetShortName()),
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

// ConnectionNamesValueFromCtyValue takes the cty value of a variable, and if the variable contains one or more
// temporary connections, it builds a list of the connection names and returns as a cty value
func ConnectionNamesValueFromCtyValue(v cty.Value) (cty.Value, bool) {
	var connectionNames []cty.Value
	ty := v.Type()
	switch {
	case ty.IsObjectType(), ty.IsMapType():
		resourceName, ok := ConnectionNameFromTemporaryConnectionMap(v.AsValueMap())
		if ok {
			connectionNames = append(connectionNames, cty.StringVal(resourceName))
		}
	case ty.IsListType(), ty.IsTupleType():
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

	if len(connectionNames) == 0 {
		return cty.NilVal, false
	}

	return cty.ListVal(connectionNames), true
}

func ConnectionNameFromTemporaryConnectionMap(valueMap map[string]cty.Value) (string, bool) {
	var resourceType, name string
	var ok bool
	resourceType, ok = cty_helpers.StringValueFromCtyMap(valueMap, "resource_type")
	if !ok || resourceType != "connection" {
		return "", false
	}
	name, ok = cty_helpers.StringValueFromCtyMap(valueMap, "name")
	if !ok {
		return "", false

	}
	return name, true
}
