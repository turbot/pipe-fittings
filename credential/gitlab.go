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

type GitLabCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *GitLabCredential) getEnv() map[string]cty.Value {
	// There is no environment variable listed in the GitLab official API docs
	// https://github.com/xanzy/go-gitlab
	return nil
}

func (c *GitLabCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GitLabCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*GitLabCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *GitLabCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		gitlabAccessTokenEnvVar := os.Getenv("GITLAB_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &GitLabCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &gitlabAccessTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *GitLabCredential) GetTtl() int {
	return -1
}

func (c *GitLabCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type GitlabConnectionConfig struct {
	BaseUrl *string `cty:"baseurl" hcl:"baseurl,optional"`
	Token   *string `cty:"token" hcl:"token"`
}

func (c *GitlabConnectionConfig) GetCredential(name string, shortName string) Credential {

	gitlabCred := &GitLabCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "gitlab",
		},

		Token: c.Token,
	}

	return gitlabCred
}
