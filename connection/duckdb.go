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

const DuckDbConnectionType = "duckdb"

type DuckDbConnection struct {
	ConnectionImpl
	FileName *string `json:"file_name,omitempty" cty:"file_name" hcl:"file_name,optional"`
	// used only to set the connection string from command line variable value with a connection string
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string"`
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
		return c.Pipes.Resolve(ctx, &DuckDbConnection{ConnectionImpl: c.ConnectionImpl})
	}

	return c, nil
}

func (c *DuckDbConnection) Validate() hcl.Diagnostics {
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

func (c *DuckDbConnection) GetEnv() map[string]cty.Value {
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

	return utils.PtrEqual(c.FileName, other.FileName)

}

func (c *DuckDbConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *DuckDbConnection) GetConnectionString(opts ...ConnectionStringOpt) (string, error) {
	for _, opt := range opts {
		opt(c)
	}

	if c.ConnectionString != nil {
		return *c.ConnectionString, nil
	}

	return fmt.Sprintf("duckdb://%s", c.getFileName()), nil
}

func (c *DuckDbConnection) getFileName() any {
	if c.FileName != nil {
		return *c.FileName
	}
	return os.Getenv("DUCKDB_FILENAME")
}
