package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type ZendeskCredential struct {
	CredentialImpl

	Email     *string `json:"email,omitempty" cty:"email" hcl:"email,optional"`
	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
	Token     *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *ZendeskCredential) getEnv() map[string]cty.Value {
	return nil
}

func (c *ZendeskCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ZendeskCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*ZendeskCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Email, other.Email) {
		return false
	}

	if !utils.PtrEqual(c.Subdomain, other.Subdomain) {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
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

type ZendeskConnectionConfig struct {
	Email     *string `cty:"email" hcl:"email"`
	Subdomain *string `cty:"subdomain" hcl:"subdomain"`
	Token     *string `cty:"token" hcl:"token"`
}

func (c *ZendeskConnectionConfig) GetCredential(name string, shortName string) Credential {

	zendeskCred := &ZendeskCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "zendesk",
		},

		Email:     c.Email,
		Subdomain: c.Subdomain,
		Token:     c.Token,
	}

	return zendeskCred
}
