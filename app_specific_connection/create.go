package app_specific_connection

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

func NewPipelingConnection(connectionType, shortName string, declRange hcl.Range) (connection.PipelingConnection, error) {
	ctor, exists := ConnectionTypeRegistry[connectionType]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid connection type " + connectionType)
	}

	return ctor(shortName, declRange), nil
}

func ConnectionCtyType(connectionType string) cty.Type {
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
		defaultCred, err := NewPipelingConnection(k, "default", hcl.Range{})
		if err != nil {
			return nil, err
		}

		conns[k+".default"] = defaultCred

		RegisterConnectionType(k)
	}

	return conns, nil
}

var ConnectionTypeLookup = make(map[string]struct{}, 0)

func RegisterConnectionType(connectionType string) {
	ConnectionTypeLookup[connectionType] = struct{}{}
}

func ConnectionBlockForType(connectionType string) *hcl.Block {
	return &hcl.Block{
		Type:   connectionType,
		Labels: []string{connectionType},
	}
}
