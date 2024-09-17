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

type PagerDutyConnection struct {
	modconfig.ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PagerDutyConnection) GetConnectionType() string {
	return "pagerduty"
}

func (c *PagerDutyConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
	if c.Token == nil {
		pagerDutyTokenEnvVar := os.Getenv("PAGERDUTY_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &PagerDutyConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &pagerDutyTokenEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *PagerDutyConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*PagerDutyConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *PagerDutyConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *PagerDutyConnection) GetTtl() int {
	return -1
}

func (c *PagerDutyConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PagerDutyConnection) GetEnv() map[string]cty.Value {
	return nil
}
