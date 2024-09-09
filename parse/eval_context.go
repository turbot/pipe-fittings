package parse

import (
	"context"
	"log/slog"
	"strings"

	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
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
func BuildTemporaryConnectionMapForEvalContext(ctx context.Context, allConnections map[string]modconfig.PipelingConnection) (map[string]cty.Value, error) {
	connectionMap := map[string]cty.Value{}

	for _, c := range allConnections {
		parts := strings.Split(c.Name(), ".")
		if len(parts) != 2 {
			return nil, perr.BadRequestWithMessage("invalid credential name: " + c.Name())
		}

		tempMap := map[string]cty.Value{
			"name":          cty.StringVal(c.GetHclResourceImpl().ShortName),
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

	return connectionMap, nil
}
