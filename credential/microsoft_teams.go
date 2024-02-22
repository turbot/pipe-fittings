package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type MicrosoftTeamsCredential struct {
	CredentialImpl

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func (c *MicrosoftTeamsCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessToken != nil {
		env["TEAMS_ACCESS_TOKEN"] = cty.StringVal(*c.AccessToken)
	}
	return env
}

func (c *MicrosoftTeamsCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *MicrosoftTeamsCredential) Equals(other *MicrosoftTeamsCredential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && other == nil {
		return true
	}

	if (c == nil && other != nil) || (c != nil && other == nil) {
		return false
	}

	if !utils.StringPtrEqual(c.AccessToken, other.AccessToken) {
		return false
	}

	return true
}

func (c *MicrosoftTeamsCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.AccessToken == nil {
		msTeamsAccessTokenEnvVar := os.Getenv("TEAMS_ACCESS_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &MicrosoftTeamsCredential{
			CredentialImpl: c.CredentialImpl,
			AccessToken:    &msTeamsAccessTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *MicrosoftTeamsCredential) GetTtl() int {
	return -1
}

func (c *MicrosoftTeamsCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type MicrosoftTeamsConnectionConfig struct {
	AccessToken *string `cty:"access_token" hcl:"access_token"`
}

func (c *MicrosoftTeamsConnectionConfig) GetCredential(name string) Credential {

	microsoftTeamsCred := &MicrosoftTeamsCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		AccessToken: c.AccessToken,
	}

	return microsoftTeamsCred
}
