package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const GithubConnectionType = "github"

type GithubConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *GithubConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &GithubConnection{ConnectionImpl: c.ConnectionImpl})
	}

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

func NewGithubConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &GithubConnection{
		ConnectionImpl: NewConnectionImpl(GithubConnectionType, shortName, declRange),
	}
}
func (c *GithubConnection) GetConnectionType() string {
	return GithubConnectionType
}

func (c *GithubConnection) Equals(otherConnection PipelingConnection) bool {
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

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *GithubConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.Token != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}
	return hcl.Diagnostics{}
}

func (c *GithubConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *GithubConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["GITHUB_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}
