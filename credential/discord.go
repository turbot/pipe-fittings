package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type DiscordCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *DiscordCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["DISCORD_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *DiscordCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *DiscordCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		discordTokenEnvVar := os.Getenv("DISCORD_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &DiscordCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &discordTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *DiscordCredential) GetTtl() int {
	return -1
}

func (c *DiscordCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}
