package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type IP2LocationIOCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *IP2LocationIOCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["IP2LOCATIONIO_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *IP2LocationIOCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IP2LocationIOCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		ip2locationAPIKeyEnvVar := os.Getenv("IP2LOCATIONIO_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &IP2LocationIOCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &ip2locationAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *IP2LocationIOCredential) GetTtl() int {
	return -1
}

func (c *IP2LocationIOCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type IP2LocationIOConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
}

func (c *IP2LocationIOConnectionConfig) GetCredential(name string) Credential {

	ip2LocationIOCred := &IP2LocationIOCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		APIKey: c.APIKey,
	}

	return ip2LocationIOCred
}
