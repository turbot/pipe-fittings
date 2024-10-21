package connection

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const PostgresConnectionType = "postgres"

type PostgresConnection struct {
	ConnectionImpl
	ConnectionString *string   `json:"connection_string,omitempty" cty:"connection_string" hcl:"connection_string,optional"`
	DbName           *string   `json:"db,omitempty" cty:"db" hcl:"db,optional"`
	UserName         *string   `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string   `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int      `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	Password         *string   `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	SearchPath       *[]string `json:"search_path,omitempty" cty:"search_path" hcl:"search_path,optional"`
	SearchPathPrefix *[]string `json:"search_path_prefix,omitempty" cty:"search_path_prefix" hcl:"search_path_prefix,optional"`
	SslMode          *string   `json:"sslmode,omitempty" cty:"sslmode" hcl:"sslmode,optional"`
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
	// validate sslmode
	if c.SslMode != nil {
		return validateSSlMode(*c.SslMode, c.DeclRange.HclRangePointer())
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

func (c *PostgresConnection) GetConnectionString() string {
	if c.ConnectionString != nil {
		return *c.ConnectionString
	}

	// we know that db and user are set (as it is in the validation_ sop we can ignore the error
	connString, _ := buildPostgresConnectionString(c.DbName, c.UserName, c.Host, c.Port, c.Password, c.SslMode)
	return connString
}

func (c *PostgresConnection) GetEnv() map[string]cty.Value {
	return postgresConnectionToEnvVarMap(c.ConnectionString, c.DbName, c.UserName, c.Password, c.Host, c.Port, c.SslMode)
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
		utils.PtrEqual(c.ConnectionString, other.ConnectionString) &&
		c.GetConnectionImpl().Equals(other.GetConnectionImpl())

}

func (c *PostgresConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func buildPostgresConnectionString(pDbName *string, pUserName *string, pHost *string, pPort *int, pPassword *string, pSslMode *string) (string, error) {
	if pDbName == nil || pUserName == nil {
		return "", fmt.Errorf("both username and db must be set to build a connection string")
	}

	user := typehelpers.SafeString(pUserName)
	db := typehelpers.SafeString(pDbName)
	var host, password string
	var port int
	if pHost != nil {
		host = *pHost
	} else {
		host = "localhost"
	}
	if pPort != nil {
		port = *pPort
	} else {
		port = 5432
	}
	if pPassword != nil {
		password = *pPassword
	}
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

	return connStr.String(), nil
}

func postgresConnectionToEnvVarMap(connectionString, db, username, password, host *string, port *int, sslmode *string) map[string]cty.Value {
	if connectionString != nil {
		envVars, err := connectionStringToEnvVarMap(*connectionString)
		if err != nil {
			slog.Error("error converting connection string to env vars", "error", err)
			return map[string]cty.Value{}
		}
		return envVars
	}
	return postgresConnectionParamsToEnvValueMap(*db, *username, password, host, port, sslmode)

}

func connectionStringToEnvVarMap(connString string) (map[string]cty.Value, error) {
	// Create a map to hold the environment variables and their values
	envVars := make(map[string]cty.Value)

	// Parse the connection string
	u, err := url.Parse(connString)
	if err != nil {
		if err != nil {
			return envVars, err
		}
	}

	var db, username string
	var password, host, sslmode *string
	var port *int
	// Extract username and password (username required)
	if u.User != nil {
		username = u.User.Username()
		if username == "" {
			return envVars, fmt.Errorf("username is required")
		}
		if pw, ok := u.User.Password(); ok {
			password = &pw
		}
	} else {
		return envVars, fmt.Errorf("username is required")
	}

	// Extract host and port (optional)
	h := u.Hostname()
	if h != "" {
		host = &h
	}
	if p, err := strconv.Atoi(u.Port()); err == nil {
		port = &p
	}

	// Extract DB name
	if len(u.Path) > 1 {
		db = u.Path[1:] // Remove leading "/"
	} else {
		return envVars, fmt.Errorf("database name is required")
	}

	// Extract sslmode (optional)
	sm := u.Query().Get("sslmode")
	if sm != "" {
		sslmode = &sm
	}

	return postgresConnectionParamsToEnvValueMap(db, username, password, host, port, sslmode), nil
}

func postgresConnectionParamsToEnvValueMap(db, username string, password, host *string, port *int, sslmode *string) map[string]cty.Value {
	envVars := make(map[string]cty.Value)

	// Add required fields
	envVars["PGDATABASE"] = cty.StringVal(db)
	envVars["PGUSER"] = cty.StringVal(username)

	// Add optional fields if not nil
	if password != nil {
		envVars["PGPASSWORD"] = cty.StringVal(*password)
	}
	if host != nil {
		envVars["PGHOST"] = cty.StringVal(*host)
	}
	if port != nil {
		envVars["PGPORT"] = cty.StringVal(strconv.Itoa(*port))
	}
	if sslmode != nil {
		envVars["PGSSLMODE"] = cty.StringVal(*sslmode)
	}

	// Convert the map to a cty.Value map
	return envVars
}
