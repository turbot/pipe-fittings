package pipeline_test

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/tests/test_init"
	"github.com/zclconf/go-cty/cty"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeModTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeModTestSuite) SetupSuite() {

	err := os.Setenv("RUN_MODE", "TEST_ES")
	if err != nil {
		panic(err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// clear the output dir before each test
	outputPath := path.Join(cwd, "output")

	// Check if the directory exists
	_, err = os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
		if err != nil {
			panic(err)
		}

	}

	pipelineDirPath := path.Join(cwd, "pipelines")

	viper.GetViper().Set("pipeline.dir", pipelineDirPath)
	viper.GetViper().Set("output.dir", outputPath)
	viper.GetViper().Set("log.dir", outputPath)

	// Create a single, global context for the application
	ctx := context.Background()

	suite.ctx = ctx

	// set app specific constants
	test_init.SetAppSpecificConstants()

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeModTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

func (suite *FlowpipeModTestSuite) TestGoodMod() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.Load(suite.ctx, "./good_mod", workspace.WithCredentials(map[string]credential.Credential{}))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	assert.Equal("0.1.0", mod.Require.Flowpipe.MinVersionString)
	assert.Equal("day", mod.Tags["green"])

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines

	jsonForPipeline := pipelines["test_mod.pipeline.json_for"]
	if jsonForPipeline == nil {
		assert.Fail("json_for pipeline not found")
		return
	}

	// check if all steps are there
	assert.Equal(2, len(jsonForPipeline.Steps), "wrong number of steps")
	assert.Equal(jsonForPipeline.Steps[0].GetName(), "json", "wrong step name")
	assert.Equal(jsonForPipeline.Steps[0].GetType(), "transform", "wrong step type")
	assert.Equal(jsonForPipeline.Steps[1].GetName(), "json_for", "wrong step name")
	assert.Equal(jsonForPipeline.Steps[1].GetType(), "transform", "wrong step type")

	// check if all triggers are there
	triggers := mod.ResourceMaps.Triggers
	assert.Equal(1, len(triggers), "wrong number of triggers")
	assert.Equal("test_mod.trigger.schedule.my_hourly_trigger", triggers["test_mod.trigger.schedule.my_hourly_trigger"].FullName, "wrong trigger name")

	inlineDocPipeline := pipelines["test_mod.pipeline.inline_documentation"]
	if inlineDocPipeline == nil {
		assert.Fail("inline_documentation pipeline not found")
		return
	}

	assert.Equal("inline doc", *inlineDocPipeline.Description)
	assert.Equal("inline pipeline documentation", *inlineDocPipeline.Documentation)

	docFromFile := pipelines["test_mod.pipeline.doc_from_file"]
	if docFromFile == nil {
		assert.Fail("doc_from_file pipeline not found")
		return
	}

	assert.Contains(*docFromFile.Documentation, "the quick brown fox jumps over the lazy dog")

	pipeline := pipelines["test_mod.pipeline.step_with_if_and_depends"]
	if pipeline == nil {
		assert.Fail("step_with_if_and_depends pipeline not found")
		return
	}

	// get the last step
	step := pipeline.Steps[len(pipeline.Steps)-1]
	assert.Equal("three", step.GetName())

	dependsOn := step.GetDependsOn()
	assert.Equal(2, len(dependsOn))

	slices.Sort[[]string, string](dependsOn)

	assert.Equal("transform.one", dependsOn[0])
	assert.Equal("transform.two", dependsOn[1])
}

func (suite *FlowpipeModTestSuite) TestModReferences() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_references", workspace.WithCredentials(map[string]credential.Credential{}))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")
	assert.Equal(2, len(pipelines), "wrong number of pipelines")
	assert.NotNil(pipelines["pipeline_with_references.pipeline.foo"])
	assert.NotNil(pipelines["pipeline_with_references.pipeline.foo_two"])
}

func (suite *FlowpipeModTestSuite) TestModWithCreds() {
	assert := assert.New(suite.T())

	credentials := map[string]credential.Credential{
		"aws.default": &credential.AwsCredential{
			CredentialImpl: credential.CredentialImpl{
				HclResourceImpl: modconfig.HclResourceImpl{
					FullName:        "aws.default",
					ShortName:       "default",
					UnqualifiedName: "aws.default",
				},
				Type: "aws",
			},
		},
	}

	os.Setenv("ACCESS_KEY", "foobarbaz")
	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_creds", workspace.WithCredentials(credentials))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")

	pipeline := pipelines["mod_with_creds.pipeline.with_creds"]
	assert.Equal("aws.default", pipeline.Steps[0].GetCredentialDependsOn()[0], "there's only 1 step in this pipeline and it should have a credential dependency")

	stepInputs, err := pipeline.Steps[1].GetInputs(nil)
	assert.Nil(err)

	assert.Equal("foobarbaz", stepInputs["value"], "token should be set to foobarbaz")
	os.Unsetenv("ACCESS_KEY")
}

