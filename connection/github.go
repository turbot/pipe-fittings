package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type GithubConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *GithubConnection) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["GITHUB_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *GithubConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GithubConnection) Equals(otherCredential PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*GithubConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *GithubConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.Token == nil {
		githubAccessTokenEnvVar := os.Getenv("GITHUB_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &GithubConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &githubAccessTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *GithubConnection) GetTtl() int {
	return -1
}

func (c *GithubConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type GithubConnectionConfig struct {
	Token *string `cty:"token" hcl:"token"`
}

func (c *GithubConnectionConfig) GetCredential(name string, shortName string) PipelingConnection {

	githubCred := &GithubConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "github",
		},

		Token: c.Token,
	}

	return githubCred
}
