package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const PagerDutyConnectionType = "pagerduty"

type PagerDutyConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewPagerDutyConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &PagerDutyConnection{
		ConnectionImpl: NewConnectionImpl(PagerDutyConnectionType, shortName, declRange),
	}
}
func (c *PagerDutyConnection) GetConnectionType() string {
	return PagerDutyConnectionType
}

func (c *PagerDutyConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &PagerDutyConnection{ConnectionImpl: c.ConnectionImpl})
	}

	if c.Token == nil {
		pagerDutyTokenEnvVar := os.Getenv("PAGERDUTY_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &PagerDutyConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &pagerDutyTokenEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *PagerDutyConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*PagerDutyConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *PagerDutyConnection) Validate() hcl.Diagnostics {
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

func (c *PagerDutyConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *PagerDutyConnection) GetEnv() map[string]cty.Value {
	return nil
}
