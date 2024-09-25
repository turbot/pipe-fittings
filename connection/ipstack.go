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

const IPstackConnectionType = "ipstack"

type IPstackConnection struct {
	ConnectionImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
}

func NewIPstackConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &IPstackConnection{
		ConnectionImpl: NewConnectionImpl(IPstackConnectionType, shortName, declRange),
	}
}
func (c *IPstackConnection) GetConnectionType() string {
	return IPstackConnectionType
}

func (c *IPstackConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &IPstackConnection{})
	}

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

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *IPstackConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.AccessKey != nil) {
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

func (c *IPstackConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IPstackConnection) GetEnv() map[string]cty.Value {
	return nil
}
