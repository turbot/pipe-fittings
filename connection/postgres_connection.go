package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PostgresConnection struct {
	ConnectionImpl
	UserName         *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host             *string `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port             *int    `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	ConnectionString *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
	Password         *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	SearchPath       *string `json:"search_path,omitempty" cty:"search_path" hcl:"search_path,optional"`
}

func NewPostgresConnection(block *hcl.Block) PipelingConnection {
	return &PostgresConnection{
		ConnectionImpl: NewConnectionImpl(block),
	}
}

func (p *PostgresConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if we have a connection string, return it as is
	if p.ConnectionString != nil {
		return p, nil
	}

	// TODO KAI build a connection string from the other fields
	panic("implement me")

}

func (p *PostgresConnection) GetTtl() int {
	return -1
}

func (p *PostgresConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (p *PostgresConnection) GetEnv() map[string]cty.Value {
	// TODO POSTGRES ENV
	return map[string]cty.Value{}
}

func (p *PostgresConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (p == nil && !helpers.IsNil(otherConnection)) || (p != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*PostgresConnection)
	if !ok {
		return false
	}

	return utils.PtrEqual(p.UserName, other.UserName) &&
		utils.PtrEqual(p.Host, other.Host) &&
		utils.PtrEqual(p.Port, other.Port) &&
		utils.PtrEqual(p.ConnectionString, other.ConnectionString) &&
		utils.PtrEqual(p.Password, other.Password) &&
		utils.PtrEqual(p.SearchPath, other.SearchPath)

}

func (p *PostgresConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(p)
}

func (p *PostgresConnection) GetConnectionString() string {
	if p.ConnectionString != nil {
		return *p.ConnectionString
	}
	return ""
}
