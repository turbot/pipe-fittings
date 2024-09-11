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

type VirusTotalConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *VirusTotalConnection) GetConnectionType() string {
	return "virustotal"
}

func (c *VirusTotalConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.APIKey == nil {
		virusTotalAPIKeyEnvVar := os.Getenv("VTCLI_APIKEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &VirusTotalConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &virusTotalAPIKeyEnvVar,
		}

		return newConnection, nil

	}
	return c, nil
}

func (c *VirusTotalConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*VirusTotalConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *VirusTotalConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *VirusTotalConnection) GetTtl() int {
	return -1
}

func (c *VirusTotalConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *VirusTotalConnection) getEnv() map[string]cty.Value {
	return nil
}
