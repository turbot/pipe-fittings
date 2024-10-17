package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const SendGridConnectionType = "sendgrid"

type SendGridConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func NewSendGridConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &SendGridConnection{
		ConnectionImpl: NewConnectionImpl(SendGridConnectionType, shortName, declRange),
	}
}
func (c *SendGridConnection) GetConnectionType() string {
	return SendGridConnectionType
}

func (c *SendGridConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &SendGridConnection{ConnectionImpl: c.ConnectionImpl})
	}

	if c.APIKey == nil {
		sendGridAPIKeyEnvVar := os.Getenv("SENDGRID_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &SendGridConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &sendGridAPIKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *SendGridConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*SendGridConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *SendGridConnection) Validate() hcl.Diagnostics {
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

func (c *SendGridConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *SendGridConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["SENDGRID_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}
