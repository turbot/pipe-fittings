package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type GitLabCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *GitLabCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["GITLAB_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
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

func (c *GitlabConnectionConfig) GetCredential(name string) Credential {

	gitlabCred := &GitLabCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		Token: c.Token,
	}

	return gitlabCred
}
