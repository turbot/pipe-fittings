package test_init

import (
	"reflect"

	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
)

func SetAppSpecificConstants() {
	installDir, err := files.Tildefy("~/.flowpipe")
	if err != nil {
		panic(err)
	}

	app_specific.AppName = "flowpipe"
	// TODO unify version logic with steampipe
	//app_specific.AppVersion
	app_specific.AutoVariablesExtensions = []string{".auto.fpvars"}
	//app_specific.ClientConnectionAppNamePrefix
	//app_specific.ClientSystemConnectionAppNamePrefix
	app_specific.DefaultInstallDir = installDir
	app_specific.DefaultVarsFileName = "flowpipe.fpvars"
	//app_specific.DefaultDatabase
	//app_specific.EnvAppPrefix
	app_specific.EnvInputVarPrefix = "FP_VAR_"
	//app_specific.InstallDir
	app_specific.ConfigExtension = ".fpc"
	app_specific.ModDataExtensions = []string{".fp"}
	app_specific.VariablesExtensions = []string{".fpvars"}
	//app_specific.ServiceConnectionAppNamePrefix
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	app_specific.WorkspaceDataDir = ".flowpipe"
	app_specific_connection.ConnectionTypRegistry = map[string]reflect.Type{
		(&connection.AbuseIPDBConnection{}).GetConnectionType(): reflect.TypeOf(connection.AbuseIPDBConnection{}),
		(&connection.AlicloudConnection{}).GetConnectionType():  reflect.TypeOf(connection.AlicloudConnection{}),
		(&connection.AwsConnection{}).GetConnectionType():       reflect.TypeOf(connection.AwsConnection{}),
		"azure":             reflect.TypeOf(connection.AzureConnection{}),
		"bitbucket":         reflect.TypeOf(connection.BitbucketConnection{}),
		"clickup":           reflect.TypeOf(connection.ClickUpConnection{}),
		"datadog":           reflect.TypeOf(connection.DatadogConnection{}),
		"discord":           reflect.TypeOf(connection.DiscordConnection{}),
		"freshdesk":         reflect.TypeOf(connection.FreshdeskConnection{}),
		"gcp":               reflect.TypeOf(connection.GcpConnection{}),
		"github":            reflect.TypeOf(connection.GithubConnection{}),
		"gitlab":            reflect.TypeOf(connection.GitLabConnection{}),
		"ip2locationio":     reflect.TypeOf(connection.IP2LocationIOConnection{}),
		"ipstack":           reflect.TypeOf(connection.IPstackConnection{}),
		"jira":              reflect.TypeOf(connection.JiraConnection{}),
		"jumpcloud":         reflect.TypeOf(connection.JumpCloudConnection{}),
		"mastodon":          reflect.TypeOf(connection.MastodonConnection{}),
		"microsoft_teams":   reflect.TypeOf(connection.MicrosoftTeamsConnection{}),
		"okta":              reflect.TypeOf(connection.OktaConnection{}),
		"openai":            reflect.TypeOf(connection.OpenAIConnection{}),
		"opsgenie":          reflect.TypeOf(connection.OpsgenieConnection{}),
		"pagerduty":         reflect.TypeOf(connection.PagerDutyConnection{}),
		"sendgrid":          reflect.TypeOf(connection.SendGridConnection{}),
		"servicenow":        reflect.TypeOf(connection.ServiceNowConnection{}),
		"slack":             reflect.TypeOf(connection.SlackConnection{}),
		"trello":            reflect.TypeOf(connection.TrelloConnection{}),
		"turbot_guardrails": reflect.TypeOf(connection.GuardrailsConnection{}),
		"turbot_pipes":      reflect.TypeOf(connection.PipesConnection{}),
		"uptime_robot":      reflect.TypeOf(connection.UptimeRobotConnection{}),
		"urlscan":           reflect.TypeOf(connection.UrlscanConnection{}),
		"vault":             reflect.TypeOf(connection.VaultConnection{}),
		"virus_total":       reflect.TypeOf(connection.VirusTotalConnection{}),
		"zendesk":           reflect.TypeOf(connection.ZendeskConnection{}),
	}
}
