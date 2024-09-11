package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type SendGridConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *SendGridConnection) GetConnectionType() string {
	return "sendgrid"
}

func (c *SendGridConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.APIKey == nil {
		sendGridAPIKeyEnvVar := os.Getenv("SENDGRID_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &SendGridConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &sendGridAPIKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *SendGridConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*SendGridConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *SendGridConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *SendGridConnection) GetTtl() int {
	return -1
}

func (c *SendGridConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *SendGridConnection) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["SENDGRID_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}
