package connection

import (
	"context"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
)

var connectionTypRegistry = map[string]reflect.Type{
	(&AbuseIPDBConnection{}).GetConnectionType(): reflect.TypeOf(AbuseIPDBConnection{}),
	(&AlicloudConnection{}).GetConnectionType():  reflect.TypeOf(AlicloudConnection{}),
	(&AwsConnection{}).GetConnectionType():       reflect.TypeOf(AwsConnection{}),
	"azure":                                      reflect.TypeOf(AzureConnection{}),
	"bitbucket":                                  reflect.TypeOf(BitbucketConnection{}),
	"clickup":                                    reflect.TypeOf(ClickUpConnection{}),
	"datadog":                                    reflect.TypeOf(DatadogConnection{}),
	"discord":                                    reflect.TypeOf(DiscordConnection{}),
	"freshdesk":                                  reflect.TypeOf(FreshdeskConnection{}),
	"gcp":                                        reflect.TypeOf(GcpConnection{}),
	"github":                                     reflect.TypeOf(GithubConnection{}),
	"gitlab":                                     reflect.TypeOf(GitLabConnection{}),
	"ip2locationio":                              reflect.TypeOf(IP2LocationIOConnection{}),
	"ipstack":                                    reflect.TypeOf(IPstackConnection{}),
	"jira":                                       reflect.TypeOf(JiraConnection{}),
	"jumpcloud":                                  reflect.TypeOf(JumpCloudConnection{}),
	"mastodon":                                   reflect.TypeOf(MastodonConnection{}),
	"microsoft_teams":                            reflect.TypeOf(MicrosoftTeamsConnection{}),
	"okta":                                       reflect.TypeOf(OktaConnection{}),
	"openai":                                     reflect.TypeOf(OpenAIConnection{}),
	"opsgenie":                                   reflect.TypeOf(OpsgenieConnection{}),
	"pagerduty":                                  reflect.TypeOf(PagerDutyConnection{}),
	"sendgrid":                                   reflect.TypeOf(SendGridConnection{}),
	"servicenow":                                 reflect.TypeOf(ServiceNowConnection{}),
	"slack":                                      reflect.TypeOf(SlackConnection{}),
	"trello":                                     reflect.TypeOf(TrelloConnection{}),
	"turbot_guardrails":                          reflect.TypeOf(GuardrailsConnection{}),
	"turbot_pipes":                               reflect.TypeOf(PipesConnection{}),
	"uptime_robot":                               reflect.TypeOf(UptimeRobotConnection{}),
	"urlscan":                                    reflect.TypeOf(UrlscanConnection{}),
	"vault":                                      reflect.TypeOf(VaultConnection{}),
	"virus_total":                                reflect.TypeOf(VirusTotalConnection{}),
	"zendesk":                                    reflect.TypeOf(ZendeskConnection{}),
}

type PipelingConnection interface {
	modconfig.HclResource
	modconfig.ResourceWithMetadata

	SetHclResourceImpl(hclResourceImpl modconfig.HclResourceImpl)
	GetConnectionType() string
	GetUnqualifiedName() string

	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (PipelingConnection, error)
	GetTtl() int // in seconds

	Validate() hcl.Diagnostics
	getEnv() map[string]cty.Value

	Equals(PipelingConnection) bool
}

func ConnectionCtyType(connectionType string) cty.Type {
	goType := connectionTypRegistry[connectionType]
	if goType == nil {
		return cty.NilType
	}

	return cty.Capsule(connectionType, goType)
}

func NewConnection(block *hcl.Block) (PipelingConnection, error) {
	connectionType := block.Labels[0]
	connectionName := block.Labels[1]

	hclResourceImpl := modconfig.NewHclResourceImplNoMod(block, connectionType, connectionName)

	conn, err := instantiateConnection(connectionType, hclResourceImpl)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func instantiateConnection(key string, hclResourceImpl modconfig.HclResourceImpl) (PipelingConnection, error) {
	t, exists := connectionTypRegistry[key]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid connection type " + key)
	}
	credInterface := reflect.New(t).Interface()
	cred, ok := credInterface.(PipelingConnection)
	if !ok {
		return nil, perr.InternalWithMessage("Failed to create connection")
	}
	cred.SetHclResourceImpl(hclResourceImpl)

	return cred, nil
}

type ConnectionImpl struct {
	modconfig.HclResourceImpl
	modconfig.ResourceWithMetadataImpl

	// required to allow partial decoding
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`
}

func (c *ConnectionImpl) GetUnqualifiedName() string {
	return c.HclResourceImpl.UnqualifiedName
}

func (c *ConnectionImpl) SetHclResourceImpl(hclResourceImpl modconfig.HclResourceImpl) {
	c.HclResourceImpl = hclResourceImpl
}

func DefaultPipelingConnections() (map[string]PipelingConnection, error) {
	conns := make(map[string]PipelingConnection)

	for k := range connectionTypRegistry {
		hclResourceImpl := modconfig.HclResourceImpl{
			FullName:        k + ".default",
			ShortName:       "default",
			UnqualifiedName: k + ".default",
		}

		defaultCred, err := instantiateConnection(k, hclResourceImpl)
		if err != nil {
			return nil, err
		}

		conns[k+".default"] = defaultCred

		error_helpers.RegisterConnectionType(k)
	}

	return conns, nil
}
