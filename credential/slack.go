package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type SlackCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *SlackCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["SLACK_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *SlackCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *SlackCredential) Equals(other *SlackCredential) bool {
	if c.Type != other.Type {
		return false
	}

	if !utils.StringPtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *SlackCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		slackTokenEnvVar := os.Getenv("SLACK_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &SlackCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &slackTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *SlackCredential) GetTtl() int {
	return -1
}

func (c *SlackCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type SlackConnectionConfig struct {
	Token *string `cty:"token" hcl:"token"`
}

func (c *SlackConnectionConfig) GetCredential(name string) Credential {

	slackCred := &SlackCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		Token: c.Token,
	}

	return slackCred
}
