package credential

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/oauth2/google"
)

type Credential interface {
	modconfig.HclResource
	modconfig.ResourceWithMetadata

	SetHclResourceImpl(hclResourceImpl modconfig.HclResourceImpl)
	GetCredentialType() string
	SetCredentialType(string)
	GetUnqualifiedName() string

	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (Credential, error)
	GetTtl() int // in seconds

	Validate() hcl.Diagnostics
	getEnv() map[string]cty.Value
}

type CredentialImpl struct {
	modconfig.HclResourceImpl
	modconfig.ResourceWithMetadataImpl

	// required to allow partial decoding
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`
}

func (c *CredentialImpl) GetUnqualifiedName() string {
	return c.HclResourceImpl.UnqualifiedName
}

func (c *CredentialImpl) SetHclResourceImpl(hclResourceImpl modconfig.HclResourceImpl) {
	c.HclResourceImpl = hclResourceImpl
}

func (c *CredentialImpl) GetCredentialType() string {
	return c.Type
}

func (c *CredentialImpl) SetCredentialType(credType string) {
	c.Type = credType
}

// TODO:
// TODO: move these to individual file
// TODO:

type OktaCredential struct {
	CredentialImpl

	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	Domain *string `json:"domain,omitempty" cty:"domain" hcl:"domain,optional"`
}

func (c *OktaCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["OKTA_TOKEN"] = cty.StringVal(*c.Token)
	}
	if c.Domain != nil {
		env["OKTA_ORGURL"] = cty.StringVal(*c.Domain)
	}
	return env
}

func (c *OktaCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OktaCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.Token == nil && c.Domain == nil {
		apiTokenEnvVar := os.Getenv("OKTA_TOKEN")
		domainEnvVar := os.Getenv("OKTA_ORGURL")

		// Don't modify existing credential, resolve to a new one
		newCreds := &OktaCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &apiTokenEnvVar,
			Domain:         &domainEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *OktaCredential) GetTtl() int {
	return -1
}

func (c *OktaCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type UptimeRobotCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *UptimeRobotCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["UPTIMEROBOT_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *UptimeRobotCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *UptimeRobotCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		uptimeRobotAPIKeyEnvVar := os.Getenv("UPTIMEROBOT_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &UptimeRobotCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &uptimeRobotAPIKeyEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *UptimeRobotCredential) GetTtl() int {
	return -1
}

func (c *UptimeRobotCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type UrlscanCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *UrlscanCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["URLSCAN_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *UrlscanCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *UrlscanCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		urlscanAPIKeyEnvVar := os.Getenv("URLSCAN_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &UrlscanCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &urlscanAPIKeyEnvVar,
		}
		return newCreds, nil
	}

	return c, nil
}

func (c *UrlscanCredential) GetTtl() int {
	return -1
}

func (c *UrlscanCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type ClickUpCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *ClickUpCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["CLICKUP_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *ClickUpCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ClickUpCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		clickUpAPITokenEnvVar := os.Getenv("CLICKUP_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &ClickUpCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &clickUpAPITokenEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *ClickUpCredential) GetTtl() int {
	return -1
}

func (c *ClickUpCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type PagerDutyCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PagerDutyCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PAGERDUTY_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *PagerDutyCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PagerDutyCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		pagerDutyTokenEnvVar := os.Getenv("PAGERDUTY_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &PagerDutyCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &pagerDutyTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *PagerDutyCredential) GetTtl() int {
	return -1
}

func (c *PagerDutyCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type DiscordCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *DiscordCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["DISCORD_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *DiscordCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *DiscordCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		discordTokenEnvVar := os.Getenv("DISCORD_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &DiscordCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &discordTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *DiscordCredential) GetTtl() int {
	return -1
}

func (c *DiscordCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type IP2LocationIOCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *IP2LocationIOCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["IP2LOCATIONIO_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *IP2LocationIOCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IP2LocationIOCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		ip2locationAPIKeyEnvVar := os.Getenv("IP2LOCATIONIO_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &IP2LocationIOCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &ip2locationAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *IP2LocationIOCredential) GetTtl() int {
	return -1
}

func (c *IP2LocationIOCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type IPstackCredential struct {
	CredentialImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
}

func (c *IPstackCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessKey != nil {
		env["IPSTACK_ACCESS_KEY"] = cty.StringVal(*c.AccessKey)
	}
	return env
}

func (c *IPstackCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IPstackCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.AccessKey == nil {
		ipstackAccessKeyEnvVar := os.Getenv("IPSTACK_ACCESS_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &IPstackCredential{
			CredentialImpl: c.CredentialImpl,
			AccessKey:      &ipstackAccessKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *IPstackCredential) GetTtl() int {
	return -1
}

func (c *IPstackCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type MicrosoftTeamsCredential struct {
	CredentialImpl

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func (c *MicrosoftTeamsCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessToken != nil {
		env["TEAMS_ACCESS_TOKEN"] = cty.StringVal(*c.AccessToken)
	}
	return env
}

func (c *MicrosoftTeamsCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *MicrosoftTeamsCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.AccessToken == nil {
		msTeamsAccessTokenEnvVar := os.Getenv("TEAMS_ACCESS_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &MicrosoftTeamsCredential{
			CredentialImpl: c.CredentialImpl,
			AccessToken:    &msTeamsAccessTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *MicrosoftTeamsCredential) GetTtl() int {
	return -1
}

func (c *MicrosoftTeamsCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type PipesCredential struct {
	CredentialImpl

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *PipesCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PIPES_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *PipesCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *PipesCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		pipesTokenEnvVar := os.Getenv("PIPES_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &PipesCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &pipesTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *PipesCredential) GetTtl() int {
	return -1
}

func (c *PipesCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

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

type VaultCredential struct {
	CredentialImpl

	Token   *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	Address *string `json:"address,omitempty" cty:"address" hcl:"address,optional"`
}

func (c *VaultCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["VAULT_TOKEN"] = cty.StringVal(*c.Token)
	}
	if c.Address != nil {
		env["VAULT_ADDR"] = cty.StringVal(*c.Address)
	}
	return env
}

func (c *VaultCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *VaultCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.Token == nil && c.Address == nil {
		tokenEnvVar := os.Getenv("VAULT_TOKEN")
		addressEnvVar := os.Getenv("VAULT_ADDR")

		// Don't modify existing credential, resolve to a new one
		newCreds := &VaultCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &tokenEnvVar,
			Address:        &addressEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *VaultCredential) GetTtl() int {
	return -1
}

func (c *VaultCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type JiraCredential struct {
	CredentialImpl

	APIToken *string `json:"api_token,omitempty" cty:"api_token" hcl:"api_token,optional"`
	BaseURL  *string `json:"base_url,omitempty" cty:"base_url" hcl:"base_url,optional"`
	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
}

func (c *JiraCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIToken != nil {
		env["JIRA_API_TOKEN"] = cty.StringVal(*c.APIToken)
	}
	if c.BaseURL != nil {
		env["JIRA_URL"] = cty.StringVal(*c.BaseURL)
	}
	if c.Username != nil {
		env["JIRA_USER"] = cty.StringVal(*c.Username)
	}
	return env
}

func (c *JiraCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *JiraCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIToken == nil && c.BaseURL == nil && c.Username == nil {
		jiraAPITokenEnvVar := os.Getenv("JIRA_API_TOKEN")
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

type OpsgenieCredential struct {
	CredentialImpl

	AlertAPIKey    *string `json:"alert_api_key,omitempty" cty:"alert_api_key" hcl:"alert_api_key,optional"`
	IncidentAPIKey *string `json:"incident_api_key,omitempty" cty:"incident_api_key" hcl:"incident_api_key,optional"`
}

func (c *OpsgenieCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AlertAPIKey != nil {
		env["OPSGENIE_ALERT_API_KEY"] = cty.StringVal(*c.AlertAPIKey)
	}
	if c.IncidentAPIKey != nil {
		env["OPSGENIE_INCIDENT_API_KEY"] = cty.StringVal(*c.IncidentAPIKey)
	}
	return env
}

func (c *OpsgenieCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OpsgenieCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.AlertAPIKey == nil && c.IncidentAPIKey == nil {
		alertAPIKeyEnvVar := os.Getenv("OPSGENIE_ALERT_API_KEY")
		incidentAPIKeyEnvVar := os.Getenv("OPSGENIE_INCIDENT_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &OpsgenieCredential{
			CredentialImpl: c.CredentialImpl,
			AlertAPIKey:    &alertAPIKeyEnvVar,
			IncidentAPIKey: &incidentAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *OpsgenieCredential) GetTtl() int {
	return -1
}

func (c *OpsgenieCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type OpenAICredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *OpenAICredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["OPENAI_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *OpenAICredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OpenAICredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		apiKeyEnvVar := os.Getenv("OPENAI_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &OpenAICredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &apiKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *OpenAICredential) GetTtl() int {
	return -1
}

func (c *OpenAICredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type GcpCredential struct {
	CredentialImpl

	Credentials *string `json:"credentials,omitempty" cty:"credentials" hcl:"credentials,optional"`
	Ttl         *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func (c *GcpCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}

func (c *GcpCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GcpCredential) Resolve(ctx context.Context) (Credential, error) {

	// First check if the credential file is supplied
	var credentialFile string
	if c.Credentials != nil && *c.Credentials != "" {
		credentialFile = *c.Credentials
	} else {
		credentialFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credentialFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, perr.InternalWithMessage("failed to get user home directory " + err.Error())
			}

			// If not, check if the default credential file exists
			credentialFile = filepath.Join(homeDir, ".config/gcloud/application_default_credentials.json")
		}
	}

	if credentialFile == "" {
		return c, nil
	}

	// Try to resolve this credential file
	creds, err := os.ReadFile(credentialFile)
	if err != nil {
		return nil, perr.InternalWithMessage("failed to read credential file " + err.Error())
	}

	var credData map[string]interface{}
	if err := json.Unmarshal(creds, &credData); err != nil {
		return nil, perr.InternalWithMessage("failed to parse credential file " + err.Error())
	}

	// Service Account / Authorized User flow
	if credData["type"] == "service_account" || credData["type"] == "authorized_user" {
		// Get a token source using the service account key file

		credentialParam := google.CredentialsParams{
			Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
		}

		credentials, err := google.CredentialsFromJSONWithParams(context.TODO(), creds, credentialParam)
		if err != nil {
			return nil, perr.InternalWithMessage("failed to get credentials from JSON " + err.Error())
		}

		tokenSource := credentials.TokenSource

		// Get the token
		token, err := tokenSource.Token()
		if err != nil {
			return nil, perr.InternalWithMessage("failed to get token from token source " + err.Error())
		}

		newCreds := &GcpCredential{
			AccessToken: &token.AccessToken,
			Credentials: &credentialFile,
		}
		return newCreds, nil
	}

	// oauth2 flow (untested)
	config, err := google.ConfigFromJSON(creds, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, perr.InternalWithMessage("failed to get config from JSON " + err.Error())
	}

	token, err := config.Exchange(context.Background(), "authorization-code")
	if err != nil {
		return nil, perr.InternalWithMessage("failed to get token from config " + err.Error())
	}

	newCreds := &GcpCredential{
		AccessToken: &token.AccessToken,
		Credentials: &credentialFile,
	}
	return newCreds, nil
}

func (c *GcpCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *GcpCredential) GetTtl() int {
	if c.Ttl == nil {
		return 5 * 60 // in seconds
	}
	return *c.Ttl
}

type AzureCredential struct {
	CredentialImpl

	ClientID     *string `json:"client_id,omitempty" cty:"client_id" hcl:"client_id,optional"`
	ClientSecret *string `json:"client_secret,omitempty" cty:"client_secret" hcl:"client_secret,optional"`
	TenantID     *string `json:"tenant_id,omitempty" cty:"tenant_id" hcl:"tenant_id,optional"`
	Environment  *string `json:"environment,omitempty" cty:"environment" hcl:"environment,optional"`
}

func (c *AzureCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.ClientID != nil {
		env["AZURE_CLIENT_ID"] = cty.StringVal(*c.ClientID)
	}
	if c.ClientSecret != nil {
		env["AZURE_CLIENT_SECRET"] = cty.StringVal(*c.ClientSecret)
	}
	if c.TenantID != nil {
		env["AZURE_TENANT_ID"] = cty.StringVal(*c.TenantID)
	}
	if c.Environment != nil {
		env["AZURE_ENVIRONMENT"] = cty.StringVal(*c.Environment)
	}
	return env
}

func (c *AzureCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *AzureCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.ClientID == nil && c.ClientSecret == nil && c.TenantID == nil && c.Environment == nil {
		clientIDEnvVar := os.Getenv("AZURE_CLIENT_ID")
		clientSecretEnvVar := os.Getenv("AZURE_CLIENT_SECRET")
		tenantIDEnvVar := os.Getenv("AZURE_TENANT_ID")
		environmentEnvVar := os.Getenv("AZURE_ENVIRONMENT")

		// Don't modify existing credential, resolve to a new one
		newCreds := &AzureCredential{
			CredentialImpl: c.CredentialImpl,
			ClientID:       &clientIDEnvVar,
			ClientSecret:   &clientSecretEnvVar,
			TenantID:       &tenantIDEnvVar,
			Environment:    &environmentEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *AzureCredential) GetTtl() int {
	return -1
}

func (c *AzureCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

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

type DatadogCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	AppKey *string `json:"app_key,omitempty" cty:"app_key" hcl:"app_key,optional"`
	APIUrl *string `json:"api_url,omitempty" cty:"api_url" hcl:"api_url,optional"`
}

func (c *DatadogCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["DD_CLIENT_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	if c.AppKey != nil {
		env["DD_CLIENT_APP_KEY"] = cty.StringVal(*c.AppKey)
	}
	return env
}

func (c *DatadogCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *DatadogCredential) Resolve(ctx context.Context) (Credential, error) {
	datadogAPIKeyEnvVar := os.Getenv("DD_CLIENT_API_KEY")
	datadogAppKeyEnvVar := os.Getenv("DD_CLIENT_APP_KEY")

	// Don't modify existing credential, resolve to a new one
	newCreds := &DatadogCredential{
		CredentialImpl: c.CredentialImpl,
		APIUrl:         c.APIUrl,
	}

	if c.APIKey == nil {
		newCreds.APIKey = &datadogAPIKeyEnvVar
	} else {
		newCreds.APIKey = c.APIKey
	}

	if c.AppKey == nil {
		newCreds.AppKey = &datadogAppKeyEnvVar
	} else {
		newCreds.AppKey = c.AppKey
	}

	return newCreds, nil
}

func (c *DatadogCredential) GetTtl() int {
	return -1
}

func (c *DatadogCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type FreshdeskCredential struct {
	CredentialImpl

	APIKey    *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
}

func (c *FreshdeskCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["FRESHDESK_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	if c.Subdomain != nil {
		env["FRESHDESK_SUBDOMAIN"] = cty.StringVal(*c.Subdomain)
	}
	return env
}

func (c *FreshdeskCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *FreshdeskCredential) Resolve(ctx context.Context) (Credential, error) {
	freshdeskAPIKeyEnvVar := os.Getenv("FRESHDESK_API_KEY")
	freshdeskSubdomainEnvVar := os.Getenv("FRESHDESK_SUBDOMAIN")

	// Don't modify existing credential, resolve to a new one
	newCreds := &FreshdeskCredential{
		CredentialImpl: c.CredentialImpl,
	}

	if c.APIKey == nil {
		newCreds.APIKey = &freshdeskAPIKeyEnvVar
	} else {
		newCreds.APIKey = c.APIKey
	}

	if c.Subdomain == nil {
		newCreds.Subdomain = &freshdeskSubdomainEnvVar
	} else {
		newCreds.Subdomain = c.Subdomain
	}

	return newCreds, nil
}

func (c *FreshdeskCredential) GetTtl() int {
	return -1
}

func (c *FreshdeskCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type GuardrailsCredential struct {
	CredentialImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	Workspace *string `json:"workspace,omitempty" cty:"workspace" hcl:"workspace,optional"`
}

func (c *GuardrailsCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessKey != nil {
		env["TURBOT_ACCESS_KEY"] = cty.StringVal(*c.AccessKey)
	}
	if c.SecretKey != nil {
		env["TURBOT_SECRET_KEY"] = cty.StringVal(*c.SecretKey)
	}
	if c.Workspace != nil {
		env["TURBOT_WORKSPACE"] = cty.StringVal(*c.Workspace)
	}
	return env
}

func (c *GuardrailsCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GuardrailsCredential) Resolve(ctx context.Context) (Credential, error) {
	guardrailsAccessKeyEnvVar := os.Getenv("TURBOT_ACCESS_KEY")
	guardrailsSecretKeyEnvVar := os.Getenv("TURBOT_SECRET_KEY")
	guardrailsWorkspaceEnvVar := os.Getenv("TURBOT_WORKSPACE")

	// Don't modify existing credential, resolve to a new one
	newCreds := &GuardrailsCredential{
		CredentialImpl: c.CredentialImpl,
		Workspace:      c.Workspace,
	}

	if c.AccessKey == nil {
		newCreds.AccessKey = &guardrailsAccessKeyEnvVar
	} else {
		newCreds.AccessKey = c.AccessKey
	}

	if c.SecretKey == nil {
		newCreds.SecretKey = &guardrailsSecretKeyEnvVar
	} else {
		newCreds.SecretKey = c.SecretKey
	}

	if c.Workspace == nil {
		newCreds.Workspace = &guardrailsWorkspaceEnvVar
	} else {
		newCreds.Workspace = c.Workspace
	}

	return newCreds, nil
}

func (c *GuardrailsCredential) GetTtl() int {
	return -1
}

func (c *GuardrailsCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type ServiceNowCredential struct {
	CredentialImpl

	InstanceURL *string `json:"instance_url,omitempty" cty:"instance_url" hcl:"instance_url,optional"`
	Username    *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password    *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func (c *ServiceNowCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.InstanceURL != nil {
		env["SERVICENOW_INSTANCE_URL"] = cty.StringVal(*c.InstanceURL)
	}
	if c.Username != nil {
		env["SERVICENOW_USERNAME"] = cty.StringVal(*c.Username)
	}
	if c.Password != nil {
		env["SERVICENOW_PASSWORD"] = cty.StringVal(*c.Password)
	}
	return env
}

func (c *ServiceNowCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ServiceNowCredential) Resolve(ctx context.Context) (Credential, error) {
	servicenowInstanceURLEnvVar := os.Getenv("SERVICENOW_INSTANCE_URL")
	servicenowUsernameEnvVar := os.Getenv("SERVICENOW_USERNAME")
	servicenowPasswordEnvVar := os.Getenv("SERVICENOW_PASSWORD")

	// Don't modify existing credential, resolve to a new one
	newCreds := &ServiceNowCredential{
		CredentialImpl: c.CredentialImpl,
	}

	if c.InstanceURL == nil {
		newCreds.InstanceURL = &servicenowInstanceURLEnvVar
	} else {
		newCreds.InstanceURL = c.InstanceURL
	}

	if c.Username == nil {
		newCreds.Username = &servicenowUsernameEnvVar
	} else {
		newCreds.Username = c.Username
	}

	if c.Password == nil {
		newCreds.Password = &servicenowPasswordEnvVar
	} else {
		newCreds.Password = c.Password
	}

	return newCreds, nil
}

func (c *ServiceNowCredential) GetTtl() int {
	return -1
}

func (c *ServiceNowCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type JumpCloudCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *JumpCloudCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["JUMPCLOUD_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *JumpCloudCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *JumpCloudCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		apiKeyEnvVar := os.Getenv("JUMPCLOUD_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &JumpCloudCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &apiKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *JumpCloudCredential) GetTtl() int {
	return -1
}

func (c *JumpCloudCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func DefaultCredentials() (map[string]Credential, error) {
	credentials := make(map[string]Credential)

	for k := range credentialTypeRegistry {
		hclResourceImpl := modconfig.HclResourceImpl{
			FullName:        k + ".default",
			ShortName:       "default",
			UnqualifiedName: k + ".default",
		}

		defaultCred, err := instantiateCredential(k, hclResourceImpl)
		if err != nil {
			return nil, err
		}

		credentials[k+".default"] = defaultCred
	}

	return credentials, nil
}

var credentialTypeRegistry = map[string]reflect.Type{
	"aws":           reflect.TypeOf((*AwsCredential)(nil)).Elem(),
	"slack":         reflect.TypeOf((*SlackCredential)(nil)).Elem(),
	"abuseipdb":     reflect.TypeOf((*AbuseIPDBCredential)(nil)).Elem(),
	"sendgrid":      reflect.TypeOf((*SendGridCredential)(nil)).Elem(),
	"virustotal":    reflect.TypeOf((*VirusTotalCredential)(nil)).Elem(),
	"zendesk":       reflect.TypeOf((*ZendeskCredential)(nil)).Elem(),
	"trello":        reflect.TypeOf((*TrelloCredential)(nil)).Elem(),
	"okta":          reflect.TypeOf((*OktaCredential)(nil)).Elem(),
	"uptimerobot":   reflect.TypeOf((*UptimeRobotCredential)(nil)).Elem(),
	"urlscan":       reflect.TypeOf((*UrlscanCredential)(nil)).Elem(),
	"clickup":       reflect.TypeOf((*ClickUpCredential)(nil)).Elem(),
	"pagerduty":     reflect.TypeOf((*PagerDutyCredential)(nil)).Elem(),
	"discord":       reflect.TypeOf((*DiscordCredential)(nil)).Elem(),
	"ip2locationio": reflect.TypeOf((*IP2LocationIOCredential)(nil)).Elem(),
	"ipstack":       reflect.TypeOf((*IPstackCredential)(nil)).Elem(),
	"teams":         reflect.TypeOf((*MicrosoftTeamsCredential)(nil)).Elem(),
	"pipes":         reflect.TypeOf((*PipesCredential)(nil)).Elem(),
	"github":        reflect.TypeOf((*GithubCredential)(nil)).Elem(),
	"gitlab":        reflect.TypeOf((*GitLabCredential)(nil)).Elem(),
	"vault":         reflect.TypeOf((*VaultCredential)(nil)).Elem(),
	"jira":          reflect.TypeOf((*JiraCredential)(nil)).Elem(),
	"opsgenie":      reflect.TypeOf((*OpsgenieCredential)(nil)).Elem(),
	"openai":        reflect.TypeOf((*OpenAICredential)(nil)).Elem(),
	"azure":         reflect.TypeOf((*AzureCredential)(nil)).Elem(),
	"gcp":           reflect.TypeOf((*GcpCredential)(nil)).Elem(),
	"bitbucket":     reflect.TypeOf((*BitbucketCredential)(nil)).Elem(),
	"datadog":       reflect.TypeOf((*DatadogCredential)(nil)).Elem(),
	"freshdesk":     reflect.TypeOf((*FreshdeskCredential)(nil)).Elem(),
	"guardrails":    reflect.TypeOf((*GuardrailsCredential)(nil)).Elem(),
	"servicenow":    reflect.TypeOf((*ServiceNowCredential)(nil)).Elem(),
	"jumpcloud":     reflect.TypeOf((*JumpCloudCredential)(nil)).Elem(),
}

func instantiateCredential(key string, hclResourceImpl modconfig.HclResourceImpl) (Credential, error) {
	t, exists := credentialTypeRegistry[key]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid credential type " + key)
	}
	credInterface := reflect.New(t).Interface()
	cred, ok := credInterface.(Credential)
	if !ok {
		return nil, perr.InternalWithMessage("Failed to create credential")
	}
	cred.SetHclResourceImpl(hclResourceImpl)
	cred.SetCredentialType(key)

	return cred, nil
}

func NewCredential(block *hcl.Block) (Credential, error) {
	credentialType := block.Labels[0]
	credentialName := block.Labels[1]

	hclResourceImpl := modconfig.NewHclResourceImplNoMod(block, credentialType, credentialName)

	credential, err := instantiateCredential(credentialType, hclResourceImpl)
	if err != nil {
		return nil, err
	}

	return credential, err
}
