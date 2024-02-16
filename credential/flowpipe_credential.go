package credential

import (
	"context"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
)

var credentialTypeRegistry = map[string]reflect.Type{
	"abuseipdb":     reflect.TypeOf((*AbuseIPDBCredential)(nil)).Elem(),
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
	"aws":           reflect.TypeOf((*AwsConnectionConfig)(nil)).Elem(),
	"clickup":       reflect.TypeOf((*ClickUpConnectionConfig)(nil)).Elem(),
	"discord":       reflect.TypeOf((*DiscordConnectionConfig)(nil)).Elem(),
	"github":        reflect.TypeOf((*GithubConnectionConfig)(nil)).Elem(),
	"gitlab":        reflect.TypeOf((*GitlabConnectionConfig)(nil)).Elem(),
	"ip2locationio": reflect.TypeOf((*IP2LocationIOConnectionConfig)(nil)).Elem(),
	"jumpcloud":     reflect.TypeOf((*JumpCloudConnectionConfig)(nil)).Elem(),
	"openai":        reflect.TypeOf((*OpenAIConnectionConfig)(nil)).Elem(),
	"pagerduty":     reflect.TypeOf((*PagerDutyConnectionConfig)(nil)).Elem(),
	"pipes":         reflect.TypeOf((*PipesConnectionConfig)(nil)).Elem(),
	"sendgrid":      reflect.TypeOf((*SendGridConnectionConfig)(nil)).Elem(),
	"slack":         reflect.TypeOf((*SlackConnectionConfig)(nil)).Elem(),
	"teams":         reflect.TypeOf((*MicrosoftTeamsConnectionConfig)(nil)).Elem(),
	"trello":        reflect.TypeOf((*TrelloConnectionConfig)(nil)).Elem(),
	"uptimerobot":   reflect.TypeOf((*UptimeRobotConnectionConfig)(nil)).Elem(),
	"urlscan":       reflect.TypeOf((*UrlscanConnectionConfig)(nil)).Elem(),
	"vault":         reflect.TypeOf((*VaultConnectionConfig)(nil)).Elem(),
	"virustotal":    reflect.TypeOf((*VirusTotalConnectionConfig)(nil)).Elem(),
	// "ipstack":       reflect.TypeOf((*IPStackConnectionConfig)(nil)).Elem(),
	// "zendesk":       reflect.TypeOf((*ZendeskCredential)(nil)).Elem(),
	// "okta":          reflect.TypeOf((*OktaCredential)(nil)).Elem(),
	// "jira":          reflect.TypeOf((*JiraCredential)(nil)).Elem(),
	// "opsgenie":      reflect.TypeOf((*OpsgenieCredential)(nil)).Elem(),
	// "azure":         reflect.TypeOf((*AzureCredential)(nil)).Elem(),
	// "gcp":           reflect.TypeOf((*GcpCredential)(nil)).Elem(),
	// "bitbucket":     reflect.TypeOf((*BitbucketCredential)(nil)).Elem(),
	// "datadog":       reflect.TypeOf((*DatadogCredential)(nil)).Elem(),
	// "freshdesk":     reflect.TypeOf((*FreshdeskCredential)(nil)).Elem(),
	// "guardrails":    reflect.TypeOf((*GuardrailsCredential)(nil)).Elem(),
	// "servicenow":    reflect.TypeOf((*ServiceNowCredential)(nil)).Elem(),
}

func InstantiateCredentialConfig(key string) (CredentialConfig, error) {
	// If the key has a slash, extract the last part of the key
	if strings.Contains(key, "/") {
		strParts := strings.Split(key, "/")
		key = strParts[len(strParts)-1]
	}

	t, exists := credentialConfigTypeRegistry[key]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid credential type " + key)
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
	GetCredential(string) Credential
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
