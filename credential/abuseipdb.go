package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type AbuseIPDBCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *AbuseIPDBCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["ABUSEIPDB_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *AbuseIPDBCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *AbuseIPDBCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		abuseIPDBAPIKeyEnvVar := os.Getenv("ABUSEIPDB_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &AbuseIPDBCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &abuseIPDBAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *AbuseIPDBCredential) GetTtl() int {
	return -1
}

func (c *AbuseIPDBCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}