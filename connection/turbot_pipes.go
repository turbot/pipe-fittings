package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const PipesConnectionType = "turbot_pipes"

type PipesConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewPipesConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &PipesConnection{
		ConnectionImpl: NewConnectionImpl(PipesConnectionType, shortName, declRange),
	}
}
func (c *PipesConnection) GetConnectionType() string {
	return PipesConnectionType
}

func (c *PipesConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &PipesConnection{})
	}

	if c.Token == nil {
		pipesTokenEnvVar := os.Getenv("PIPES_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &PipesConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &pipesTokenEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *PipesConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*PipesConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *PipesConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.Token != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}
	return hcl.Diagnostics{}
}

func (c *PipesConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *PipesConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PIPES_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}
