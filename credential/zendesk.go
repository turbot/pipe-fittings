package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type ZendeskCredential struct {
	CredentialImpl

	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
	Email     *string `json:"email,omitempty" cty:"email" hcl:"email,optional"`
	Token     *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *ZendeskCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Subdomain != nil {
		env["ZENDESK_SUBDOMAIN"] = cty.StringVal(*c.Subdomain)
	}
	if c.Email != nil {
		env["ZENDESK_EMAIL"] = cty.StringVal(*c.Email)
	}
	if c.Token != nil {
		env["ZENDESK_API_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *ZendeskCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ZendeskCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.Subdomain == nil && c.Email == nil && c.Token == nil {
		subdomainEnvVar := os.Getenv("ZENDESK_SUBDOMAIN")
		emailEnvVar := os.Getenv("ZENDESK_EMAIL")
		tokenEnvVar := os.Getenv("ZENDESK_API_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &ZendeskCredential{
			CredentialImpl: c.CredentialImpl,
			Subdomain:      &subdomainEnvVar,
			Email:          &emailEnvVar,
			Token:          &tokenEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *ZendeskCredential) GetTtl() int {
	return -1
}

func (c *ZendeskCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}