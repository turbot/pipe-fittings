package credential

import (
	"context"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"golang.org/x/exp/maps"
	"log/slog"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
)

var credentialTypeRegistry = map[string]reflect.Type{
	"abuseipdb":     reflect.TypeOf((*AbuseIPDBCredential)(nil)).Elem(),
	"alicloud":      reflect.TypeOf((*AlicloudCredential)(nil)).Elem(),
	"aws":           reflect.TypeOf((*AwsCredential)(nil)).Elem(),
	"azure":         reflect.TypeOf((*AzureCredential)(nil)).Elem(),
	"bitbucket":     reflect.TypeOf((*BitbucketCredential)(nil)).Elem(),
	"clickup":       reflect.TypeOf((*ClickUpCredential)(nil)).Elem(),
	"datadog":       reflect.TypeOf((*DatadogCredential)(nil)).Elem(),
	"discord":       reflect.TypeOf((*DiscordCredential)(nil)).Elem(),
	"freshdesk":     reflect.TypeOf((*FreshdeskCredential)(nil)).Elem(),
	"gcp":           reflect.TypeOf((*GcpCredential)(nil)).Elem(),
	"github":        reflect.TypeOf((*GithubCredential)(nil)).Elem(),
	"gitlab":        reflect.TypeOf((*GitLabCredential)(nil)).Elem(),
	"guardrails":    reflect.TypeOf((*GuardrailsCredential)(nil)).Elem(),
	"ip2locationio": reflect.TypeOf((*IP2LocationIOCredential)(nil)).Elem(),
	"ipstack":       reflect.TypeOf((*IPstackCredential)(nil)).Elem(),
	"jira":          reflect.TypeOf((*JiraCredential)(nil)).Elem(),
	"jumpcloud":     reflect.TypeOf((*JumpCloudCredential)(nil)).Elem(),
	"mastodon":      reflect.TypeOf((*MastodonCredential)(nil)).Elem(),
	"okta":          reflect.TypeOf((*OktaCredential)(nil)).Elem(),
	"openai":        reflect.TypeOf((*OpenAICredential)(nil)).Elem(),
	"opsgenie":      reflect.TypeOf((*OpsgenieCredential)(nil)).Elem(),
	"pagerduty":     reflect.TypeOf((*PagerDutyCredential)(nil)).Elem(),
	"pipes":         reflect.TypeOf((*PipesCredential)(nil)).Elem(),
	"sendgrid":      reflect.TypeOf((*SendGridCredential)(nil)).Elem(),
	"servicenow":    reflect.TypeOf((*ServiceNowCredential)(nil)).Elem(),
	"slack":         reflect.TypeOf((*SlackCredential)(nil)).Elem(),
	"teams":         reflect.TypeOf((*MicrosoftTeamsCredential)(nil)).Elem(),
	"trello":        reflect.TypeOf((*TrelloCredential)(nil)).Elem(),
	"uptimerobot":   reflect.TypeOf((*UptimeRobotCredential)(nil)).Elem(),
	"urlscan":       reflect.TypeOf((*UrlscanCredential)(nil)).Elem(),
	"vault":         reflect.TypeOf((*VaultCredential)(nil)).Elem(),
	"virustotal":    reflect.TypeOf((*VirusTotalCredential)(nil)).Elem(),
	"zendesk":       reflect.TypeOf((*ZendeskCredential)(nil)).Elem(),
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

var credentialConfigTypeRegistry = map[string]reflect.Type{
	"abuseipdb":     reflect.TypeOf((*AbuseIPDBConnectionConfig)(nil)).Elem(),
	"alicloud":      reflect.TypeOf((*AlicloudConnectionConfig)(nil)).Elem(),
	"aws":           reflect.TypeOf((*AwsConnectionConfig)(nil)).Elem(),
	"azure":         reflect.TypeOf((*AzureConnectionConfig)(nil)).Elem(),
	"bitbucket":     reflect.TypeOf((*BitbucketConnectionConfig)(nil)).Elem(),
	"clickup":       reflect.TypeOf((*ClickUpConnectionConfig)(nil)).Elem(),
	"datadog":       reflect.TypeOf((*DatadogConnectionConfig)(nil)).Elem(),
	"discord":       reflect.TypeOf((*DiscordConnectionConfig)(nil)).Elem(),
	"freshdesk":     reflect.TypeOf((*FreshdeskConnectionConfig)(nil)).Elem(),
	"gcp":           reflect.TypeOf((*GcpConnectionConfig)(nil)).Elem(),
	"github":        reflect.TypeOf((*GithubConnectionConfig)(nil)).Elem(),
	"gitlab":        reflect.TypeOf((*GitlabConnectionConfig)(nil)).Elem(),
	"guardrails":    reflect.TypeOf((*GuardrailsConnectionConfig)(nil)).Elem(),
	"ip2locationio": reflect.TypeOf((*IP2LocationIOConnectionConfig)(nil)).Elem(),
	"ipstack":       reflect.TypeOf((*IPStackConnectionConfig)(nil)).Elem(),
	"jira":          reflect.TypeOf((*JiraConnectionConfig)(nil)).Elem(),
	"jumpcloud":     reflect.TypeOf((*JumpCloudConnectionConfig)(nil)).Elem(),
	"mastodon":      reflect.TypeOf((*MastodonConnectionConfig)(nil)).Elem(),
	"okta":          reflect.TypeOf((*OktaConnectionConfig)(nil)).Elem(),
	"openai":        reflect.TypeOf((*OpenAIConnectionConfig)(nil)).Elem(),
	"opsgenie":      reflect.TypeOf((*OpsgenieConnectionConfig)(nil)).Elem(),
	"pagerduty":     reflect.TypeOf((*PagerDutyConnectionConfig)(nil)).Elem(),
	"pipes":         reflect.TypeOf((*PipesConnectionConfig)(nil)).Elem(),
	"sendgrid":      reflect.TypeOf((*SendGridConnectionConfig)(nil)).Elem(),
	"servicenow":    reflect.TypeOf((*ServiceNowConnectionConfig)(nil)).Elem(),
	"slack":         reflect.TypeOf((*SlackConnectionConfig)(nil)).Elem(),
	"teams":         reflect.TypeOf((*MicrosoftTeamsConnectionConfig)(nil)).Elem(),
	"trello":        reflect.TypeOf((*TrelloConnectionConfig)(nil)).Elem(),
	"uptimerobot":   reflect.TypeOf((*UptimeRobotConnectionConfig)(nil)).Elem(),
	"urlscan":       reflect.TypeOf((*UrlscanConnectionConfig)(nil)).Elem(),
	"vault":         reflect.TypeOf((*VaultConnectionConfig)(nil)).Elem(),
	"virustotal":    reflect.TypeOf((*VirusTotalConnectionConfig)(nil)).Elem(),
	"zendesk":       reflect.TypeOf((*ZendeskConnectionConfig)(nil)).Elem(),
}

func InstantiateCredentialConfig(key string) (CredentialConfig, error) {
	// If the key has a slash, extract the last part of the key
	if strings.Contains(key, "/") {
		strParts := strings.Split(key, "/")
		key = strParts[len(strParts)-1]
	}

	t, exists := credentialConfigTypeRegistry[key]
	if !exists {
		// Currently the Flowpipe only supports very small set of credential types as compare to the Steampipe.
		// To avoid the parse error, skip the credential types that are not supported by the Flowpipe.
		slog.Error("Invalid credential type", "credential_type", key)
		return nil, nil
	}
	credConfigInterface := reflect.New(t).Interface()
	credConfig, ok := credConfigInterface.(CredentialConfig)
	if !ok {
		return nil, perr.InternalWithMessage("Failed to create credential config")
	}

	return credConfig, nil
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

		error_helpers.RegisterCredentialType(k)
	}

	return credentials, nil
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

type CredentialConfig interface {
	GetCredential(string, string) Credential
}

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

	Equals(Credential) bool
	GetCredentialImpl() CredentialImpl
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

func (c *CredentialImpl) GetCredentialImpl() CredentialImpl {
	return *c
}

func ctyValueForCredential(credential Credential) (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(credential)
	if err != nil {
		return cty.NilVal, err
	}
	impl := credential.GetCredentialImpl()
	baseCtyValue, err := cty_helpers.GetCtyValue(impl)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	mergedValueMap := baseCtyValue.AsValueMap()

	// copy into mergedValueMap, overriding base properties with derived properties if where there are clashes
	// we will return mergedValueMap
	maps.Copy(mergedValueMap, valueMap)

	mergedValueMap["env"] = cty.ObjectVal(credential.getEnv())
	mergedValueMap["type"] = cty.StringVal(credential.GetCredentialType())
	return cty.ObjectVal(mergedValueMap), nil
}

func CredentialToConnection(credential Credential) (connection.PipelingConnection, error) {
	ctyValue, err := credential.CtyValue()
	if err != nil {
		return nil, err
	}
	conn, err := app_specific_connection.CtyValueToConnection(ctyValue)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

//
//// ConnectionImportTarget implements the ConnectionImportTarget interface
//// marking this as a type that cvan be imported from a steampipe connection
//func (c *CredentialImpl) ConnectionImportTarget() {}
