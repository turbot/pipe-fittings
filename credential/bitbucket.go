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

type BitbucketCredential struct {
	CredentialImpl

	BaseURL  *string `json:"base_url,omitempty" cty:"base_url" hcl:"base_url,optional"`
	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func (c *BitbucketCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.BaseURL != nil {
		env["BITBUCKET_API_BASE_URL"] = cty.StringVal(*c.BaseURL)
	}
	if c.Username != nil {
		env["BITBUCKET_USERNAME"] = cty.StringVal(*c.Username)
	}
	if c.Password != nil {
		env["BITBUCKET_PASSWORD"] = cty.StringVal(*c.Password)
	}
	return env
}

func (c *BitbucketCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *BitbucketCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*BitbucketCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.BaseURL, other.BaseURL) {
		return false
	}

	if !utils.PtrEqual(c.Username, other.Username) {
		return false
	}

	if !utils.PtrEqual(c.Password, other.Password) {
		return false
	}

	return true
}

func (c *BitbucketCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Password == nil && c.BaseURL == nil && c.Username == nil {
		bitbucketURLEnvVar := os.Getenv("BITBUCKET_API_BASE_URL")
		bitbucketUsernameEnvVar := os.Getenv("BITBUCKET_USERNAME")
		bitbucketPasswordEnvVar := os.Getenv("BITBUCKET_PASSWORD")

		// Don't modify existing credential, resolve to a new one
		newCreds := &BitbucketCredential{
			CredentialImpl: c.CredentialImpl,
			Password:       &bitbucketPasswordEnvVar,
			BaseURL:        &bitbucketURLEnvVar,
			Username:       &bitbucketUsernameEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *BitbucketCredential) GetTtl() int {
	return -1
}

func (c *BitbucketCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type BitbucketConnectionConfig struct {
	BaseURL  *string `cty:"base_url" hcl:"base_url"`
	Password *string `cty:"password" hcl:"password"`
	Username *string `cty:"username" hcl:"username"`
}

func (c *BitbucketConnectionConfig) GetCredential(name string) Credential {

	bitbucketCred := &BitbucketCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		BaseURL:  c.BaseURL,
		Password: c.Password,
		Username: c.Username,
	}

	return bitbucketCred
}
