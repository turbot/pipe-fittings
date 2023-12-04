package modconfig

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
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
	return []string{"aws.default", "slack.default", "basic.default", "gcp.default", "abuseipdb.default", "sendgrid.default", "virustotal.default", "zendesk.default", "trello.default", "okta.default", "uptimerobot.default", "urlscan.default", "clickup.default", "pagerduty.default", "discord.default", "ip2location.default", "aws.<dynamic>", "slack.<dynamic>", "basic.<dynamic>", "gcp.<dynamic>", "abuseipdb.<dynamic>", "sendgrid.<dynamic>", "virustotal.<dynamic>", "zendesk.<dynamic>", "trello.<dynamic>", "okta.<dynamic>", "uptimerobot.<dynamic>", "urlscan.<dynamic>", "clickup.<dynamic>", "pagerduty.<dynamic>", "discord.<dynamic>", "ip2location.<dynamic>"}
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
		env["ZENDESK_USER"] = cty.StringVal(*c.Email)
	}
	if c.Token != nil {
		env["ZENDESK_TOKEN"] = cty.StringVal(*c.Token)
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
		emailEnvVar := os.Getenv("ZENDESK_USER")
		tokenEnvVar := os.Getenv("ZENDESK_TOKEN")

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

	APIToken *string `json:"api_token,omitempty" cty:"api_token" hcl:"api_token,optional"`
	Domain   *string `json:"domain,omitempty" cty:"domain" hcl:"domain,optional"`
}

func (*OktaCredential) GetCredentialType() string {
	return "okta"
}

func (c *OktaCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIToken != nil {
		env["OKTA_TOKEN"] = cty.StringVal(*c.APIToken)
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

	if c.APIToken == nil && c.Domain == nil {
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
			Type:     c.Type,
			APIToken: &apiTokenEnvVar,
			Domain:   &domainEnvVar,
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

	APIToken *string `json:"api_token,omitempty" cty:"api_token" hcl:"api_token,optional"`
}

func (*ClickUpCredential) GetCredentialType() string {
	return "clickup"
}

func (c *ClickUpCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIToken != nil {
		env["CLICKUP_TOKEN"] = cty.StringVal(*c.APIToken)
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
	if c.APIToken == nil {
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
			Type:     c.Type,
			APIToken: &clickUpAPITokenEnvVar,
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

type GcpCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Credentials *string `json:"credentials,omitempty" cty:"credentials" hcl:"credentials,optional"`
	Ttl         *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`
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
	return c, nil
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
	}

	return nil
}
