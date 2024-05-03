package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/zclconf/go-cty/cty"
)

type FreshdeskCredential struct {
	CredentialImpl

	APIKey    *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
}

func (c *FreshdeskCredential) getEnv() map[string]cty.Value {
	return nil
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

func (c *FreshdeskCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*FreshdeskCredential)
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

type FreshdeskConnectionConfig struct {
	APIKey    *string `cty:"api_key" hcl:"api_key"`
	Subdomain *string `cty:"subdomain" hcl:"subdomain"`
}

func (c *FreshdeskConnectionConfig) GetCredential(name string, shortName string) Credential {

	freshdeskCred := &FreshdeskCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "freshdesk",
		},

		APIKey:    c.APIKey,
		Subdomain: c.Subdomain,
	}

	return freshdeskCred
}
