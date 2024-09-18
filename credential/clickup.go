package credential

import (
	"context"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type ClickUpCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *ClickUpCredential) getEnv() map[string]cty.Value {
	// There is no environment variable listed in the ClickUp official API docs
	// https://clickup.com/api/developer-portal/authentication/
	return nil
}

func (c *ClickUpCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ClickUpCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*ClickUpCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *ClickUpCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		clickUpAPITokenEnvVar := os.Getenv("CLICKUP_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &ClickUpCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &clickUpAPITokenEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *ClickUpCredential) GetTtl() int {
	return -1
}

func (c *ClickUpCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type ClickUpConnectionConfig struct {
	Token *string `cty:"token" hcl:"token"`
}

func (c *ClickUpConnectionConfig) GetCredential(name string, shortName string) Credential {

	clickUpCred := &ClickUpCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "clickup",
		},

		Token: c.Token,
	}

	return clickUpCred
}
