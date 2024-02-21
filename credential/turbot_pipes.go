package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type PipesCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PipesCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PIPES_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *PipesCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PipesCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		pipesTokenEnvVar := os.Getenv("PIPES_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &PipesCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &pipesTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *PipesCredential) GetTtl() int {
	return -1
}

func (c *PipesCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type PipesConnectionConfig struct {
	Host  *string `cty:"host" hcl:"host,optional"`
	Token *string `cty:"token" hcl:"token"`
}

func (c *PipesConnectionConfig) GetCredential(name string) Credential {

	pipesCred := &PipesCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		Token: c.Token,
	}

	return pipesCred
}
