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

type JiraCredential struct {
	CredentialImpl

	APIToken *string `json:"api_token,omitempty" cty:"api_token" hcl:"api_token,optional"`
	BaseURL  *string `json:"base_url,omitempty" cty:"base_url" hcl:"base_url,optional"`
	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
}

func (c *JiraCredential) getEnv() map[string]cty.Value {
	return nil
}

func (c *JiraCredential) CtyValue() (cty.Value, error) {
	return ctyValueForCredential(c)
}

func (c *JiraCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*JiraCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIToken, other.APIToken) {
		return false
	}

	if !utils.PtrEqual(c.BaseURL, other.BaseURL) {
		return false
	}

	if !utils.PtrEqual(c.Username, other.Username) {
		return false
	}

	return true
}

func (c *JiraCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIToken == nil && c.BaseURL == nil && c.Username == nil {
		// The order of precedence for the Jira API token environment variable
		// 1. JIRA_API_TOKEN
		// 2. JIRA_TOKEN
		jiraAPITokenEnvVar := os.Getenv("JIRA_TOKEN")
		if os.Getenv("JIRA_API_TOKEN") != "" {
			jiraAPITokenEnvVar = os.Getenv("JIRA_API_TOKEN")
		}

		jiraURLEnvVar := os.Getenv("JIRA_URL")
		jiraUserEnvVar := os.Getenv("JIRA_USER")

		// Don't modify existing credential, resolve to a new one
		newCreds := &JiraCredential{
			CredentialImpl: c.CredentialImpl,
			APIToken:       &jiraAPITokenEnvVar,
			BaseURL:        &jiraURLEnvVar,
			Username:       &jiraUserEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *JiraCredential) GetTtl() int {
	return -1
}

func (c *JiraCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type JiraConnectionConfig struct {
	APIToken            *string `cty:"api_token" hcl:"api_token,optional"`
	BaseURL             *string `cty:"base_url" hcl:"base_url,optional"`
	PersonalAccessToken *string `cty:"personal_access_token" hcl:"personal_access_token,optional"`
	Token               *string `cty:"token" hcl:"token,optional"`
	Username            *string `cty:"username" hcl:"username,optional"`
}

func (c *JiraConnectionConfig) GetCredential(name string, shortName string) Credential {

	// Steampipe Jira plugin uses the attribute token to configure the credential, whereas
	// the Flowpipe uses the attribute `api_token` which is intended to distinguish between 2 different token, i.e. token and personal_access_token
	// Hence, we need a special handling to support both token and api_token, and the order of precedence will be api_token and token.
	var jiraAPIToken string
	if c.Token != nil {
		jiraAPIToken = *c.Token
	}
	if c.APIToken != nil {
		jiraAPIToken = *c.APIToken
	}

	jiraCred := &JiraCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "jira",
		},

		// In Flowpipe we went with api_token (same as token in Steampipe) since
		// there is another type of token (personal access token)
		// In future we could update the Steampipe plugin arg too, but not necessary right now.
		APIToken: &jiraAPIToken,
		BaseURL:  c.BaseURL,
		Username: c.Username,
	}

	return jiraCred
}
