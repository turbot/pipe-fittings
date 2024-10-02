package connection

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"net/url"
)

const SteampipePgConnectionType = "steampipe_pg"

type SteampipePgConnection struct {
	ConnectionImpl
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string" hcl:"connection_string,optional"`
	DbName           *string `json:"db,omitempty" cty:"db" hcl:"db,optional"`
	UserName         *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int    `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	Password         *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	SearchPath       *string `json:"search_path,omitempty" cty:"search_path" hcl:"search_path,optional"`
	SslMode          *string `json:"sslmode,omitempty" cty:"sslmode" hcl:"sslmode,optional"`
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
	// if pipes metadata is set, no other properties should be sets
	if c.Pipes != nil {
		if c.ConnectionString != nil || c.UserName != nil || c.Host != nil || c.Port != nil || c.Password != nil || c.SearchPath != nil {
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
	// if pipes is not set, either connection_string or user AND db must be set
	if c.ConnectionString == nil {
		if c.UserName == nil || c.DbName == nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "either connection_string or username and db must be set",
					Subject:  c.DeclRange.HclRangePointer(),
				},
			}
		}
	} else {
		// so connection string is set, user and db should not be set
		if c.UserName != nil || c.DbName != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "cannot set both connection_string and username/db",
					Subject:  c.DeclRange.HclRangePointer(),
				},
			}
		}
	}
	return nil

}

func (c *SteampipePgConnection) GetConnectionString() string {
	if c.ConnectionString != nil {
		return *c.ConnectionString
	}

	// we know that username and db are set
	user := typehelpers.SafeString(c.UserName)
	db := typehelpers.SafeString(c.DbName)
	var host, password string
	var port int
	if c.Host != nil {
		host = *c.Host
	} else {
		host = "localhost"
	}
	if c.Port != nil {
		port = *c.Port
	} else {
		port = 5432
	}
	if c.Password != nil {
		password = *c.Password
	}
	sslmode := typehelpers.SafeString(c.SslMode)

	// Use url.URL to encode the connection string parameters safely
	connStr := url.URL{
		Scheme: "postgresql",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   db, // This adds the /dbname part in the connection string
	}

	// Set the user with or without the password
	if password == "" {
		connStr.User = url.User(user) // No password
	} else {
		connStr.User = url.UserPassword(user, password)
	}

	// Add SSL mode or other query parameters if needed
	q := connStr.Query()
	if sslmode != "" {
		q.Add("sslmode", sslmode)
	}
	connStr.RawQuery = q.Encode()

	return connStr.String()

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
