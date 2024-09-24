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

const VirusTotalConnectionType = "virus_total"

type VirusTotalConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func NewVirusTotalConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &VirusTotalConnection{
		ConnectionImpl: NewConnectionImpl(VirusTotalConnectionType, shortName, declRange),
	}
}
func (c *VirusTotalConnection) GetConnectionType() string {
	return VirusTotalConnectionType
}

func (c *VirusTotalConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

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

	impl := c.GetConnectionImpl()
	if impl.Equals(otherConnection.GetConnectionImpl()) == false {
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

func (c *VirusTotalConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *VirusTotalConnection) GetEnv() map[string]cty.Value {
	return nil
}
