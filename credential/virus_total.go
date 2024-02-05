package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type VirusTotalCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *VirusTotalCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["VTCLI_APIKEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *VirusTotalCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *VirusTotalCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		virusTotalAPIKeyEnvVar := os.Getenv("VTCLI_APIKEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &VirusTotalCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &virusTotalAPIKeyEnvVar,
		}

		return newCreds, nil

	}
	return c, nil
}

func (c *VirusTotalCredential) GetTtl() int {
	return -1
}

func (c *VirusTotalCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}
