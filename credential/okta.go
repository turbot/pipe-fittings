package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type OktaCredential struct {
	CredentialImpl

	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	Domain *string `json:"domain,omitempty" cty:"domain" hcl:"domain,optional"`
}

func (c *OktaCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["OKTA_TOKEN"] = cty.StringVal(*c.Token)
	}
	if c.Domain != nil {
		env["OKTA_ORGURL"] = cty.StringVal(*c.Domain)
	}
	return env
}

func (c *OktaCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OktaCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.Token == nil && c.Domain == nil {
		apiTokenEnvVar := os.Getenv("OKTA_TOKEN")
		domainEnvVar := os.Getenv("OKTA_ORGURL")

		// Don't modify existing credential, resolve to a new one
		newCreds := &OktaCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &apiTokenEnvVar,
			Domain:         &domainEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *OktaCredential) GetTtl() int {
	return -1
}

func (c *OktaCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}