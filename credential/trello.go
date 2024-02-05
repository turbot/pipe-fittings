package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type TrelloCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *TrelloCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["TRELLO_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	if c.Token != nil {
		env["TRELLO_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *TrelloCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *TrelloCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.APIKey == nil && c.Token == nil {
		apiKeyEnvVar := os.Getenv("TRELLO_API_KEY")
		tokenEnvVar := os.Getenv("TRELLO_TOKEN")
		// Don't modify existing credential, resolve to a new one
		newCreds := &TrelloCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &apiKeyEnvVar,
			Token:          &tokenEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *TrelloCredential) GetTtl() int {
	return -1
}

func (c *TrelloCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}
