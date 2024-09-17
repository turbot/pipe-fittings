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

type ClickUpConnection struct {
	modconfig.ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *ClickUpConnection) GetConnectionType() string {
	return "clickup"
}

func (c *ClickUpConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
	if c.Token == nil {
		clickUpAPITokenEnvVar := os.Getenv("CLICKUP_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &ClickUpConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &clickUpAPITokenEnvVar,
		}

		return newConnection, nil
	}

	return c, nil
}

func (c *ClickUpConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*ClickUpConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *ClickUpConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *ClickUpConnection) GetTtl() int {
	return -1
}

func (c *ClickUpConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ClickUpConnection) GetEnv() map[string]cty.Value {
	// There is no environment variable listed in the ClickUp official API docs
	// https://clickup.com/api/developer-portal/authentication/
	return nil
}
