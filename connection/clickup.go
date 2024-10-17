package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const ClickUpConnectionType = "clickup"

type ClickUpConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewClickUpConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &ClickUpConnection{
		ConnectionImpl: NewConnectionImpl(ClickUpConnectionType, shortName, declRange),
	}
}
func (c *ClickUpConnection) GetConnectionType() string {
	return ClickUpConnectionType
}

func (c *ClickUpConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &ClickUpConnection{ConnectionImpl: c.ConnectionImpl})
	}

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

func (c *ClickUpConnection) Equals(otherConnection PipelingConnection) bool {
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

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *ClickUpConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.Token != nil) {
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

func (c *ClickUpConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *ClickUpConnection) GetEnv() map[string]cty.Value {
	// There is no environment variable listed in the ClickUp official API docs
	// https://clickup.com/api/developer-portal/authentication/
	return nil
}
