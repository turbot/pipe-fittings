package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const SteampipePgConnectionType = "steampipe_pg"

type SteampipePgConnection struct {
	ConnectionImpl
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string" hcl:"connection_string,optional"`
	UserName         *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int    `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	Password         *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	SearchPath       *string `json:"search_path,omitempty" cty:"search_path" hcl:"search_path,optional"`
}

func NewSteampipePgConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &SteampipePgConnection{
		ConnectionImpl: NewConnectionImpl(SteampipePgConnectionType, shortName, declRange),
	}
}
func (c *SteampipePgConnection) GetConnectionType() string {
	return SteampipePgConnectionType
}

func (c *SteampipePgConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &SteampipePgConnection{})
	}

	// if pipes is nil, we can just return ourselves - we have all the info we need
	return c, nil
}

func (c *SteampipePgConnection) Validate() hcl.Diagnostics {
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

func (c *SteampipePgConnection) GetConnectionString() string {
	if c.ConnectionString != nil {
		return *c.ConnectionString
	}
	// TODO build connection string
	panic("not implemented")
}

func (c *SteampipePgConnection) GetEnv() map[string]cty.Value {
	// TODO POSTGRES ENV
	return map[string]cty.Value{}
}

func (c *SteampipePgConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*SteampipePgConnection)
	if !ok {
		return false
	}

	return utils.PtrEqual(c.UserName, other.UserName) &&
		utils.PtrEqual(c.Host, other.Host) &&
		utils.PtrEqual(c.Port, other.Port) &&
		utils.PtrEqual(c.Password, other.Password) &&
		utils.PtrEqual(c.SearchPath, other.SearchPath) &&
		utils.PtrEqual(c.ConnectionString, other.ConnectionString) &&
		c.GetConnectionImpl().Equals(other.GetConnectionImpl())

}

func (c *SteampipePgConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}
