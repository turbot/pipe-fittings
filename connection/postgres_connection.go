package connection

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const PostgresConnectionType = "postgres"

type PostgresConnection struct {
	ConnectionImpl
	UserName         *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int    `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string" hcl:"connection_string,optional"`
	Password         *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	SearchPath       *string `json:"search_path,omitempty" cty:"search_path" hcl:"search_path,optional"`
}

func NewPostgresConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &PostgresConnection{
		ConnectionImpl: NewConnectionImpl(PostgresConnectionType, shortName, declRange),
	}
}
func (c *PostgresConnection) GetConnectionType() string {
	return PostgresConnectionType
}

func (c *PostgresConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &PostgresConnection{})
	}

	// if we have a connection string, return it as is
	if c.ConnectionString != nil {
		return c, nil
	}

	// TODO KAI build a connection string from the other fields
	panic("implement me")

}

func (c *PostgresConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.UserName != nil || c.Host != nil || c.Port != nil || c.Password != nil || c.SearchPath != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}
	return hcl.Diagnostics{}
}

func (c *PostgresConnection) GetEnv() map[string]cty.Value {
	// TODO POSTGRES ENV
	return map[string]cty.Value{}
}

func (c *PostgresConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*PostgresConnection)
	if !ok {
		return false
	}

	return utils.PtrEqual(c.UserName, other.UserName) &&
		utils.PtrEqual(c.Host, other.Host) &&
		utils.PtrEqual(c.Port, other.Port) &&
		utils.PtrEqual(c.ConnectionString, other.ConnectionString) &&
		utils.PtrEqual(c.Password, other.Password) &&
		utils.PtrEqual(c.SearchPath, other.SearchPath)

}

func (c *PostgresConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *PostgresConnection) GetConnectionString() string {
	if c.ConnectionString != nil {
		return *c.ConnectionString
	}
	return ""
}
