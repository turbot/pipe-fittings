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

type FreshdeskConnection struct {
	ConnectionImpl

	APIKey    *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
}

func (c *FreshdeskConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
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

	return true
}

func (c *FreshdeskConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *FreshdeskConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *FreshdeskConnection) GetTtl() int {
	return -1
}

func (c *FreshdeskConnection) getEnv() map[string]cty.Value {
	return nil
}
