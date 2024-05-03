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

type UrlscanCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *UrlscanCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["URLSCAN_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *UrlscanCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *UrlscanCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*UrlscanCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *UrlscanCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		urlscanAPIKeyEnvVar := os.Getenv("URLSCAN_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &UrlscanCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &urlscanAPIKeyEnvVar,
		}
		return newCreds, nil
	}

	return c, nil
}

func (c *UrlscanCredential) GetTtl() int {
	return -1
}

func (c *UrlscanCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type UrlscanConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
}

func (c *UrlscanConnectionConfig) GetCredential(name string, shortName string) Credential {

	urlscanCred := &UrlscanCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "urlscan",
		},

		APIKey: c.APIKey,
	}

	return urlscanCred
}
