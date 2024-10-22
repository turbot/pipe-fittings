package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const FreshdeskConnectionType = "freshdesk"

type FreshdeskConnection struct {
	ConnectionImpl

	APIKey    *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
}

func NewFreshdeskConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &FreshdeskConnection{
		ConnectionImpl: NewConnectionImpl(FreshdeskConnectionType, shortName, declRange),
	}
}
func (c *FreshdeskConnection) GetConnectionType() string {
	return FreshdeskConnectionType
}

func (c *FreshdeskConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &FreshdeskConnection{ConnectionImpl: c.ConnectionImpl})
	}

	freshdeskAPIKeyEnvVar := os.Getenv("FRESHDESK_API_KEY")
	freshdeskSubdomainEnvVar := os.Getenv("FRESHDESK_SUBDOMAIN")

	// Don't modify existing connection, resolve to a new one
	newConnection := &FreshdeskConnection{
		ConnectionImpl: c.ConnectionImpl,
	}

	if c.APIKey == nil {
		newConnection.APIKey = &freshdeskAPIKeyEnvVar
	} else {
		newConnection.APIKey = c.APIKey
	}

	if c.Subdomain == nil {
		newConnection.Subdomain = &freshdeskSubdomainEnvVar
	} else {
		newConnection.Subdomain = c.Subdomain
	}

	return newConnection, nil
}

func (c *FreshdeskConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*FreshdeskConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	if !utils.PtrEqual(c.Subdomain, other.Subdomain) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *FreshdeskConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.APIKey != nil || c.Subdomain != nil) {
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

func (c *FreshdeskConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *FreshdeskConnection) GetEnv() map[string]cty.Value {
	return nil
}
