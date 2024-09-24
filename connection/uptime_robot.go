package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const UptimeRobotConnectionType = "uptime_robot"

type UptimeRobotConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func NewUptimeRobotConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &UptimeRobotConnection{
		ConnectionImpl: NewConnectionImpl(UptimeRobotConnectionType, shortName, declRange),
	}
}
func (c *UptimeRobotConnection) GetConnectionType() string {
	return UptimeRobotConnectionType
}

func (c *UptimeRobotConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

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

func (c *UptimeRobotConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	impl := c.GetConnectionImpl()
	if impl.Equals(otherConnection.GetConnectionImpl()) == false {
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
	if c.Pipes != nil && (c.APIKey != nil) {
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

func (c *UptimeRobotConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
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
