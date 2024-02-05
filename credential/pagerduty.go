package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type PagerDutyCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PagerDutyCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PAGERDUTY_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *PagerDutyCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PagerDutyCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		pagerDutyTokenEnvVar := os.Getenv("PAGERDUTY_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &PagerDutyCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &pagerDutyTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *PagerDutyCredential) GetTtl() int {
	return -1
}

func (c *PagerDutyCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}
