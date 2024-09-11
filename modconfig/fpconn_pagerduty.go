package modconfig

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PagerDutyConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PagerDutyConnection) GetConnectionType() string {
	return "pagerduty"
}

func (c *PagerDutyConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
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

func (c *PagerDutyConnection) Equals(otherConnection PipelingConnection) bool {
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
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PagerDutyConnection) getEnv() map[string]cty.Value {
	return nil
}
