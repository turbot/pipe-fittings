package credential

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type MastodonCredential struct {
	CredentialImpl

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
	Server      *string `json:"server,omitempty" cty:"server" hcl:"server,optional"`
}

func (c *MastodonCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	// Mastodon has no standard environment variable mentioned anywhere in the docs
	return env
}

func (c *MastodonCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
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

func (c *MastodonConnectionConfig) GetCredential(name string) Credential {

	mastodonCred := &MastodonCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		AccessToken: c.AccessToken,
		Server:      c.Server,
	}

	return mastodonCred
}
