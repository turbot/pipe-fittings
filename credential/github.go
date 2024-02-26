package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type GithubCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *GithubCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["GITHUB_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *GithubCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GithubCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*GithubCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *GithubCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		githubAccessTokenEnvVar := os.Getenv("GITHUB_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &GithubCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &githubAccessTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *GithubCredential) GetTtl() int {
	return -1
}

func (c *GithubCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type GithubConnectionConfig struct {
	Token *string `cty:"token" hcl:"token"`
}

func (c *GithubConnectionConfig) GetCredential(name string) Credential {

	githubCred := &GithubCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		Token: c.Token,
	}

	return githubCred
}
