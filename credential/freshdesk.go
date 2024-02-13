package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type FreshdeskCredential struct {
	CredentialImpl

	APIKey    *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
}

func (c *FreshdeskCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["FRESHDESK_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	if c.Subdomain != nil {
		env["FRESHDESK_SUBDOMAIN"] = cty.StringVal(*c.Subdomain)
	}
	return env
}

func (c *FreshdeskCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *FreshdeskCredential) Resolve(ctx context.Context) (Credential, error) {
	freshdeskAPIKeyEnvVar := os.Getenv("FRESHDESK_API_KEY")
	freshdeskSubdomainEnvVar := os.Getenv("FRESHDESK_SUBDOMAIN")

	// Don't modify existing credential, resolve to a new one
	newCreds := &FreshdeskCredential{
		CredentialImpl: c.CredentialImpl,
	}

	if c.APIKey == nil {
		newCreds.APIKey = &freshdeskAPIKeyEnvVar
	} else {
		newCreds.APIKey = c.APIKey
	}

	if c.Subdomain == nil {
		newCreds.Subdomain = &freshdeskSubdomainEnvVar
	} else {
		newCreds.Subdomain = c.Subdomain
	}

	return newCreds, nil
}

func (c *FreshdeskCredential) GetTtl() int {
	return -1
}

func (c *FreshdeskCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}