func (suite *FlowpipeModTestSuite) TestModWithCredsWithContextFunction() {
	assert := assert.New(suite.T())

	os.Setenv("TEST_SLACK_TOKEN", "abcdefghi")

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_with_creds_using_context_function"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_creds_using_context_function", workspace.WithCredentials(flowpipeConfig.Credentials))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	credentials := w.Credentials
	slackCreds := credentials["slack.slack_creds"]
	slackCredsCty, e := slackCreds.CtyValue()
	assert.Nil(e)

	credsMap := slackCredsCty.AsValueMap()
	tokenVal := credsMap["token"].AsString()
	assert.Equal("abcdefghi", tokenVal)

	os.Unsetenv("TEST_SLACK_TOKEN")
}

func (suite *FlowpipeModTestSuite) TestFlowpipeIntegrationSerialiseDeserialise() {
	assert := assert.New(suite.T())

	flowpipeConfig, ew := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_dir_more_integrations", "./mod_with_integration"})
	if ew.Error != nil {
		assert.FailNow(ew.Error.Error())
		return
	}

	if flowpipeConfig == nil {
		assert.Fail("flowpipeConfig is nil")
		return
	}

	notifier := flowpipeConfig.Notifiers["devs"]

	// marshall to JSON test
	jsonBytes, err := json.Marshal(notifier)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Nil(err)
	assert.NotNil(jsonBytes)

	// unmarshall from JSON test
	var notifier2 modconfig.NotifierImpl
	err = json.Unmarshal(jsonBytes, &notifier2)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Equal(2, len(notifier2.GetNotifies()))
	assert.Equal("#devs", *notifier2.GetNotifies()[0].Channel)
	assert.Equal("xoxp-111111", *notifier2.GetNotifies()[0].Integration.(*modconfig.SlackIntegration).Token)
}

func (suite *FlowpipeModTestSuite) TestFlowpipeConfigIntegrationEmail() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_dir_more_integrations", "./mod_with_integration"})
	if err.Error != nil {
		assert.FailNow(err.Error.Error())
		return
	}

	if flowpipeConfig == nil {
		assert.Fail("flowpipeConfig is nil")
		return
	}

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_integration", workspace.WithCredentials(flowpipeConfig.Credentials), workspace.WithIntegrations(flowpipeConfig.Integrations), workspace.WithNotifiers(flowpipeConfig.Notifiers))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)
	assert.Equal(5, len(w.Integrations))

	pipelines := w.Mod.ResourceMaps.Pipelines
	pipeline := pipelines["mod_with_integration.pipeline.approval_with_notifies"]
	if pipeline == nil {
		assert.Fail("pipeline approval_with_notifies not found")
		return
	}

	step, ok := pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Step is not an input step")
		return
	}

	notifies := step.Notifier.GetNotifies()
	assert.Len(notifies, 1)
	notify := notifies[0]
	assert.NotNil(notify)
	toList := notify.To
	assert.Equal(2, len(toList))
	assert.Equal("foo@bar.com", toList[0])
	assert.Equal("baz@bar.com", toList[1])

	integrations := notify.Integration
	assert.NotNil(integrations)
	assert.Equal("user@test.tld", *integrations.(*modconfig.EmailIntegration).DefaultRecipient)
}

