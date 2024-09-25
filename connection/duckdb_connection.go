package connection

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const DuckDbConnectionType = "duckdb"

type DuckDbConnection struct {
	ConnectionImpl
	ConnectionString *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
}

func NewDuckDbConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &DuckDbConnection{
		ConnectionImpl: NewConnectionImpl(DuckDbConnectionType, shortName, declRange),
	}
}
func (c *DuckDbConnection) GetConnectionType() string {
	return DuckDbConnectionType
}

func (c *DuckDbConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}
	// if we have a connection string, return it as is
	if c.ConnectionString != nil {
		return c, nil
	}

	// TODO KAI build a connection string from the other fields
	panic("implement me")

}

func (c *DuckDbConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.ConnectionString != nil) {
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

func (c *DuckDbConnection) GetEnv() map[string]cty.Value {
	// TODO DUCKDB ENV
	return map[string]cty.Value{}
}

func (c *DuckDbConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*DuckDbConnection)
	if !ok {
		return false
	}

	return utils.PtrEqual(c.ConnectionString, other.ConnectionString)

}

func (c *DuckDbConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *DuckDbConnection) GetConnectionString() string {
	if c.ConnectionString != nil {
		return *c.ConnectionString
	}
	return ""
}
