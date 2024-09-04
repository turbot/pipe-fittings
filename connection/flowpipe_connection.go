package connection

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

var connectionTypRegistry = map[string]reflect.Type{
	"aws":    reflect.TypeOf(AwsConnection{}),
	"github": reflect.TypeOf(GithubConnection{}),
	"slack":  reflect.TypeOf(SlackConnection{}),
}

func NewConnection(block *hcl.Block) (PipelingConnection, error) {
	connectionType := block.Labels[0]
	connectionName := block.Labels[1]

	hclResourceImpl := modconfig.NewHclResourceImplNoMod(block, connectionType, connectionName)

	conn, err := instantiateConnection(connectionType, hclResourceImpl)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func instantiateConnection(key string, hclResourceImpl modconfig.HclResourceImpl) (PipelingConnection, error) {
	t, exists := connectionTypRegistry[key]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid connection type " + key)
	}
	credInterface := reflect.New(t).Interface()
	cred, ok := credInterface.(PipelingConnection)
	if !ok {
		return nil, perr.InternalWithMessage("Failed to create connection")
	}
	cred.SetHclResourceImpl(hclResourceImpl)
	cred.SetConnectionType(key)

	return cred, nil
}

type ConnectionImpl struct {
	modconfig.HclResourceImpl
	modconfig.ResourceWithMetadataImpl

	// required to allow partial decoding
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`
}

func (c *ConnectionImpl) GetUnqualifiedName() string {
	return c.HclResourceImpl.UnqualifiedName
}

func (c *ConnectionImpl) SetHclResourceImpl(hclResourceImpl modconfig.HclResourceImpl) {
	c.HclResourceImpl = hclResourceImpl
}

func (c *ConnectionImpl) GetConnectionType() string {
	return c.Type
}

func (c *ConnectionImpl) SetConnectionType(credType string) {
	c.Type = credType
}

func DefaultPipelingConnections() (map[string]PipelingConnection, error) {
	conns := make(map[string]PipelingConnection)

	for k := range connectionTypRegistry {
		hclResourceImpl := modconfig.HclResourceImpl{
			FullName:        k + ".default",
			ShortName:       "default",
			UnqualifiedName: k + ".default",
		}

		defaultCred, err := instantiateConnection(k, hclResourceImpl)
		if err != nil {
			return nil, err
		}

		conns[k+".default"] = defaultCred

		error_helpers.RegisterConnectionType(k)
	}

	return conns, nil
}
