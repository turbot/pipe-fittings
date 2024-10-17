package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const JumpCloudConnectionType = "jumpcloud"

type JumpCloudConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func NewJumpCloudConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &JumpCloudConnection{
		ConnectionImpl: NewConnectionImpl(JumpCloudConnectionType, shortName, declRange),
	}
}
func (c *JumpCloudConnection) GetConnectionType() string {
	return JumpCloudConnectionType
}

func (c *JumpCloudConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &JumpCloudConnection{ConnectionImpl: c.ConnectionImpl})
	}

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

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *JumpCloudConnection) Validate() hcl.Diagnostics {
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

func (c *JumpCloudConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *JumpCloudConnection) GetEnv() map[string]cty.Value {
	return nil
}
