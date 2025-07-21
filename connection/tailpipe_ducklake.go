package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/turbot/go-kit/helpers"
)

const TailpipeDucklakeConnectionType = "tailpipe_ducklake"

type TailpipeDucklakeConnection struct {
	ConnectionImpl
}

func NewTailpipeDucklakeConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &TailpipeDucklakeConnection{
		ConnectionImpl: NewConnectionImpl(TailpipeDucklakeConnectionType, shortName, declRange),
	}
}

func (c *TailpipeDucklakeConnection) GetConnectionType() string {
	return TailpipeDucklakeConnectionType
}

func (c *TailpipeDucklakeConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// TODO #DL does this make any sense???

	// if pipes is nil, are able to get a connection string, so there is nothing to so
	return c, nil
}

func (c *TailpipeDucklakeConnection) Validate() hcl.Diagnostics {
	return nil
}

func (c *TailpipeDucklakeConnection) GetConnectionString(opts ...ConnectionStringOpt) (string, error) {
	for _, opt := range opts {
		opt(c)
	}

	return "", nil
}

func (c *TailpipeDucklakeConnection) GetEnv() map[string]cty.Value {
	return map[string]cty.Value{}
}
func (c *TailpipeDucklakeConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*TailpipeDucklakeConnection)

	// all ducklake connections are equal
	if !ok {
		return false
	}
	return c.GetConnectionImpl().Equals(other.GetConnectionImpl())
}

func (c *TailpipeDucklakeConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

// IsDynamic implements the DynamicConnectionStringProvider interface
// indicating that the connection string may change
// TODO #DL is this needed?
func (c *TailpipeDucklakeConnection) IsDynamic() {}
