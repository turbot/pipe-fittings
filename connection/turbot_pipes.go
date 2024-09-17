package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PipesConnection struct {
	modconfig.ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PipesConnection) GetConnectionType() string {
	return "pipes"
}

func (c *PipesConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
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

func (c *PipesConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
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

	return true
}

func (c *PipesConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *PipesConnection) GetTtl() int {
	return -1
}

func (c *PipesConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PipesConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PIPES_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}
