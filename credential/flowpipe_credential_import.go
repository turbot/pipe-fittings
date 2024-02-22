package credential

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"

	gokit "github.com/turbot/go-kit/helpers"
)

// The definition of a single Flowpipe CredentialImport
type CredentialImport struct {
	modconfig.HclResourceImpl
	modconfig.ResourceWithMetadataImpl

	FileName        string `json:"file_name"`
	StartLineNumber int    `json:"start_line_number"`
	EndLineNumber   int    `json:"end_line_number"`

	Source      *string  `json:"source" cty:"source" hcl:"source"`
	Connections []string `json:"connections" cty:"connections" hcl:"connections,optional"`
	Prefix      *string  `json:"prefix" cty:"prefix" hcl:"prefix,optional"`
}

func (c CredentialImport) Equals(other CredentialImport) bool {

	if c.FileName != other.FileName || c.StartLineNumber != other.StartLineNumber || c.EndLineNumber != other.EndLineNumber {
		return false
	}

	if !utils.StringPtrEqual(c.Source, other.Source) {
		return false
	}

	if !gokit.StringSliceEqualIgnoreOrder(c.Connections, other.Connections) {
		return false
	}

	if !utils.StringPtrEqual(c.Prefix, other.Prefix) {
		return false
	}

	return true
}

func (c *CredentialImport) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	c.FileName = fileName
	c.StartLineNumber = startLineNumber
	c.EndLineNumber = endLineNumber
}

func (c *CredentialImport) GetSource() *string {
	return c.Source
}

func (c *CredentialImport) GetPrefix() *string {
	return c.Prefix
}

func (c *CredentialImport) GetConnections() []string {
	return c.Connections
}

func NewCredentialImport(block *hcl.Block) *CredentialImport {

	credentialImportName := block.Labels[0]

	return &CredentialImport{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName:        credentialImportName,
			ShortName:       credentialImportName,
			UnqualifiedName: credentialImportName,
			DeclRange:       block.DefRange,
		},
		FileName:        block.DefRange.Filename,
		StartLineNumber: block.DefRange.Start.Line,
		EndLineNumber:   block.DefRange.End.Line,
	}
}

func ResolveConfigStruct(connectionType string) any {
	typeRegistry := map[string]reflect.Type{
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

	t, exists := typeRegistry[connectionType]
	if !exists {
		return nil
	}

	// Use reflect.New to create a pointer to the type, rather than a value.
	return reflect.New(t).Interface()
}
