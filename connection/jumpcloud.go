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

type JumpCloudConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *JumpCloudConnection) GetConnectionType() string {
	return "jumpcloud"
}

func (c *JumpCloudConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.APIKey == nil {
		apiKeyEnvVar := os.Getenv("JUMPCLOUD_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &JumpCloudConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &apiKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *JumpCloudConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*JumpCloudConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *JumpCloudConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *JumpCloudConnection) GetTtl() int {
	return -1
}

func (c *JumpCloudConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *JumpCloudConnection) getEnv() map[string]cty.Value {
	return nil
}
