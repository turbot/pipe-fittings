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

type GitLabConnection struct {
	modconfig.ConnectionImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *GitLabConnection) GetConnectionType() string {
	return "gitlab"
}

func (c *GitLabConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
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

func (c *GitLabConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
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
	return hcl.Diagnostics{}
}

func (c *GitLabConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GitLabConnection) GetTtl() int {
	return -1
}

func (c *GitLabConnection) GetEnv() map[string]cty.Value {
	// There is no environment variable listed in the GitLab official API docs
	// https://github.com/xanzy/go-gitlab
	return nil
}
