package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const SqliteConnectionType = "Sqlite"

type SqliteConnection struct {
	ConnectionImpl
	ConnectionString *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
}

func NewSqliteConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &SqliteConnection{
		ConnectionImpl: NewConnectionImpl(SqliteConnectionType, shortName, declRange),
	}
}
func (c *SqliteConnection) GetConnectionType() string {
	return SqliteConnectionType
}

func (c *SqliteConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

	// we must have a connection string or validaiton would have failed
	return c, nil
}

func (c *SqliteConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.ConnectionString != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}

	// one of the two should be set
	if c.Pipes == nil && c.ConnectionString == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "either pipes block or database connection string should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}

	return hcl.Diagnostics{}
}

func (c *SqliteConnection) GetEnv() map[string]cty.Value {
	// TODO Sqlite ENV
	return map[string]cty.Value{}
}

func (c *SqliteConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*SqliteConnection)
	if !ok {
		return false
	}

	return utils.PtrEqual(c.ConnectionString, other.ConnectionString)

}

func (c *SqliteConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *SqliteConnection) GetConnectionString() string {
	return typehelpers.SafeString(c.ConnectionString)
}