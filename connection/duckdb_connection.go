package connection

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const DuckDbConnectionType = "duckdb"

type DuckDbConnection struct {
	ConnectionImpl
	ConnectionString *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
}

func NewDuckDbConnection(block *hcl.Block) PipelingConnection {
	return &DuckDbConnection{
		ConnectionImpl: NewConnectionImpl(block),
	}
}

func (p *DuckDbConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if we have a connection string, return it as is
	if p.ConnectionString != nil {
		return p, nil
	}

	// TODO KAI build a connection string from the other fields
	panic("implement me")

}

func (p *DuckDbConnection) GetTtl() int {
	return -1
}

func (p *DuckDbConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (p *DuckDbConnection) GetEnv() map[string]cty.Value {
	// TODO DUCKDB ENV
	return map[string]cty.Value{}
}

func (p *DuckDbConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (p == nil && !helpers.IsNil(otherConnection)) || (p != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*DuckDbConnection)
	if !ok {
		return false
	}

	return utils.PtrEqual(p.ConnectionString, other.ConnectionString)

}

func (p *DuckDbConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(p)
}

func (p *DuckDbConnection) GetConnectionString() string {
	if p.ConnectionString != nil {
		return *p.ConnectionString
	}
	return ""
}
