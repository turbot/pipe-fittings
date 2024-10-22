package app_specific_connection

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
)

type ConnectionFunc func(string, hcl.Range) connection.PipelingConnection

var ConnectionTypeRegistry map[string]ConnectionFunc

func ConnectionTypeSupported(connectionType string) bool {
	_, exists := ConnectionTypeRegistry[connectionType]
	return exists
}

func RegisterConnections(funcs ...ConnectionFunc) {
	if ConnectionTypeRegistry == nil {
		ConnectionTypeRegistry = make(map[string]ConnectionFunc)
	}

	for _, f := range funcs {
		c := f("", hcl.Range{})
		ConnectionTypeRegistry[c.GetConnectionType()] = f
	}
}
