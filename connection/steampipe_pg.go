package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const SteampipeConnectionType = "steampipe"

var (
	defaultSteampipeDbName = "steampipe"
	defaultSteampipeUser   = "steampipe"
	defaultSteampipePort   = 9193
	defaultSteampipeHost   = "localhost"
)

type SteampipePgConnection struct {
	ConnectionImpl
	DbName           *string   `json:"db,omitempty" cty:"db" hcl:"db,optional"`
	UserName         *string   `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string   `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int      `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	Password         *string   `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	SearchPath       *[]string `json:"search_path,omitempty" cty:"search_path" hcl:"search_path,optional"`
	SearchPathPrefix *[]string `json:"search_path_prefix,omitempty" cty:"search_path_prefix" hcl:"search_path_prefix,optional"`
	SslMode          *string   `json:"sslmode,omitempty" cty:"sslmode" hcl:"sslmode,optional"`
	ConnectionString *string   `json:"connection_string,omitempty" cty:"connection_string"`
}

func NewSteampipePgConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &SteampipePgConnection{
		ConnectionImpl: NewConnectionImpl(SteampipeConnectionType, shortName, declRange),
	}
}
func (c *SteampipePgConnection) GetConnectionType() string {
	return SteampipeConnectionType
}

func (c *SteampipePgConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &SteampipePgConnection{ConnectionImpl: c.ConnectionImpl})
	}

	// if pipes is nil, we must have a connection string, so there is nothing to so
	return c, nil
}

func (c *SteampipePgConnection) Validate() hcl.Diagnostics {
	// if pipes metadata is set, no other properties should be sets
	if c.Pipes != nil {
		if c.UserName != nil || c.Host != nil || c.Port != nil || c.Password != nil || c.SearchPath != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "if pipes block is defined, no other auth properties should be set",
					Subject:  c.DeclRange.HclRangePointer(),
				},
			}
		}
		return nil
	}

	if c.Pipes == nil {
		// set nil values to default
		if c.DbName == nil {
			c.DbName = &defaultSteampipeDbName
		}
		if c.UserName == nil {
			c.UserName = &defaultSteampipeUser
		}

		if c.Host == nil {
			c.Host = &defaultSteampipeHost
		}

		if c.Port == nil {
			c.Port = &defaultSteampipePort
		}
	}

	// validate sslmode
	if c.SslMode != nil {
		return validateSSlMode(*c.SslMode, c.DeclRange.HclRangePointer())
	}
	return nil
}

func (c *SteampipePgConnection) GetConnectionString(opts ...ConnectionStringOpt) (string, error) {
	for _, opt := range opts {
		opt(c)
	}

	if c.ConnectionString != nil {
		return *c.ConnectionString, nil
	}
	// db, username, host and port all have default values if not set
	return buildPostgresConnectionString(
		c.getDbName(),
		c.getUserName(),
		c.getHost(),
		c.getPort(),

		c.Password, c.SslMode), nil
}

func (c *SteampipePgConnection) GetEnv() map[string]cty.Value {
	// db, username, host and port all have default values if not set
	return postgresConnectionParamsToEnvValueMap(c.getDbName(),
		c.getUserName(),
		c.getHost(),
		c.getPort(),
		c.Password,
		c.SslMode)
}

func (c *SteampipePgConnection) GetSearchPath() []string {
	if c.SearchPath != nil {
		return *c.SearchPath
	}
	return []string{}
}

func (c *SteampipePgConnection) GetSearchPathPrefix() []string {
	if c.SearchPathPrefix != nil {
		return *c.SearchPathPrefix
	}
	return []string{}
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
		utils.SlicePtrEqual(c.SearchPath, other.SearchPath) &&
		utils.SlicePtrEqual(c.SearchPathPrefix, other.SearchPathPrefix) &&
		utils.PtrEqual(c.SslMode, other.SslMode) &&
		c.GetConnectionImpl().Equals(other.GetConnectionImpl())

}

func (c *SteampipePgConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

// TODO these are needed for Powerpipe because resolve is not called

func (c *SteampipePgConnection) getDbName() string {
	if c.DbName != nil {
		return *c.DbName
	}

	return defaultSteampipeDbName
}

func (c *SteampipePgConnection) getUserName() string {
	if c.UserName != nil {
		return *c.UserName
	}

	return defaultSteampipeUser
}

func (c *SteampipePgConnection) getHost() string {
	if c.Host != nil {
		return *c.Host
	}

	return defaultSteampipeHost
}

func (c *SteampipePgConnection) getPort() int {
	if c.Port != nil {
		return *c.Port
	}
	return defaultSteampipePort
}
