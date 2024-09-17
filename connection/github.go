package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type GithubConnection struct {
	modconfig.ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *GithubConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
	if c.Token == nil {
		githubAccessTokenEnvVar := os.Getenv("GITHUB_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &GithubConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &githubAccessTokenEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *GithubConnection) GetConnectionType() string {
	return "github"
}

func (c *GithubConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*GithubConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *GithubConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *GithubConnection) GetTtl() int {
	return -1
}

func (c *GithubConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GithubConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["GITHUB_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}
