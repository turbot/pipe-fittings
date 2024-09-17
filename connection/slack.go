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

type SlackConnection struct {
	modconfig.ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *SlackConnection) GetConnectionType() string {
	return "slack"
}

func (c *SlackConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
	if c.Token == nil {
		slackTokenEnvVar := os.Getenv("SLACK_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newConnection := &SlackConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &slackTokenEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *SlackConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*SlackConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *SlackConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *SlackConnection) GetTtl() int {
	return -1
}

func (c *SlackConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *SlackConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["SLACK_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}
