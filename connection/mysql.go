package connection

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"strconv"
)

const MysqlConnectionType = "mysql"

var (
	defaultMysqlDbName = "mysql"
	defaultMysqlUser   = "root"
	defaultMysqlPort   = 3306
	defaultMysqlHost   = "localhost"
)

type MysqlConnection struct {
	ConnectionImpl
	DbName           *string `json:"db,omitempty" cty:"db" hcl:"db,optional"`
	UserName         *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int    `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	Password         *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	ConnectionString *string `json:"connection_string,omitempty" cty:"connection_string"`
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
	// if pipes metadata is set, no other properties should be sets
	if c.Pipes != nil {
		if c.UserName != nil || c.Host != nil || c.Port != nil || c.Password != nil {
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
			c.DbName = &defaultMysqlDbName
		}
		if c.UserName == nil {
			c.UserName = &defaultMysqlUser
		}

		if c.Host == nil {
			c.Host = &defaultMysqlHost
		}

		if c.Port == nil {
			c.Port = &defaultMysqlPort
		}
	}

	return nil
}

func (c *MysqlConnection) GetConnectionString(opts ...ConnectionStringOpt) (string, error) {
	for _, opt := range opts {
		opt(c)
	}

	if c.ConnectionString != nil {
		return *c.ConnectionString, nil
	}

	db := c.getDbName()
	user := c.getUserName()
	host := c.getHost()
	port := c.getPort()
	password := typehelpers.SafeString(c.Password)

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

func (c *MysqlConnection) GetEnv() map[string]cty.Value {
	return map[string]cty.Value{
		"MYSQL_TCP_PORT": cty.StringVal(strconv.Itoa(c.getPort())),
		"MYSQL_TCP_HOST": cty.StringVal(c.getHost()),
	}
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

	return utils.PtrEqual(c.UserName, other.UserName) &&
		utils.PtrEqual(c.Host, other.Host) &&
		utils.PtrEqual(c.Port, other.Port) &&
		utils.PtrEqual(c.Password, other.Password) &&
		c.GetConnectionImpl().Equals(other.GetConnectionImpl())
}

func (c *MysqlConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *MysqlConnection) getPort() int {
	if c.Port != nil {
		return *c.Port
	}
	return defaultMysqlPort
}

func (c *MysqlConnection) getHost() string {
	if c.Host != nil {
		return *c.Host
	}
	return defaultMysqlHost
}

func (c *MysqlConnection) getDbName() string {
	if c.DbName != nil {
		return *c.DbName
	}
	return defaultMysqlDbName
}

func (c *MysqlConnection) getUserName() string {
	if c.UserName != nil {
		return *c.UserName
	}
	return defaultMysqlUser
}
