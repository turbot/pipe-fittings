package credential

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type MastodonCredential struct {
	CredentialImpl

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
	Server      *string `json:"server,omitempty" cty:"server" hcl:"server,optional"`
}

func (c *MastodonCredential) getEnv() map[string]cty.Value {
	// Mastodon has no standard environment variable mentioned anywhere in the docs
	return nil
}

func (c *MastodonCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *MastodonCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*MastodonCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AccessToken, other.AccessToken) {
		return false
	}

	if !utils.PtrEqual(c.Server, other.Server) {
		return false
	}

	return true
}

func (c *MastodonCredential) Resolve(ctx context.Context) (Credential, error) {
	return c, nil
}

func (c *MastodonCredential) GetTtl() int {
	return -1
}

func (c *MastodonCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type MastodonConnectionConfig struct {
	AccessToken *string `cty:"access_token" hcl:"access_token"`
	App         *string `cty:"app" hcl:"app"`
	MaxToots    *string `cty:"max_toots" hcl:"max_toots"`
	Server      *string `cty:"server" hcl:"server"`
}

func (c *MastodonConnectionConfig) GetCredential(name string, shortName string) Credential {

	mastodonCred := &MastodonCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "mastodon",
		},

		AccessToken: c.AccessToken,
		Server:      c.Server,
	}

	return mastodonCred
}
