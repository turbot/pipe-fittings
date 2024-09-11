package modconfig

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PipesConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PipesConnection) GetConnectionType() string {
	return "pipes"
}

func (c *PipesConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
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

	return true
}

func (c *PipesConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *PipesConnection) GetTtl() int {
	return -1
}

func (c *PipesConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PipesConnection) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PIPES_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}
