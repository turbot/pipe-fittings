package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const SqliteConnectionType = "sqlite"

type SqliteConnection struct {
	ConnectionImpl
	FileName *string `json:"file_name,omitempty" cty:"file_name" hcl:"file_name,optional"`
	// used only to set the connection string from command line variable value with a connection string
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string"`
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
		return c.Pipes.Resolve(ctx, &SqliteConnection{ConnectionImpl: c.ConnectionImpl})
	}

	// we must have a filename string or validation would have failed
	return c, nil
}

func (c *SqliteConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.FileName != nil) {
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

func (c *SqliteConnection) GetEnv() map[string]cty.Value {
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

	return utils.PtrEqual(c.FileName, other.FileName)

}

func (c *SqliteConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *SqliteConnection) GetConnectionString() string {
	if c.ConnectionString != nil {
		return *c.ConnectionString
	}

	return fmt.Sprintf("sqlite://%s", c.getFileName())
}

func (c *SqliteConnection) getFileName() any {
	if c.FileName != nil {
		return *c.FileName
	}
	return os.Getenv("DUCKDB_FILENAME")
}
