package modconfig

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/oauth2/google"
)

type Credential interface {
	HclResource
	ResourceWithMetadata

	GetCredentialType() string
	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (Credential, error)
	GetTtl() int // in seconds

	Validate() hcl.Diagnostics
	getEnv() map[string]cty.Value
}

type AwsCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	AccessKey    *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey    *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	SessionToken *string `json:"session_token,omitempty" cty:"session_token" hcl:"session_token,optional"`
	Ttl          *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`
	Profile      *string `json:"profile,omitempty" cty:"profile" hcl:"profile,optional"`
}

func DefaultCredentialNames() []string {
	return []string{"aws.default", "slack.default", "basic.default", "gcp.default", "abuseipdb.default", "sendgrid.default", "virustotal.default", "zendesk.default", "trello.default", "okta.default", "uptimerobot.default", "urlscan.default", "clickup.default", "pagerduty.default", "discord.default", "ip2location.default", "ipstack.default", "teams.default", "pipes.default", "github.default", "gitlab.default", "vault.default", "jira.default", "aws.<dynamic>", "slack.<dynamic>", "basic.<dynamic>", "gcp.<dynamic>", "abuseipdb.<dynamic>", "sendgrid.<dynamic>", "virustotal.<dynamic>", "zendesk.<dynamic>", "trello.<dynamic>", "okta.<dynamic>", "uptimerobot.<dynamic>", "urlscan.<dynamic>", "clickup.<dynamic>", "pagerduty.<dynamic>", "discord.<dynamic>", "ip2location.<dynamic>", "ipstack.<dynamic>", "teams.<dynamic>", "pipes.<dynamic>", "github.<dynamic>", "gitlab.<dynamic>", "vault.<dynamic>", "jira.<dynamic>"}
}

func (*AwsCredential) GetCredentialType() string {
	return "aws"
}

func (c *AwsCredential) Validate() hcl.Diagnostics {

	if c.AccessKey != nil && c.SecretKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "access_key defined without secret_key",
				Subject:  &c.DeclRange,
			},
		}
	}

	if c.SecretKey != nil && c.AccessKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "secret_key defined without access_key",
				Subject:  &c.DeclRange,
			},
		}
	}

	return hcl.Diagnostics{}
}

func (c *AwsCredential) Resolve(ctx context.Context) (Credential, error) {

	// if access key and secret key are provided, just return it
	if c.AccessKey != nil && c.SecretKey != nil {
		return c, nil
	}

	var cfg aws.Config
	var err error

	// Load the AWS configuration from the shared credentials file
	if c.Profile != nil {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(*c.Profile), config.WithRegion("us-east-1"))
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
		if err != nil {
			return nil, err
		}

	}

	// Access the credentials from the configuration
	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return nil, err
	}

	// Don't modify existing credential, resolve to a new one
	newCreds := &AwsCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        c.FullName,
			UnqualifiedName: c.UnqualifiedName,
			DeclRange:       c.DeclRange,
			blockType:       c.blockType,
		},
		Type:      c.Type,
		Ttl:       c.Ttl,
		AccessKey: &creds.AccessKeyID,
		SecretKey: &creds.SecretAccessKey,
	}

	if creds.SessionToken != "" {
		newCreds.SessionToken = &creds.SessionToken
	}

	return newCreds, nil
}

// in seconds
func (c *AwsCredential) GetTtl() int {
	if c.Ttl == nil {
		return 5 * 60
	}
	return *c.Ttl
}

func (c *AwsCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessKey != nil {
		env["AWS_ACCESS_KEY_ID"] = cty.StringVal(*c.AccessKey)
	}
	if c.SecretKey != nil {
		env["AWS_SECRET_ACCESS_KEY"] = cty.StringVal(*c.SecretKey)
	}
	if c.SessionToken != nil {
		env["AWS_SESSION_TOKEN"] = cty.StringVal(*c.SessionToken)
	}
	return env
}

func (c *AwsCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

type SlackCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*SlackCredential) GetCredentialType() string {
	return "slack"
}

func (c *SlackCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["SLACK_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *SlackCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *SlackCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.Token == nil {
		slackTokenEnvVar := os.Getenv("SLACK_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &SlackCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:  c.Type,
			Token: &slackTokenEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *SlackCredential) GetTtl() int {
	return -1
}

func (c *SlackCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type AbuseIPDBCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (*AbuseIPDBCredential) GetCredentialType() string {
	return "abuseipdb"
}

func (c *AbuseIPDBCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["ABUSEIPDB_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *AbuseIPDBCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *AbuseIPDBCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		abuseIPDBAPIKeyEnvVar := os.Getenv("ABUSEIPDB_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &AbuseIPDBCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			APIKey: &abuseIPDBAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *AbuseIPDBCredential) GetTtl() int {
	return -1
}

func (c *AbuseIPDBCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type SendGridCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (*SendGridCredential) GetCredentialType() string {
	return "sendgrid"
}

func (c *SendGridCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["SENDGRID_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *SendGridCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *SendGridCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		sendGridAPIKeyEnvVar := os.Getenv("SENDGRID_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &SendGridCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			APIKey: &sendGridAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *SendGridCredential) GetTtl() int {
	return -1
}

func (c *SendGridCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type VirusTotalCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (*VirusTotalCredential) GetCredentialType() string {
	return "virustotal"
}

func (c *VirusTotalCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["VTCLI_APIKEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *VirusTotalCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *VirusTotalCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		virusTotalAPIKeyEnvVar := os.Getenv("VTCLI_APIKEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &VirusTotalCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			APIKey: &virusTotalAPIKeyEnvVar,
		}

		return newCreds, nil

	}
	return c, nil
}

func (c *VirusTotalCredential) GetTtl() int {
	return -1
}

func (c *VirusTotalCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type ZendeskCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
	Email     *string `json:"email,omitempty" cty:"email" hcl:"email,optional"`
	Token     *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*ZendeskCredential) GetCredentialType() string {
	return "zendesk"
}

func (c *ZendeskCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Subdomain != nil {
		env["ZENDESK_SUBDOMAIN"] = cty.StringVal(*c.Subdomain)
	}
	if c.Email != nil {
		env["ZENDESK_EMAIL"] = cty.StringVal(*c.Email)
	}
	if c.Token != nil {
		env["ZENDESK_API_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *ZendeskCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ZendeskCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.Subdomain == nil && c.Email == nil && c.Token == nil {
		subdomainEnvVar := os.Getenv("ZENDESK_SUBDOMAIN")
		emailEnvVar := os.Getenv("ZENDESK_EMAIL")
		tokenEnvVar := os.Getenv("ZENDESK_API_TOKEN")

		// Don't modify existing credential, resolve to a new one
		newCreds := &ZendeskCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:      c.Type,
			Subdomain: &subdomainEnvVar,
			Email:     &emailEnvVar,
			Token:     &tokenEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *ZendeskCredential) GetTtl() int {
	return -1
}

func (c *ZendeskCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type TrelloCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*TrelloCredential) GetCredentialType() string {
	return "trello"
}

func (c *TrelloCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["TRELLO_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	if c.Token != nil {
		env["TRELLO_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *TrelloCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *TrelloCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.APIKey == nil && c.Token == nil {
		apiKeyEnvVar := os.Getenv("TRELLO_API_KEY")
		tokenEnvVar := os.Getenv("TRELLO_TOKEN")
		// Don't modify existing credential, resolve to a new one
		newCreds := &TrelloCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			APIKey: &apiKeyEnvVar,
			Token:  &tokenEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *TrelloCredential) GetTtl() int {
	return -1
}

func (c *TrelloCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type OktaCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	Domain *string `json:"domain,omitempty" cty:"domain" hcl:"domain,optional"`
}

func (*OktaCredential) GetCredentialType() string {
	return "okta"
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
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			Token:  &apiTokenEnvVar,
			Domain: &domainEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (*UptimeRobotCredential) GetCredentialType() string {
	return "uptimerobot"
}

func (c *UptimeRobotCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["UPTIMEROBOT_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *UptimeRobotCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			APIKey: &uptimeRobotAPIKeyEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (*UrlscanCredential) GetCredentialType() string {
	return "urlscan"
}

func (c *UrlscanCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["URLSCAN_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *UrlscanCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			APIKey: &urlscanAPIKeyEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*ClickUpCredential) GetCredentialType() string {
	return "clickup"
}

func (c *ClickUpCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["CLICKUP_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *ClickUpCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:  c.Type,
			Token: &clickUpAPITokenEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*PagerDutyCredential) GetCredentialType() string {
	return "pagerduty"
}

func (c *PagerDutyCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PAGERDUTY_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *PagerDutyCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:  c.Type,
			Token: &pagerDutyTokenEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*DiscordCredential) GetCredentialType() string {
	return "discord"
}

func (c *DiscordCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["DISCORD_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *DiscordCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:  c.Type,
			Token: &discordTokenEnvVar,
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

type IP2LocationCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (*IP2LocationCredential) GetCredentialType() string {
	return "ip2location"
}

func (c *IP2LocationCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["IP2LOCATION_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *IP2LocationCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IP2LocationCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		ip2locationAPIKeyEnvVar := os.Getenv("IP2LOCATION_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &IP2LocationCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:   c.Type,
			APIKey: &ip2locationAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *IP2LocationCredential) GetTtl() int {
	return -1
}

func (c *IP2LocationCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type IPstackCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
}

func (*IPstackCredential) GetCredentialType() string {
	return "ipstack"
}

func (c *IPstackCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessKey != nil {
		env["IPSTACK_ACCESS_KEY"] = cty.StringVal(*c.AccessKey)
	}
	return env
}

func (c *IPstackCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:      c.Type,
			AccessKey: &ipstackAccessKeyEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func (*MicrosoftTeamsCredential) GetCredentialType() string {
	return "teams"
}

func (c *MicrosoftTeamsCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessToken != nil {
		env["TEAMS_ACCESS_TOKEN"] = cty.StringVal(*c.AccessToken)
	}
	return env
}

func (c *MicrosoftTeamsCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:        c.Type,
			AccessToken: &msTeamsAccessTokenEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*PipesCredential) GetCredentialType() string {
	return "pipes"
}

func (c *PipesCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["PIPES_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *PipesCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:  c.Type,
			Token: &pipesTokenEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*GithubCredential) GetCredentialType() string {
	return "github"
}

func (c *GithubCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["GITHUB_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *GithubCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:  c.Type,
			Token: &githubAccessTokenEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*GitLabCredential) GetCredentialType() string {
	return "gitlab"
}

func (c *GitLabCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["GITLAB_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *GitLabCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:  c.Type,
			Token: &gitlabAccessTokenEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token   *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	Address *string `json:"address,omitempty" cty:"address" hcl:"address,optional"`
}

func (*VaultCredential) GetCredentialType() string {
	return "vault"
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
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:    c.Type,
			Token:   &tokenEnvVar,
			Address: &addressEnvVar,
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
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	APIToken *string `json:"api_token,omitempty" cty:"api_token" hcl:"api_token,optional"`
	BaseURL  *string `json:"base_url,omitempty" cty:"base_url" hcl:"base_url,optional"`
	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
}

func (*JiraCredential) GetCredentialType() string {
	return "jira"
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
	ctyValue, err := GetCtyValue(c)
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
			HclResourceImpl: HclResourceImpl{
				FullName:        c.FullName,
				UnqualifiedName: c.UnqualifiedName,
				ShortName:       c.ShortName,
				DeclRange:       c.DeclRange,
				blockType:       c.blockType,
			},
			Type:     c.Type,
			APIToken: &jiraAPITokenEnvVar,
			BaseURL:  &jiraURLEnvVar,
			Username: &jiraUserEnvVar,
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

type GcpCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Credentials *string `json:"credentials,omitempty" cty:"credentials" hcl:"credentials,optional"`
	Ttl         *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func (*GcpCredential) GetCredentialType() string {
	return "gcp"
}

func (c *GcpCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}

func (c *GcpCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
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

type BasicCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func (c *BasicCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}

func (*BasicCredential) GetCredentialType() string {
	return "basic"
}

func (c *BasicCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *BasicCredential) Resolve(ctx context.Context) (Credential, error) {
	return c, nil
}

func (c *BasicCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *BasicCredential) GetTtl() int {
	return -1
}

func DefaultCredentials() map[string]Credential {
	credentials := make(map[string]Credential)

	credentials["aws.default"] = &AwsCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "aws.default",
			ShortName:       "default",
			UnqualifiedName: "aws.default",
		},
		Type: "aws",
	}
	credentials["slack.default"] = &SlackCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "slack.default",
			ShortName:       "default",
			UnqualifiedName: "slack.default",
		},
		Type: "slack",
	}
	credentials["gcp.default"] = &GcpCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "gcp.default",
			ShortName:       "default",
			UnqualifiedName: "gcp.default",
		},
		Type: "gcp",
	}
	credentials["abuseipdb.default"] = &AbuseIPDBCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "abuseipdb.default",
			ShortName:       "default",
			UnqualifiedName: "abuseipdb.default",
		},
		Type: "abuseipdb",
	}

	credentials["sendgrid.default"] = &SendGridCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "sendgrid.default",
			ShortName:       "default",
			UnqualifiedName: "sendgrid.default",
		},
		Type: "sendgrid",
	}
	credentials["virustotal.default"] = &VirusTotalCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "virustotal.default",
			ShortName:       "default",
			UnqualifiedName: "virustotal.default",
		},
		Type: "virustotal",
	}
	credentials["zendesk.default"] = &ZendeskCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "zendesk.default",
			ShortName:       "default",
			UnqualifiedName: "zendesk.default",
		},
		Type: "zendesk",
	}
	credentials["trello.default"] = &TrelloCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "trello.default",
			ShortName:       "default",
			UnqualifiedName: "trello.default",
		},
		Type: "trello",
	}
	credentials["okta.default"] = &OktaCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "okta.default",
			ShortName:       "default",
			UnqualifiedName: "okta.default",
		},
		Type: "okta",
	}
	credentials["uptimerobot.default"] = &UptimeRobotCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "uptimerobot.default",
			ShortName:       "default",
			UnqualifiedName: "uptimerobot.default",
		},
		Type: "uptimerobot",
	}
	credentials["urlscan.default"] = &UrlscanCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "urlscan.default",
			ShortName:       "default",
			UnqualifiedName: "urlscan.default",
		},
		Type: "urlscan",
	}
	credentials["clickup.default"] = &ClickUpCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "clickup.default",
			ShortName:       "default",
			UnqualifiedName: "clickup.default",
		},
		Type: "clickup",
	}
	credentials["pagerduty.default"] = &PagerDutyCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "pagerduty.default",
			ShortName:       "default",
			UnqualifiedName: "pagerduty.default",
		},
		Type: "pagerduty",
	}
	credentials["discord.default"] = &DiscordCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "discord.default",
			ShortName:       "default",
			UnqualifiedName: "discord.default",
		},
		Type: "discord",
	}
	credentials["ip2location.default"] = &IP2LocationCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "ip2location.default",
			ShortName:       "default",
			UnqualifiedName: "ip2location.default",
		},
		Type: "ip2location",
	}
	credentials["ipstack.default"] = &IPstackCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "ipstack.default",
			ShortName:       "default",
			UnqualifiedName: "ipstack.default",
		},
		Type: "ipstack",
	}
	credentials["teams.default"] = &MicrosoftTeamsCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "teams.default",
			ShortName:       "default",
			UnqualifiedName: "teams.default",
		},
		Type: "teams",
	}
	credentials["pipes.default"] = &PipesCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "pipes.default",
			ShortName:       "default",
			UnqualifiedName: "pipes.default",
		},
		Type: "pipes",
	}
	credentials["github.default"] = &GithubCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "github.default",
			ShortName:       "default",
			UnqualifiedName: "github.default",
		},
		Type: "github",
	}
	credentials["gitlab.default"] = &GitLabCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "gitlab.default",
			ShortName:       "default",
			UnqualifiedName: "gitlab.default",
		},
		Type: "gitlab",
	}
	credentials["vault.default"] = &VaultCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "vault.default",
			ShortName:       "default",
			UnqualifiedName: "vault.default",
		},
		Type: "vault",
	}
	credentials["jira.default"] = &JiraCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "jira.default",
			ShortName:       "default",
			UnqualifiedName: "jira.default",
		},
		Type: "jira",
	}

	return credentials
}

func NewCredential(block *hcl.Block) Credential {

	credentialType := block.Labels[0]
	credentialName := block.Labels[1]

	credentialFullName := credentialType + "." + credentialName

	if credentialType == "aws" {
		credential := &AwsCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "aws",
		}
		return credential
	} else if credentialType == "basic" {
		credential := &BasicCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "basic",
		}
		return credential
	} else if credentialType == "slack" {
		credential := &SlackCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "slack",
		}
		return credential
	} else if credentialType == "abuseipdb" {
		credential := &AbuseIPDBCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "abuseipdb",
		}
		return credential
	} else if credentialType == "sendgrid" {
		credential := &SendGridCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "sendgrid",
		}
		return credential
	} else if credentialType == "virustotal" {
		credential := &VirusTotalCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "virustotal",
		}
		return credential
	} else if credentialType == "zendesk" {
		credential := &ZendeskCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "zendesk",
		}
		return credential
	} else if credentialType == "trello" {
		credential := &TrelloCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "trello",
		}
		return credential
	} else if credentialType == "okta" {
		credential := &OktaCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "okta",
		}
		return credential
	} else if credentialType == "uptimerobot" {
		credential := &UptimeRobotCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "uptimerobot",
		}
		return credential
	} else if credentialType == "urlscan" {
		credential := &UrlscanCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "urlscan",
		}
		return credential
	} else if credentialType == "clickup" {
		credential := &ClickUpCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "clickup",
		}
		return credential
	} else if credentialType == "pagerduty" {
		credential := &PagerDutyCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "pagerduty",
		}
		return credential
	} else if credentialType == "discord" {
		credential := &DiscordCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "discord",
		}
		return credential
	} else if credentialType == "ip2location" {
		credential := &IP2LocationCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "ip2location",
		}
		return credential
	} else if credentialType == "ipstack" {
		credential := &IPstackCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "ipstack",
		}
		return credential
	} else if credentialType == "teams" {
		credential := &MicrosoftTeamsCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "teams",
		}
		return credential
	} else if credentialType == "pipes" {
		credential := &PipesCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "pipes",
		}
		return credential
	} else if credentialType == "github" {
		credential := &GithubCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "github",
		}
		return credential
	} else if credentialType == "gitlab" {
		credential := &GitLabCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "gitlab",
		}
		return credential
	} else if credentialType == "vault" {
		credential := &VaultCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "vault",
		}
		return credential
	} else if credentialType == "jira" {
		credential := &JiraCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "jira",
		}
		return credential
	}

	return nil
}
