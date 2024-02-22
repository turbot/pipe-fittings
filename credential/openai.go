package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type OpenAICredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *OpenAICredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["OPENAI_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *OpenAICredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OpenAICredential) Equals(other *OpenAICredential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && other == nil {
		return true
	}

	if (c == nil && other != nil) || (c != nil && other == nil) {
		return false
	}

	if !utils.StringPtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *OpenAICredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		apiKeyEnvVar := os.Getenv("OPENAI_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &OpenAICredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &apiKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *OpenAICredential) GetTtl() int {
	return -1
}

func (c *OpenAICredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type OpenAIConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
}

func (c *OpenAIConnectionConfig) GetCredential(name string) Credential {

	openAICred := &OpenAICredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		APIKey: c.APIKey,
	}

	return openAICred
}
