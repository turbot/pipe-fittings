package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type TrelloCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *TrelloCredential) getEnv() map[string]cty.Value {
	return nil
}

func (c *TrelloCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *TrelloCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*TrelloCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
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

type TrelloConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
	Token  *string `cty:"token" hcl:"token"`
}

func (c *TrelloConnectionConfig) GetCredential(name string, shortName string) Credential {

	trelloCred := &TrelloCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "trello",
		},

		APIKey: c.APIKey,
		Token:  c.Token,
	}

	return trelloCred
}
