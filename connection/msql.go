package connection

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const MysqlConnectionType = "Mysql"

type MysqlConnection struct {
	ConnectionImpl
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string" hcl:"connection_string,optional"`
	DbName           *string `json:"db,omitempty" cty:"db" hcl:"db,optional"`
	UserName         *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int    `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	Password         *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func NewMysqlConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &MysqlConnection{
		ConnectionImpl: NewConnectionImpl(MysqlConnectionType, shortName, declRange),
	}
}
func (c *MysqlConnection) GetConnectionType() string {
	return MysqlConnectionType
}

func (c *MysqlConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{ConnectionImpl: c.ConnectionImpl})
	}

	// we must have a connection string or validaiton would have failed
	return c, nil
}

func (c *MysqlConnection) Validate() hcl.Diagnostics {
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

func (c *MysqlConnection) GetEnv() map[string]cty.Value {
	return map[string]cty.Value{}
}

func (c *MysqlConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*MysqlConnection)
	if !ok {
		return false
	}

	return utils.PtrEqual(c.ConnectionString, other.ConnectionString)

}

func (c *MysqlConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *MysqlConnection) GetConnectionString() string {
	if c.ConnectionString != nil {
		return *c.ConnectionString
	}

	// we know that db and user are set (as it is in the validation_ sop we can ignore the error
	connString, _ := buildMysqlConnectionString(c.DbName, c.UserName, c.Host, c.Port, c.Password)
	return connString
}

func buildMysqlConnectionString(pDbName *string, pUserName *string, pHost *string, pPort *int, pPassword *string) (string, error) {
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
		port = 3306
	}
	if pPassword != nil {
		password = *pPassword
	}

	// MySQL connection string format: "mysql://user:password@tcp(host:port)/dbname
	var userString string
	if password == "" {
		userString = user
	} else {
		userString = fmt.Sprintf("%s:%s",
			user, password)
	}
	connString := fmt.Sprintf("mysql://%s@tcp(%s:%d)/%s", userString, host, port, db)

	return connString, nil
}
