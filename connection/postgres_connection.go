package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const PostgresConnectionType = "postgres"

type PostgresConnection struct {
	ConnectionImpl
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string" hcl:"connection_string,optional"`
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

	// if pipes is nil, we must have a connection string, so there is nothing to so
	return c, nil
}

func (c *PostgresConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.ConnectionString != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}
	if c.Pipes == nil && c.ConnectionString == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "either pipes block or connection_string should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}
	return hcl.Diagnostics{}
}

func (c *PostgresConnection) GetConnectionString() string {
	// we always expect connection string to be set if GetConnectionString is called
	return typehelpers.SafeString(c.ConnectionString)
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

	return utils.PtrEqual(c.ConnectionString, other.ConnectionString) &&
		utils.PtrEqual(c.SearchPath, other.SearchPath) &&
		c.GetConnectionImpl().Equals(other.GetConnectionImpl())

}

func (c *PostgresConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}
