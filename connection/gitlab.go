package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const GitLabConnectionType = "gitlab"

type GitLabConnection struct {
	ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewGitLabConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &GitLabConnection{
		ConnectionImpl: NewConnectionImpl(GitLabConnectionType, shortName, declRange),
	}
}
func (c *GitLabConnection) GetConnectionType() string {
	return GitLabConnectionType
}

func (c *GitLabConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

	if c.Token == nil {
		gitlabAccessTokenEnvVar := os.Getenv("GITLAB_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &GitLabConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &gitlabAccessTokenEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *GitLabConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	impl := c.GetConnectionImpl()
	if impl.Equals(otherConnection.GetConnectionImpl()) == false {
		return false
	}

	other, ok := otherConnection.(*GitLabConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *GitLabConnection) Validate() hcl.Diagnostics {
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

func (c *GitLabConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GitLabConnection) GetEnv() map[string]cty.Value {
	// There is no environment variable listed in the GitLab official API docs
	// https://github.com/xanzy/go-gitlab
	return nil
}
