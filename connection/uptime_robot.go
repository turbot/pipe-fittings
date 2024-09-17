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

type UptimeRobotConnection struct {
	modconfig.ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *UptimeRobotConnection) GetConnectionType() string {
	return "uptimerobot"
}

func (c *UptimeRobotConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
	if c.APIKey == nil {
		uptimeRobotAPIKeyEnvVar := os.Getenv("UPTIMEROBOT_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &UptimeRobotConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &uptimeRobotAPIKeyEnvVar,
		}

		return newConnection, nil
	}

	return c, nil
}

func (c *UptimeRobotConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*UptimeRobotConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *UptimeRobotConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *UptimeRobotConnection) GetTtl() int {
	return -1
}

func (c *UptimeRobotConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *UptimeRobotConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["UPTIMEROBOT_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}
