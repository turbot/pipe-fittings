package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type DatadogCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	AppKey *string `json:"app_key,omitempty" cty:"app_key" hcl:"app_key,optional"`
	APIUrl *string `json:"api_url,omitempty" cty:"api_url" hcl:"api_url,optional"`
}

func (c *DatadogCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["DD_CLIENT_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	if c.AppKey != nil {
		env["DD_CLIENT_APP_KEY"] = cty.StringVal(*c.AppKey)
	}
	return env
}

func (c *DatadogCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *DatadogCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*DatadogCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	if !utils.PtrEqual(c.AppKey, other.AppKey) {
		return false
	}

	if !utils.PtrEqual(c.APIUrl, other.APIUrl) {
		return false
	}

	return true
}

func (c *DatadogCredential) Resolve(ctx context.Context) (Credential, error) {
	datadogAPIKeyEnvVar := os.Getenv("DD_CLIENT_API_KEY")
	datadogAppKeyEnvVar := os.Getenv("DD_CLIENT_APP_KEY")

	// Don't modify existing credential, resolve to a new one
	newCreds := &DatadogCredential{
		CredentialImpl: c.CredentialImpl,
		APIUrl:         c.APIUrl,
	}

	if c.APIKey == nil {
		newCreds.APIKey = &datadogAPIKeyEnvVar
	} else {
		newCreds.APIKey = c.APIKey
	}

	if c.AppKey == nil {
		newCreds.AppKey = &datadogAppKeyEnvVar
	} else {
		newCreds.AppKey = c.AppKey
	}

	return newCreds, nil
}

func (c *DatadogCredential) GetTtl() int {
	return -1
}

func (c *DatadogCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type DatadogConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
	APIUrl *string `cty:"api_url" hcl:"api_url"`
	AppKey *string `cty:"app_key" hcl:"app_key"`
}

func (c *DatadogConnectionConfig) GetCredential(name string) Credential {

	datadogCred := &DatadogCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		APIKey: c.APIKey,
		APIUrl: c.APIUrl,
		AppKey: c.AppKey,
	}

	return datadogCred
}
