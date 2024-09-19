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
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PipesCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*PipesCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
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

func (c *PipesConnectionConfig) GetCredential(name string, shortName string) Credential {

	pipesCred := &PipesCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "pipes",
		},

		Token: c.Token,
	}

	return pipesCred
}