func (suite *FlowpipeModTestSuite) TestFlowpipeConfigWithCredImport() {
	assert := assert.New(suite.T())

	// Load the config from 2 different directories to test that we can load from multiple directories where the integration is defined after
	// we load the notifiers.
	//
	// ensure that "config_dir" is loaded first, that's where the notifier is.
	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_dir_with_cred_import", "./empty_mod"})
	if err.Error != nil {
		assert.FailNow(err.Error.Error())
		return
	}

	if flowpipeConfig == nil {
		assert.Fail("flowpipeConfig is nil")
		return
	}

	// AbuseIPDB
	assert.Equal("steampipe_abuseipdb", flowpipeConfig.CredentialImports["steampipe_abuseipdb"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_abuseipdb"].Prefix)
	assert.Equal("abuseipdb.sp1_abuseipdb_1", flowpipeConfig.Credentials["abuseipdb.sp1_abuseipdb_1"].GetHclResourceImpl().FullName)
	assert.Equal("abuseipdb.sp1_abuseipdb_2", flowpipeConfig.Credentials["abuseipdb.sp1_abuseipdb_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["abuseipdb.sp1_abuseipdb_1"].(*credential.AbuseIPDBCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["abuseipdb.sp1_abuseipdb_2"].(*credential.AbuseIPDBCredential).APIKey)

	// AWS
	assert.Equal("steampipe", flowpipeConfig.CredentialImports["steampipe"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe"].Prefix)
	assert.Equal("aws.sp1_aws", flowpipeConfig.Credentials["aws.sp1_aws"].GetHclResourceImpl().FullName)
	assert.Equal("aws.sp1_aws_keys1", flowpipeConfig.Credentials["aws.sp1_aws_keys1"].GetHclResourceImpl().FullName)
	assert.Equal("abc", *flowpipeConfig.Credentials["aws.sp1_aws_keys1"].(*credential.AwsCredential).AccessKey)
	assert.Equal("123", *flowpipeConfig.Credentials["aws.sp1_aws_keys1"].(*credential.AwsCredential).SecretKey)

	// ClickUp
	assert.Equal("steampipe_clickup", flowpipeConfig.CredentialImports["steampipe_clickup"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_clickup"].Prefix)
	assert.Equal("clickup.sp1_clickup_1", flowpipeConfig.Credentials["clickup.sp1_clickup_1"].GetHclResourceImpl().FullName)
	assert.Equal("clickup.sp1_clickup_2", flowpipeConfig.Credentials["clickup.sp1_clickup_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["clickup.sp1_clickup_1"].(*credential.ClickUpCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["clickup.sp1_clickup_2"].(*credential.ClickUpCredential).Token)

	// Discord
	assert.Equal("steampipe_discord", flowpipeConfig.CredentialImports["steampipe_discord"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_discord"].Prefix)
	assert.Equal("discord.sp1_discord_1", flowpipeConfig.Credentials["discord.sp1_discord_1"].GetHclResourceImpl().FullName)
	assert.Equal("discord.sp1_discord_2", flowpipeConfig.Credentials["discord.sp1_discord_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["discord.sp1_discord_1"].(*credential.DiscordCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["discord.sp1_discord_2"].(*credential.DiscordCredential).Token)

	// Freshdesk
	assert.Equal("steampipe_freshdesk", flowpipeConfig.CredentialImports["steampipe_freshdesk"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_freshdesk"].Prefix)
	assert.Equal("freshdesk.sp1_freshdesk_1", flowpipeConfig.Credentials["freshdesk.sp1_freshdesk_1"].GetHclResourceImpl().FullName)
	assert.Equal("freshdesk.sp1_freshdesk_2", flowpipeConfig.Credentials["freshdesk.sp1_freshdesk_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["freshdesk.sp1_freshdesk_1"].(*credential.FreshdeskCredential).APIKey)
	assert.Equal("test", *flowpipeConfig.Credentials["freshdesk.sp1_freshdesk_1"].(*credential.FreshdeskCredential).Subdomain)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["freshdesk.sp1_freshdesk_2"].(*credential.FreshdeskCredential).APIKey)
	assert.Equal("test", *flowpipeConfig.Credentials["freshdesk.sp1_freshdesk_2"].(*credential.FreshdeskCredential).Subdomain)

	// Github
	assert.Equal("steampipe_github", flowpipeConfig.CredentialImports["steampipe_github"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_github"].Prefix)
	assert.Equal("github.sp1_github_1", flowpipeConfig.Credentials["github.sp1_github_1"].GetHclResourceImpl().FullName)
	assert.Equal("github.sp1_github_2", flowpipeConfig.Credentials["github.sp1_github_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["github.sp1_github_1"].(*credential.GithubCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["github.sp1_github_2"].(*credential.GithubCredential).Token)

	// Gitlab
	assert.Equal("steampipe_gitlab", flowpipeConfig.CredentialImports["steampipe_gitlab"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_gitlab"].Prefix)
	assert.Equal("gitlab.sp1_gitlab_1", flowpipeConfig.Credentials["gitlab.sp1_gitlab_1"].GetHclResourceImpl().FullName)
	assert.Equal("gitlab.sp1_gitlab_2", flowpipeConfig.Credentials["gitlab.sp1_gitlab_2"].GetHclResourceImpl().FullName)
	assert.Equal("f7Ea3C3ojOY0GLzmhS5kE", *flowpipeConfig.Credentials["gitlab.sp1_gitlab_1"].(*credential.GitLabCredential).Token)
	assert.Equal("f7Ea3C3ojOY0GLzmhS5kE", *flowpipeConfig.Credentials["gitlab.sp1_gitlab_2"].(*credential.GitLabCredential).Token)

	// IP2LocationIO
	assert.Equal("steampipe_ip2locationio", flowpipeConfig.CredentialImports["steampipe_ip2locationio"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_ip2locationio"].Prefix)
	assert.Equal("ip2locationio.sp1_ip2locationio_1", flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_1"].GetHclResourceImpl().FullName)
	assert.Equal("ip2locationio.sp1_ip2locationio_2", flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_1"].(*credential.IP2LocationIOCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_2"].(*credential.IP2LocationIOCredential).APIKey)

	// JumpCloud
	assert.Equal("steampipe_jumpcloud", flowpipeConfig.CredentialImports["steampipe_jumpcloud"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_jumpcloud"].Prefix)
	assert.Equal("jumpcloud.sp1_jumpcloud_1", flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_1"].GetHclResourceImpl().FullName)
	assert.Equal("jumpcloud.sp1_jumpcloud_2", flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_1"].(*credential.JumpCloudCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_2"].(*credential.JumpCloudCredential).APIKey)

	// Microsoft Teams
	assert.Equal("steampipe_teams", flowpipeConfig.CredentialImports["steampipe_teams"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_teams"].Prefix)
	assert.Equal("teams.sp1_teams_1", flowpipeConfig.Credentials["teams.sp1_teams_1"].GetHclResourceImpl().FullName)
	assert.Equal("teams.sp1_teams_2", flowpipeConfig.Credentials["teams.sp1_teams_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["teams.sp1_teams_1"].(*credential.MicrosoftTeamsCredential).AccessToken)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["teams.sp1_teams_2"].(*credential.MicrosoftTeamsCredential).AccessToken)

	// OpenAI
	assert.Equal("steampipe_openai", flowpipeConfig.CredentialImports["steampipe_openai"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_openai"].Prefix)
	assert.Equal("openai.sp1_openai_1", flowpipeConfig.Credentials["openai.sp1_openai_1"].GetHclResourceImpl().FullName)
	assert.Equal("openai.sp1_openai_2", flowpipeConfig.Credentials["openai.sp1_openai_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["openai.sp1_openai_1"].(*credential.OpenAICredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["openai.sp1_openai_2"].(*credential.OpenAICredential).APIKey)

	// PagerDuty
	assert.Equal("steampipe_pagerduty", flowpipeConfig.CredentialImports["steampipe_pagerduty"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_pagerduty"].Prefix)
	assert.Equal("pagerduty.sp1_pagerduty_1", flowpipeConfig.Credentials["pagerduty.sp1_pagerduty_1"].GetHclResourceImpl().FullName)
	assert.Equal("pagerduty.sp1_pagerduty_2", flowpipeConfig.Credentials["pagerduty.sp1_pagerduty_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["pagerduty.sp1_pagerduty_1"].(*credential.PagerDutyCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["pagerduty.sp1_pagerduty_2"].(*credential.PagerDutyCredential).Token)

	// Pipes
	assert.Equal("steampipe_pipes", flowpipeConfig.CredentialImports["steampipe_pipes"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_pipes"].Prefix)
	assert.Equal("pipes.sp1_pipes_1", flowpipeConfig.Credentials["pipes.sp1_pipes_1"].GetHclResourceImpl().FullName)
	assert.Equal("pipes.sp1_pipes_2", flowpipeConfig.Credentials["pipes.sp1_pipes_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["pipes.sp1_pipes_1"].(*credential.PipesCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["pipes.sp1_pipes_2"].(*credential.PipesCredential).Token)

	// SendGrid
	assert.Equal("steampipe_sendgrid", flowpipeConfig.CredentialImports["steampipe_sendgrid"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_sendgrid"].Prefix)
	assert.Equal("sendgrid.sp1_sendgrid_1", flowpipeConfig.Credentials["sendgrid.sp1_sendgrid_1"].GetHclResourceImpl().FullName)
	assert.Equal("sendgrid.sp1_sendgrid_2", flowpipeConfig.Credentials["sendgrid.sp1_sendgrid_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["sendgrid.sp1_sendgrid_1"].(*credential.SendGridCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["sendgrid.sp1_sendgrid_2"].(*credential.SendGridCredential).APIKey)

	// ServiceNow
	assert.Equal("steampipe_servicenow", flowpipeConfig.CredentialImports["steampipe_servicenow"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_servicenow"].Prefix)
	assert.Equal("servicenow.sp1_servicenow_1", flowpipeConfig.Credentials["servicenow.sp1_servicenow_1"].GetHclResourceImpl().FullName)
	assert.Equal("servicenow.sp1_servicenow_2", flowpipeConfig.Credentials["servicenow.sp1_servicenow_2"].GetHclResourceImpl().FullName)
	assert.Equal("https://test.service-now.com", *flowpipeConfig.Credentials["servicenow.sp1_servicenow_1"].(*credential.ServiceNowCredential).InstanceURL)
	assert.Equal("flowpipe", *flowpipeConfig.Credentials["servicenow.sp1_servicenow_1"].(*credential.ServiceNowCredential).Username)
	assert.Equal("somepassword", *flowpipeConfig.Credentials["servicenow.sp1_servicenow_1"].(*credential.ServiceNowCredential).Password)
	assert.Equal("https://test1.service-now.com", *flowpipeConfig.Credentials["servicenow.sp1_servicenow_2"].(*credential.ServiceNowCredential).InstanceURL)
	assert.Equal("flowpipe", *flowpipeConfig.Credentials["servicenow.sp1_servicenow_2"].(*credential.ServiceNowCredential).Username)
	assert.Equal("somepassword1", *flowpipeConfig.Credentials["servicenow.sp1_servicenow_2"].(*credential.ServiceNowCredential).Password)

	// Slack
	assert.Equal("steampipe_slack", flowpipeConfig.CredentialImports["steampipe_slack"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_slack"].Prefix)
	assert.Equal("slack.sp1_slack_l1", flowpipeConfig.Credentials["slack.sp1_slack_l1"].GetHclResourceImpl().FullName)
	assert.Equal("slack.sp1_slack_l2", flowpipeConfig.Credentials["slack.sp1_slack_l2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["slack.sp1_slack_l1"].(*credential.SlackCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["slack.sp1_slack_l2"].(*credential.SlackCredential).Token)

	// Trello
	assert.Equal("steampipe_trello", flowpipeConfig.CredentialImports["steampipe_trello"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_trello"].Prefix)
	assert.Equal("trello.sp1_trello_1", flowpipeConfig.Credentials["trello.sp1_trello_1"].GetHclResourceImpl().FullName)
	assert.Equal("trello.sp1_trello_2", flowpipeConfig.Credentials["trello.sp1_trello_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["trello.sp1_trello_1"].(*credential.TrelloCredential).APIKey)
	assert.Equal("testtoken", *flowpipeConfig.Credentials["trello.sp1_trello_1"].(*credential.TrelloCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["trello.sp1_trello_2"].(*credential.TrelloCredential).APIKey)
	assert.Equal("testtoken", *flowpipeConfig.Credentials["trello.sp1_trello_2"].(*credential.TrelloCredential).Token)

	// UptimeRobot
	assert.Equal("steampipe_uptimerobot", flowpipeConfig.CredentialImports["steampipe_uptimerobot"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_uptimerobot"].Prefix)
	assert.Equal("uptimerobot.sp1_uptimerobot_1", flowpipeConfig.Credentials["uptimerobot.sp1_uptimerobot_1"].GetHclResourceImpl().FullName)
	assert.Equal("uptimerobot.sp1_uptimerobot_2", flowpipeConfig.Credentials["uptimerobot.sp1_uptimerobot_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["uptimerobot.sp1_uptimerobot_1"].(*credential.UptimeRobotCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["uptimerobot.sp1_uptimerobot_2"].(*credential.UptimeRobotCredential).APIKey)

	// Urlscan
	assert.Equal("steampipe_urlscan", flowpipeConfig.CredentialImports["steampipe_urlscan"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_urlscan"].Prefix)
	assert.Equal("urlscan.sp1_urlscan_1", flowpipeConfig.Credentials["urlscan.sp1_urlscan_1"].GetHclResourceImpl().FullName)
	assert.Equal("urlscan.sp1_urlscan_2", flowpipeConfig.Credentials["urlscan.sp1_urlscan_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["urlscan.sp1_urlscan_1"].(*credential.UrlscanCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["urlscan.sp1_urlscan_2"].(*credential.UrlscanCredential).APIKey)

	// Vault
	assert.Equal("steampipe_vault", flowpipeConfig.CredentialImports["steampipe_vault"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_vault"].Prefix)
	assert.Equal("vault.sp1_vault_1", flowpipeConfig.Credentials["vault.sp1_vault_1"].GetHclResourceImpl().FullName)
	assert.Equal("vault.sp1_vault_2", flowpipeConfig.Credentials["vault.sp1_vault_2"].GetHclResourceImpl().FullName)
	assert.Equal("https://vault.mycorp.com/", *flowpipeConfig.Credentials["vault.sp1_vault_1"].(*credential.VaultCredential).Address)
	assert.Equal("sometoken", *flowpipeConfig.Credentials["vault.sp1_vault_1"].(*credential.VaultCredential).Token)
	assert.Equal("https://vault.mycorp.com/", *flowpipeConfig.Credentials["vault.sp1_vault_2"].(*credential.VaultCredential).Address)
	assert.Nil(flowpipeConfig.Credentials["vault.sp1_vault_2"].(*credential.VaultCredential).Token)

	// VirusTotal
	assert.Equal("steampipe_virustotal", flowpipeConfig.CredentialImports["steampipe_virustotal"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_virustotal"].Prefix)
	assert.Equal("virustotal.sp1_virustotal_1", flowpipeConfig.Credentials["virustotal.sp1_virustotal_1"].GetHclResourceImpl().FullName)
	assert.Equal("virustotal.sp1_virustotal_2", flowpipeConfig.Credentials["virustotal.sp1_virustotal_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["virustotal.sp1_virustotal_1"].(*credential.VirusTotalCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["virustotal.sp1_virustotal_2"].(*credential.VirusTotalCredential).APIKey)
}

func (suite *FlowpipeModTestSuite) TestFlowpipeConfigIntegration() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_dir", "./mod_with_integration"})
	if err.Error != nil {
		assert.FailNow(err.Error.Error())
		return
	}

	if flowpipeConfig == nil {
		assert.Fail("flowpipeConfig is nil")
		return
	}

	assert.Equal(2, len(flowpipeConfig.Integrations))
	assert.Equal("slack.my_slack_app", flowpipeConfig.Integrations["slack.my_slack_app"].GetHclResourceImpl().FullName)

	// ensure that the default integration exist
	assert.Equal("webform.default", flowpipeConfig.Integrations["webform.default"].GetHclResourceImpl().FullName)

	assert.Equal(3, len(flowpipeConfig.Notifiers))

	// ensure that default notifier exist
	assert.Equal("default", flowpipeConfig.Notifiers["default"].GetHclResourceImpl().FullName)
	assert.Equal(1, len(flowpipeConfig.Notifiers["default"].GetNotifies()))

	// TODO: test this when we have webform up and running
	//assert.Equal("Q#$$#@#$$#W", flowpipeConfig.Notifiers["default"].GetNotifies()[0].Integration.AsValueMap()["name"].AsString())

	assert.Equal("admins", flowpipeConfig.Notifiers["admins"].GetHclResourceImpl().FullName)
	// Check the notify -> integration link
	assert.Equal(1, len(flowpipeConfig.Notifiers["admins"].GetNotifies()))

	assert.Equal("Q#$$#@#$$#W", *flowpipeConfig.Notifiers["admins"].GetNotifies()[0].Integration.(*modconfig.SlackIntegration).SigningSecret)
	assert.Equal("xoxp-111111", *flowpipeConfig.Notifiers["admins"].GetNotifies()[0].Integration.(*modconfig.SlackIntegration).Token)

	devsNotifier := flowpipeConfig.Notifiers["devs"]
	assert.Equal("devs", devsNotifier.GetHclResourceImpl().FullName)
	assert.Equal(2, len(devsNotifier.GetNotifies()))

	dvCtyVal, err2 := devsNotifier.CtyValue()
	if err2 != nil {
		assert.Fail(err2.Error())
		return
	}

	if dvCtyVal == cty.NilVal {
		assert.Fail("cty value is nil")
		return
	}

	devsNotifierMap := dvCtyVal.AsValueMap()
	devsNotifiesSlice := devsNotifierMap["notifies"].AsValueSlice()
	assert.Equal(2, len(devsNotifiesSlice))
	assert.Equal("#devs", devsNotifiesSlice[0].AsValueMap()["channel"].AsString())

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_integration", workspace.WithCredentials(flowpipeConfig.Credentials), workspace.WithIntegrations(flowpipeConfig.Integrations), workspace.WithNotifiers(flowpipeConfig.Notifiers))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)
	assert.Equal(2, len(w.Integrations))
	assert.NotNil(w.Integrations["slack.my_slack_app"])
	if i, ok := w.Integrations["slack.my_slack_app"].(*modconfig.SlackIntegration); !ok {
		assert.Fail("integration failed to parse to SlackIntegration")
	} else {
		assert.Equal("slack.my_slack_app", i.FullName)
		assert.Equal("#infosec", *i.Channel)
	}

	pipelines := w.Mod.ResourceMaps.Pipelines
	pipeline := pipelines["mod_with_integration.pipeline.approval_with_notifies"]
	if pipeline == nil {
		assert.Fail("pipeline approval_with_notifies not found")
		return
	}

	step, ok := pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Step is not an input step")
		return
	}
	assert.Equal("Do you want to approve?", *step.Prompt)

	// This notifier CtyValue function
	ctyVal, err2 := step.Notifier.CtyValue()
	if err2 != nil {
		assert.Fail(err2.Error())
		return
	}

	notifierMap := ctyVal.AsValueMap()
	notifiesSlice := notifierMap["notifies"].AsValueSlice()
	assert.Equal(1, len(notifiesSlice))

	notifies := step.Notifier.GetNotifies()
	assert.Len(notifies, 1)
	notify := notifies[0]
	assert.NotNil(notify)

	integration := notify.Integration
	assert.NotNil(integration)
	assert.Equal("Q#$$#@#$$#W", *integration.(*modconfig.SlackIntegration).SigningSecret)

	step, ok = pipeline.Steps[1].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Step is not an input step")
		return
	}

	assert.Equal("Do you want to approve (2)?", *step.Prompt)
	notifies = step.Notifier.GetNotifies()

	assert.Len(notifies, 1)
	notify = notifies[0]
	assert.NotNil(notify)

	integration = notify.Integration
	assert.NotNil(integration)
	assert.Equal("Q#$$#@#$$#W", *integration.(*modconfig.SlackIntegration).SigningSecret)

	pipeline = pipelines["mod_with_integration.pipeline.approval_with_notifies_dynamic"]
	if pipeline == nil {
		assert.Fail("pipeline approval_with_notifies_dynamic not found")
		return
	}

	step, ok = pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Step is not an input step")
		return
	}

	assert.NotNil(step.UnresolvedAttributes["notifier"])
}

func (suite *FlowpipeModTestSuite) TestModWithCredsNoEnvVarSet() {
	assert := assert.New(suite.T())

	credentials := map[string]credential.Credential{
		"aws.default": &credential.AwsCredential{
			CredentialImpl: credential.CredentialImpl{
				HclResourceImpl: modconfig.HclResourceImpl{
					FullName:        "aws.default",
					ShortName:       "default",
					UnqualifiedName: "aws.default",
				},
				Type: "aws",
			},
		},
	}

	// This is the same test with TestModWithCreds but with no ACCESS_KEY env var set, the value for the second step should be nil
	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_creds", workspace.WithCredentials(credentials))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")

	pipeline := pipelines["mod_with_creds.pipeline.with_creds"]
	assert.Equal("aws.default", pipeline.Steps[0].GetCredentialDependsOn()[0], "there's only 1 step in this pipeline and it should have a credential dependency")

	stepInputs, err := pipeline.Steps[1].GetInputs(nil)
	assert.Nil(err)
	assert.Equal("", stepInputs["value"])
}

func (suite *FlowpipeModTestSuite) TestModDynamicCreds() {
	assert := assert.New(suite.T())

	credentials := map[string]credential.Credential{
		"aws.aws_static": &credential.AwsCredential{
			CredentialImpl: credential.CredentialImpl{
				HclResourceImpl: modconfig.HclResourceImpl{
					FullName:        "aws.static",
					ShortName:       "static",
					UnqualifiedName: "aws.static",
				},
				Type: "aws",
			},
		},
	}

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_dynamic_creds", workspace.WithCredentials(credentials))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")

	pipeline := pipelines["mod_with_dynamic_creds.pipeline.cred_aws"]

	assert.Equal("aws.<dynamic>", pipeline.Steps[0].GetCredentialDependsOn()[0], "there's only 1 step in this pipeline and it should have a credential dependency")
}

func (suite *FlowpipeModTestSuite) TestModWithCredsResolved() {
	assert := assert.New(suite.T())

	credentials := map[string]credential.Credential{
		"slack.slack_static": &credential.SlackCredential{
			CredentialImpl: credential.CredentialImpl{
				HclResourceImpl: modconfig.HclResourceImpl{
					FullName:        "slack.slack_static",
					ShortName:       "slack_static",
					UnqualifiedName: "slack.slack_static",
				},
				Type: "slack",
			},
			Token: types.String("sfhshfhslfh"),
		},
	}

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_creds_resolved", workspace.WithCredentials(credentials))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")

	pipeline := pipelines["mod_with_creds_resolved.pipeline.staic_creds_test"]
	assert.Equal("slack.slack_static", pipeline.Steps[0].GetCredentialDependsOn()[0], "there's only 1 step in this pipeline and it should have a credential dependency")

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"slack": cty.ObjectVal(map[string]cty.Value{
			"slack_static": cty.ObjectVal(map[string]cty.Value{
				"token": cty.StringVal("sfhshfhslfh"),
			}),
		}),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["credential"] = paramVal

	stepInputs, err := pipeline.Steps[0].GetInputs(evalContext)
	assert.Nil(err)

	assert.Equal("sfhshfhslfh", stepInputs["value"], "token should be set to sfhshfhslfh")
}

func (suite *FlowpipeModTestSuite) TestStepOutputParsing() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_step_output", workspace.WithCredentials(map[string]credential.Credential{}))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")
	assert.Equal(1, len(pipelines), "wrong number of pipelines")

	assert.Equal(2, len(pipelines["test_mod.pipeline.with_step_output"].Steps), "wrong number of steps")
	assert.False(pipelines["test_mod.pipeline.with_step_output"].Steps[0].IsResolved())
	assert.False(pipelines["test_mod.pipeline.with_step_output"].Steps[1].IsResolved())

}

func (suite *FlowpipeModTestSuite) TestModDependencies() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_dep_one", workspace.WithCredentials(map[string]credential.Credential{}))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipelines := mod.ResourceMaps.Pipelines

	assert.NotNil(mod, "mod is nil")
	jsonForPipeline := pipelines["mod_parent.pipeline.json"]
	if jsonForPipeline == nil {
		assert.Fail("json pipeline not found")
		return
	}

	fooPipeline := pipelines["mod_parent.pipeline.foo"]
	if fooPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	fooTwoPipeline := pipelines["mod_parent.pipeline.foo_two"]
	if fooTwoPipeline == nil {
		assert.Fail("foo_two pipeline not found")
		return
	}

	referToChildPipeline := pipelines["mod_parent.pipeline.refer_to_child"]
	if referToChildPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	referToChildBPipeline := pipelines["mod_parent.pipeline.refer_to_child_b"]
	if referToChildBPipeline == nil {
		assert.Fail("refer_to_child_b pipeline not found")
		return
	}

	childModA := mod.ResourceMaps.Mods["mod_child_a@v1.0.0"]
	assert.NotNil(childModA)

	thisPipelineIsInTheChildPipelineModA := childModA.ResourceMaps.Pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child"]
	assert.NotNil(thisPipelineIsInTheChildPipelineModA)

	// check for the triggers
	triggers := mod.ResourceMaps.Triggers
	myHourlyTrigger := triggers["mod_parent.trigger.schedule.my_hourly_trigger"]
	if myHourlyTrigger == nil {
		assert.Fail("my_hourly_trigger not found")
		return
	}

}

func (suite *FlowpipeModTestSuite) TestModDependenciesSimple() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_dep_simple", workspace.WithCredentials(map[string]credential.Credential{}))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipelines := mod.ResourceMaps.Pipelines
	jsonForPipeline := pipelines["mod_parent.pipeline.json"]
	if jsonForPipeline == nil {
		assert.Fail("json pipeline not found")
		return
	}

	fooPipeline := pipelines["mod_parent.pipeline.foo"]
	if fooPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	assert.Equal(2, len(fooPipeline.Steps), "wrong number of steps")
	assert.Equal("baz", fooPipeline.Steps[0].GetName())
	assert.Equal("bar", fooPipeline.Steps[1].GetName())

	referToChildPipeline := pipelines["mod_parent.pipeline.refer_to_child"]
	if referToChildPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	childPipeline := pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child"]
	if childPipeline == nil {
		assert.Fail("this_pipeline_is_in_the_child pipeline not found")
		return
	}

	childPipelineWithVar := pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child_using_variable"]
	if childPipelineWithVar == nil {
		assert.Fail("this_pipeline_is_in_the_child pipeline not found")
		return
	}

	assert.Equal("foo: this is the value of var_one", childPipelineWithVar.Steps[0].(*modconfig.PipelineStepTransform).Value)

	childPipelineWithVarPassedFromParent := pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child_using_variable_passed_from_parent"]
	if childPipelineWithVarPassedFromParent == nil {
		assert.Fail("this_pipeline_is_in_the_child pipeline not found")
		return
	}

	assert.Equal("foo: var_two from parent .pvars file", childPipelineWithVarPassedFromParent.Steps[0].(*modconfig.PipelineStepTransform).Value)
}

func (suite *FlowpipeModTestSuite) TestModVariable() {
	assert := assert.New(suite.T())

	os.Setenv("FP_VAR_var_six", "set from env var")

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_variable", workspace.WithCredentials(map[string]credential.Credential{}))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipelines := mod.ResourceMaps.Pipelines
	pipelineOne := pipelines["test_mod.pipeline.one"]
	if pipelineOne == nil {
		assert.Fail("pipeline one not found")
		return
	}

	assert.Equal("prefix text here and this is the value of var_one and suffix", pipelineOne.Steps[0].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("prefix text here and value from var file and suffix", pipelineOne.Steps[1].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("prefix text here and var_three from var file and suffix", pipelineOne.Steps[2].(*modconfig.PipelineStepTransform).Value)

	assert.True(pipelineOne.Steps[0].IsResolved())
	assert.True(pipelineOne.Steps[1].IsResolved())
	assert.True(pipelineOne.Steps[2].IsResolved())

	// step transform.one_echo should not be resolved, it has reference to transform.one step
	assert.False(pipelineOne.Steps[3].IsResolved())

	assert.Equal("using value from locals: value of locals_one", pipelineOne.Steps[4].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("using value from locals: 10", pipelineOne.Steps[5].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("using value from locals: value of key_two", pipelineOne.Steps[6].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("using value from locals: value of key_two", pipelineOne.Steps[7].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("using value from locals: 33", pipelineOne.Steps[8].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("var_four value is: value from auto.vars file", pipelineOne.Steps[9].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("var_five value is: value from two.auto.vars file", pipelineOne.Steps[10].(*modconfig.PipelineStepTransform).Value)
	assert.Equal("var_six value is: set from env var", pipelineOne.Steps[11].(*modconfig.PipelineStepTransform).Value)

	githubIssuePipeline := pipelines["test_mod.pipeline.github_issue"]
	if githubIssuePipeline == nil {
		assert.Fail("github_issue pipeline not found")
		return
	}

	assert.Equal(1, len(githubIssuePipeline.Params))
	assert.NotNil(githubIssuePipeline.Params["gh_repo"])
	assert.Equal("hello-world", githubIssuePipeline.Params["gh_repo"].Default.AsString())

	githubGetIssueWithNumber := pipelines["test_mod.pipeline.github_get_issue_with_number"]
	if githubGetIssueWithNumber == nil {
		assert.Fail("github_get_issue_with_number pipeline not found")
		return
	}

	assert.Equal(2, len(githubGetIssueWithNumber.Params))
	assert.Equal("cty.String", githubGetIssueWithNumber.Params["github_token"].Type.GoString())
	assert.Equal("cty.Number", githubGetIssueWithNumber.Params["github_issue_number"].Type.GoString())

	triggers := mod.ResourceMaps.Triggers

	if len(triggers) == 0 {
		assert.Fail("triggers not loaded")
		return
	}

	reportTrigger := triggers["test_mod.trigger.schedule.report_triggers"]
	if reportTrigger == nil {
		assert.Fail("report_triggers not found")
		return
	}

	_, ok := reportTrigger.Config.(*modconfig.TriggerSchedule)
	assert.True(ok, "report_triggers is not of type TriggerSchedule")

	// Check the schedule with the default var
	reportTriggersWithScheduleVarWithDefaultValue := triggers["test_mod.trigger.schedule.report_triggers_with_schedule_var_with_default_value"]
	if reportTriggersWithScheduleVarWithDefaultValue == nil {
		assert.Fail("report_triggers_with_schedule_var_with_default_value not found")
		return
	}
	configSchedule := reportTriggersWithScheduleVarWithDefaultValue.Config.(*modconfig.TriggerSchedule)

	// This value is set in the pvar file
	assert.Equal("5 * * * *", configSchedule.Schedule)

	reportTriggersWithIntervalVarWithDefaultValue := triggers["test_mod.trigger.schedule.report_triggers_with_interval_var_with_default_value"]
	if reportTriggersWithIntervalVarWithDefaultValue == nil {
		assert.Fail("report_triggers_with_interval_var_with_default_value not found")
		return
	}

	intervalSchedule := reportTriggersWithIntervalVarWithDefaultValue.Config.(*modconfig.TriggerSchedule)
	assert.Equal("weekly", intervalSchedule.Schedule)

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeModTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeModTestSuite))
}
