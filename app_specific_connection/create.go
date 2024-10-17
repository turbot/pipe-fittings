package app_specific_connection

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"reflect"
	"strings"
)

var BaseConnectionCtyType = cty.Capsule("BaseConnectionCtyType", reflect.TypeOf(&connection.ConnectionImpl{}))

func NewPipelingConnection(connectionType, shortName string, declRange hcl.Range) (connection.PipelingConnection, error) {
	ctor, exists := ConnectionTypeRegistry[connectionType]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid connection type " + connectionType)
	}

	return ctor(shortName, declRange), nil
}

func ConnectionCtyType(connectionType string) cty.Type {
	// NOTE: if not type is provided, just return the type of ConnectionImpl
	if connectionType == "" {
		return BaseConnectionCtyType
	}

	ctor, exists := ConnectionTypeRegistry[connectionType]
	if !exists {
		return cty.NilType
	}
	// instantiate connection
	inst := ctor("", hcl.Range{})
	goType := reflect.TypeOf(inst)
	// dereference pointer
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	return cty.Capsule(connectionType, goType)
}

func DefaultPipelingConnections() (map[string]connection.PipelingConnection, error) {
	conns := make(map[string]connection.PipelingConnection)

	for k := range ConnectionTypeRegistry {
		// check for a default connection of this type
		defaultConnection, exists := DefaultConnections[k]
		// if no default was specified, create an empty connection
		if !exists {
			var err error
			defaultConnection, err = NewPipelingConnection(k, "default", hcl.Range{})
			if err != nil {
				return nil, err
			}
		}

		conns[k+".default"] = defaultConnection
	}

	return conns, nil
}

// ConnectionStringFromConnectionName resolves the connection name to a conneciton onbject in the eval context and
// return the connection string
func ConnectionStringFromConnectionName(evalContext *hcl.EvalContext, longName string) (string, error) {
	ty, name, err := parseConnectionName(longName)
	if err != nil {
		return "", err
	}
	// look in the eval context for the connection
	connectionMap, ok := evalContext.Variables["connection"]
	if !ok {
		return "", perr.BadRequestWithMessage("unable to resolve connection - not found in eval context: " + longName)
	}
	connextionsOfType, ok := connectionMap.AsValueMap()[ty]
	if !ok {
		return "", perr.BadRequestWithMessage("unable to resolve connection - not found in eval context: " + longName)
	}
	connCty, ok := connextionsOfType.AsValueMap()[name]
	if !ok {
		return "", perr.BadRequestWithMessage("unable to resolve connection - not found in eval context: " + longName)
	}
	conn, err := CtyValueToConnection(connCty)
	if err != nil {
		return "", err
	}
	csp, ok := conn.(connection.ConnectionStringProvider)
	if !ok {
		return "", perr.BadRequestWithMessage("connection does not support connection string: " + longName)
	}
	return csp.GetConnectionString(), nil
}

// parseConnectionName parses the connection name in the form "connection.<type>.<name>", and returns the type and name
func parseConnectionName(longName string) (ty, name string, err error) {
	if !strings.HasPrefix(longName, "connection.") {
		return "", "", perr.BadRequestWithMessage("invalid connection reference: " + longName)
	}

	parts := strings.Split(longName, ".")
	if len(parts) != 3 {
		return "", "", perr.BadRequestWithMessage("invalid connection reference: " + longName)
	}
	ty = parts[1]
	name = parts[2]

	return ty, name, nil
}
