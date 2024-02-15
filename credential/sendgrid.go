package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type SendGridCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *SendGridCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["SENDGRID_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *SendGridCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *SendGridCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		sendGridAPIKeyEnvVar := os.Getenv("SENDGRID_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &SendGridCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &sendGridAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *SendGridCredential) GetTtl() int {
	return -1
}

func (c *SendGridCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type SendGridConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
}

func (c *SendGridConnectionConfig) GetCredential(name string) Credential {

	sendGridCred := &SendGridCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		APIKey: c.APIKey,
	}

	return sendGridCred
}
