package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type IPstackConnection struct {
	ConnectionImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
}

func (c *IPstackConnection) GetConnectionType() string {
	return "ipstack"
}

func (c *IPstackConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.AccessKey == nil {
		// The order of precedence for the IPstack access key environment variable
		// 1. IPSTACK_ACCESS_KEY
		// 2. IPSTACK_TOKEN

		ipstackAccessKeyEnvVar := os.Getenv("IPSTACK_TOKEN")
		if os.Getenv("IPSTACK_ACCESS_KEY") != "" {
			ipstackAccessKeyEnvVar = os.Getenv("IPSTACK_ACCESS_KEY")
		}

		// Don't modify existing connection, resolve to a new one
		newConnection := &IPstackConnection{
			ConnectionImpl: c.ConnectionImpl,
			AccessKey:      &ipstackAccessKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *IPstackConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*IPstackConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AccessKey, other.AccessKey) {
		return false
	}

	return true
}

func (c *IPstackConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *IPstackConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IPstackConnection) GetTtl() int {
	return -1
}

func (c *IPstackConnection) GetEnv() map[string]cty.Value {
	return nil
}
