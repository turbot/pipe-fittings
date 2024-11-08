package connection

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const PostgresConnectionType = "postgres"

var (
	defaultPostgresDbName = "postgres"
	defaultPostgresUser   = "postgres"
	defaultPostgresPort   = 5432
	defaultPostgresHost   = "localhost"
)

type PostgresConnection struct {
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
		return c.Pipes.Resolve(ctx, &PostgresConnection{ConnectionImpl: c.ConnectionImpl})
	}

	// if pipes is nil, we must have a connection string, so there is nothing to so
	return c, nil
}

func (c *PostgresConnection) Validate() hcl.Diagnostics {
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
			c.DbName = &defaultPostgresDbName
		}

		if c.UserName == nil {
			c.UserName = &defaultPostgresUser
		}

		if c.Host == nil {
			c.Host = &defaultPostgresHost
		}

		if c.Port == nil {
			c.Port = &defaultPostgresPort
		}

		// validate sslmode
		if c.SslMode != nil {
			return validateSSlMode(*c.SslMode, c.DeclRange.HclRangePointer())
		}
	}
	return nil
}

func validateSSlMode(s string, declRange *hcl.Range) hcl.Diagnostics {
	/*
		1. disable: No SSL connection.
		2.	allow: Prefer an SSL connection, but connect without SSL if unavailable.
		3.	prefer: Prefer SSL but allow non-SSL if SSL is unavailable (default behavior).
		4.	require: Always connect with SSL but no server certificate verification.
		5.	verify-ca: Require SSL and verify the server’s certificate is signed by a trusted CA.
		6.	verify-full: Require SSL and verify that the server’s certificate is signed by a trusted CA and that the hostname matches.

	*/
	switch s {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		return nil
	default:
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "sslmode must be one of disable, allow, prefer, require, verify-ca, verify-full",
				Subject:  declRange,
			},
		}
	}
}

func (c *PostgresConnection) GetConnectionString(opts ...ConnectionStringOpt) string {
	for _, opt := range opts {
		opt(c)
	}

	if c.ConnectionString != nil {
		return *c.ConnectionString
	}

	// db, username, host and port all have default values if not set
	connString := buildPostgresConnectionString(
		c.getDbName(),
		c.getUserName(),
		c.getHost(),
		c.getPort(),
		c.Password, c.SslMode)
	return connString
}

func (c *PostgresConnection) GetEnv() map[string]cty.Value {
	// db, username, host and port all have default values if not set
	return postgresConnectionParamsToEnvValueMap(
		c.getDbName(),
		c.getUserName(),
		c.getHost(),
		c.getPort(),
		c.Password,
		c.SslMode)
}

func (c *PostgresConnection) GetSearchPath() []string {
	if c.SearchPath != nil {
		return *c.SearchPath
	}
	return []string{}
}

func (c *PostgresConnection) GetSearchPathPrefix() []string {
	if c.SearchPathPrefix != nil {
		return *c.SearchPathPrefix
	}
	return []string{}
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
		utils.PtrEqual(c.Password, other.Password) &&
		utils.SlicePtrEqual(c.SearchPath, other.SearchPath) &&
		utils.SlicePtrEqual(c.SearchPathPrefix, other.SearchPathPrefix) &&
		utils.PtrEqual(c.SslMode, other.SslMode) &&
		c.GetConnectionImpl().Equals(other.GetConnectionImpl())

}

func (c *PostgresConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func buildPostgresConnectionString(db string, user string, host string, port int, pPassword *string, pSslMode *string) string {
	password := typehelpers.SafeString(pPassword)
	sslmode := typehelpers.SafeString(pSslMode)

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

func postgresConnectionParamsToEnvValueMap(db, username, host string, port int, password, sslmode *string) map[string]cty.Value {
	envVars := map[string]cty.Value{
		"PGDATABASE": cty.StringVal(db),
		"PGUSER":     cty.StringVal(username),
		"PGHOST":     cty.StringVal(host),
		"PGPORT":     cty.StringVal(strconv.Itoa(port)),
	}

	// Add optional fields if not nil
	if password != nil {
		envVars["PGPASSWORD"] = cty.StringVal(*password)
	}
	if sslmode != nil {
		envVars["PGSSLMODE"] = cty.StringVal(*sslmode)
	}

	// Convert the map to a cty.Value map
	return envVars
}

// get port, apply default if not set
func (c *PostgresConnection) getPort() int {
	if c.Port != nil {
		return *c.Port
	}
	return defaultPostgresPort
}

// get host, apply default if not set
func (c *PostgresConnection) getHost() string {
	if c.Host != nil {
		return *c.Host
	}
	return defaultPostgresHost
}

// get db name, apply default if not set
func (c *PostgresConnection) getDbName() string {
	if c.DbName != nil {
		return *c.DbName
	}
	return defaultPostgresDbName
}

// get user name, apply default if not set
func (c *PostgresConnection) getUserName() string {
	if c.UserName != nil {
		return *c.UserName
	}
	return defaultPostgresUser
}
