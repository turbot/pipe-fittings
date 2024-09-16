package app_specific_connection

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

func NewPipelingConnection(block *hcl.Block) (connection.PipelingConnection, error) {
	if len(block.Labels) == 0 {
		return nil, perr.InternalWithMessage("block must include at least type label")
	}
	connectionType := block.Labels[0]

	ctor, exists := ConnectionTypeRegistry[connectionType]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid connection type " + connectionType)
	}

	return ctor(block), nil
}
func ConnectionCtyType(connectionType string) cty.Type {
	ctor, exists := ConnectionTypeRegistry[connectionType]
	if !exists {
		return cty.NilType
	}

	goType := reflect.TypeOf(ctor(ConnectionBlockForType(connectionType)))

	return cty.Capsule(connectionType, goType)
}

func DefaultPipelingConnections() (map[string]connection.PipelingConnection, error) {
	conns := make(map[string]connection.PipelingConnection)

	for k := range ConnectionTypeRegistry {
		defaultCred, err := NewPipelingConnection(ConnectionBlockForType(k))
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
