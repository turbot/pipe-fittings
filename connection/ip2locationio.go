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

type IP2LocationIOConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *IP2LocationIOConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.APIKey == nil {
		ip2locationAPIKeyEnvVar := os.Getenv("IP2LOCATIONIO_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &IP2LocationIOConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &ip2locationAPIKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *IP2LocationIOConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *IP2LocationIOConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*IP2LocationIOConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *IP2LocationIOConnection) GetTtl() int {
	return -1
}

func (c *IP2LocationIOConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IP2LocationIOConnection) getEnv() map[string]cty.Value {
	// There is no environment variable listed in the IP2LocationIO official API docs
	// https://www.ip2location.io/ip2location-documentation
	return nil
}
