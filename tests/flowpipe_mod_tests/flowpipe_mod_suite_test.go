package pipeline_test

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/tests/test_init"
	"github.com/turbot/pipe-fittings/utils"
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

func (suite *FlowpipeModTestSuite) TestModThrowConfig() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_throw_config", workspace.WithCredentials(map[string]credential.Credential{}))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	pipelines := w.Mod.ResourceMaps.Pipelines

	pipeline := pipelines["throw_config.pipeline.error_with_throw_does_not_ignore"]

	stepWithThrow := pipeline.Steps[1]

	// Message attribute is unresolved and refer to itself (using the "result" attribute)
	assert.NotNil(stepWithThrow.GetThrowConfig()[0].UnresolvedAttributes[schema.AttributeTypeMessage])
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

func (suite *FlowpipeModTestSuite) TestFlowpipeConfigInvalidIntegration() {
	assert := assert.New(suite.T())

	// Reading from different file will always result in different config
	_, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_dir_invalid_integration"})
	assert.NotNil(err.Error)
}

func (suite *FlowpipeModTestSuite) TestFlowpipeConfigEquality() {
	assert := assert.New(suite.T())

	// Reading from different file will always result in different config
	flowpipeConfigA, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_equality_test_a"})
	assert.Nil(err.Error)

	flowpipeConfigA2, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_equality_test_a"})
	assert.Nil(err.Error)

	assert.True(flowpipeConfigA.Equals(flowpipeConfigA2))

	flowpipeConfigB, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_equality_test_b"})
	assert.Nil(err.Error)

	assert.False(flowpipeConfigA.Equals(flowpipeConfigB))

	utils.EmptyDir("./config_equality_test_dir")                            //nolint:errcheck // test only
	utils.CopyDir("./config_equality_test_a", "./config_equality_test_dir") //nolint:errcheck // test only

	flowpipeConfigA, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_equality_test_dir"})
	assert.Nil(err.Error)

	utils.EmptyDir("./config_equality_test_dir")                            //nolint:errcheck // test only
	utils.CopyDir("./config_equality_test_b", "./config_equality_test_dir") //nolint:errcheck // test only

	flowpipeConfigB, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_equality_test_dir"})
	assert.Nil(err.Error)

	assert.False(flowpipeConfigA.Equals(flowpipeConfigB))
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

func (suite *FlowpipeModTestSuite) TestModWithCredsInOutput() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_with_creds_output"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_creds_output", workspace.WithCredentials(flowpipeConfig.Credentials))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	credentials := w.Credentials
	awsExampleCreds := credentials["aws.example"]
	slackCredsCty, e := awsExampleCreds.CtyValue()
	assert.Nil(e)

	credsMap := slackCredsCty.AsValueMap()
	accessKeyVal := credsMap["access_key"].AsString()
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", accessKeyVal)

	pipeline := w.Mod.ResourceMaps.Pipelines["test_mod.pipeline.cred_in_step_output"]
	assert.NotNil(pipeline)

	step := pipeline.Steps[0]
	assert.Equal(1, len(step.GetCredentialDependsOn()))
	assert.Equal("aws.example", step.GetCredentialDependsOn()[0])

	pipeline = w.Mod.ResourceMaps.Pipelines["test_mod.pipeline.cred_in_output"]
	assert.NotNil(pipeline)

	assert.Equal(1, len(pipeline.OutputConfig))
	assert.Equal("aws.example", pipeline.OutputConfig[0].CredentialDependsOn[0])

}

func (suite *FlowpipeModTestSuite) TestModIntegrationNotifierParam() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_integration_notifier_param"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_integration_notifier_param", workspace.WithCredentials(flowpipeConfig.Credentials))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	pipeline := w.Mod.ResourceMaps.Pipelines["mod_integration_notifier_param.pipeline.integration_pipe_default_with_param"]
	unresolvedAttributes := pipeline.Steps[0].GetUnresolvedAttributes()
	assert.Equal(1, len(unresolvedAttributes))
	assert.NotNil(unresolvedAttributes["notifier"])
}

func (suite *FlowpipeModTestSuite) TestModSimpleInputStep() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_with_input_step_simple"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_with_input_step_simple", workspace.WithCredentials(flowpipeConfig.Credentials), workspace.WithNotifiers(flowpipeConfig.Notifiers))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	pipeline := w.Mod.ResourceMaps.Pipelines["mod_with_input_step_simple.pipeline.simple_input_step"]

	step := pipeline.Steps[0]
	inputStep := step.(*modconfig.PipelineStepInput)

	assert.Equal(2, len(inputStep.OptionList))

	assert.Equal("Approve", *inputStep.OptionList[0].OptionLabel)
	assert.Equal("Deny", *inputStep.OptionList[1].OptionLabel)

	pipeline = w.Mod.ResourceMaps.Pipelines["mod_with_input_step_simple.pipeline.simple_input_step_with_option_list"]

	step = pipeline.Steps[0]
	inputStep = step.(*modconfig.PipelineStepInput)

	assert.Equal(2, len(inputStep.OptionList))

	assert.Equal("N. Virginia", *inputStep.OptionList[0].Label)
	assert.Equal("Ohio", *inputStep.OptionList[1].Label)
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

	assert.Equal("bar", *notifier.GetHclResourceImpl().Description)
	assert.Equal("dev notifier", *notifier.GetHclResourceImpl().Title)

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

func (suite *FlowpipeModTestSuite) TestFlowpipeModWithOneIntegration() {
	assert := assert.New(suite.T())

	flowpipeConfig, ew := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_with_one_notifier"})
	if ew.Error != nil {
		assert.FailNow(ew.Error.Error())
		return
	}

	if flowpipeConfig == nil {
		assert.Fail("flowpipeConfig is nil")
		return
	}

	notifier := flowpipeConfig.Notifiers["notify_one"]

	assert.Equal("foo", *notifier.GetHclResourceImpl().Description)
	assert.Equal("foo bar", *notifier.GetHclResourceImpl().Title)
}

func (suite *FlowpipeModTestSuite) TestFlowpipeConfigIntegrationEmail() {
	assert := assert.New(suite.T())

	// the order of directories matter because we determine which one has precedent. the "admins" notifier used will be the one defined in config_dir_more_integrations
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
	assert.Equal("user@test.tld", integrations.(*modconfig.EmailIntegration).To[0])
	assert.Equal("turbie@flowpipe.io", *integrations.(*modconfig.EmailIntegration).From)
	assert.Equal("email.email_with_all", integrations.GetHclResourceImpl().FullName)
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

	// Alicloud
	assert.Equal("steampipe_alicloud", flowpipeConfig.CredentialImports["steampipe_alicloud"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_alicloud"].Prefix)
	assert.Equal("alicloud.sp1_alicloud_1", flowpipeConfig.Credentials["alicloud.sp1_alicloud_1"].GetHclResourceImpl().FullName)
	assert.Equal("alicloud.sp1_alicloud_2", flowpipeConfig.Credentials["alicloud.sp1_alicloud_2"].GetHclResourceImpl().FullName)
	assert.Equal("XXXXGBV", *flowpipeConfig.Credentials["alicloud.sp1_alicloud_1"].(*credential.AlicloudCredential).AccessKey)
	assert.Equal("6iNPvThisIsNotARealSecretk1sZF", *flowpipeConfig.Credentials["alicloud.sp1_alicloud_1"].(*credential.AlicloudCredential).SecretKey)
	assert.Equal("XXXXGBV", *flowpipeConfig.Credentials["alicloud.sp1_alicloud_2"].(*credential.AlicloudCredential).AccessKey)
	assert.Equal("6iNPvThisIsNotARealSecretk1sZF", *flowpipeConfig.Credentials["alicloud.sp1_alicloud_2"].(*credential.AlicloudCredential).SecretKey)

	// AWS
	assert.Equal("steampipe", flowpipeConfig.CredentialImports["steampipe"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe"].Prefix)
	assert.Equal("aws.sp1_aws", flowpipeConfig.Credentials["aws.sp1_aws"].GetHclResourceImpl().FullName)
	assert.Equal("aws.sp1_aws_keys1", flowpipeConfig.Credentials["aws.sp1_aws_keys1"].GetHclResourceImpl().FullName)
	assert.Equal("abc", *flowpipeConfig.Credentials["aws.sp1_aws_keys1"].(*credential.AwsCredential).AccessKey)
	assert.Equal("123", *flowpipeConfig.Credentials["aws.sp1_aws_keys1"].(*credential.AwsCredential).SecretKey)

	// Azure
	assert.Equal("steampipe_azure", flowpipeConfig.CredentialImports["steampipe_azure"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_azure"].Prefix)
	assert.Equal("azure.sp1_azure_1", flowpipeConfig.Credentials["azure.sp1_azure_1"].GetHclResourceImpl().FullName)
	assert.Equal("azure.sp1_azure_2", flowpipeConfig.Credentials["azure.sp1_azure_2"].GetHclResourceImpl().FullName)
	assert.Equal("00000000-0000-0000-0000-000000000000", *flowpipeConfig.Credentials["azure.sp1_azure_1"].(*credential.AzureCredential).ClientID)
	assert.Equal("~dummy@3password", *flowpipeConfig.Credentials["azure.sp1_azure_1"].(*credential.AzureCredential).ClientSecret)
	assert.Nil(flowpipeConfig.Credentials["azure.sp1_azure_1"].(*credential.AzureCredential).Environment)
	assert.Equal("00000000-0000-0000-0000-000000000000", *flowpipeConfig.Credentials["azure.sp1_azure_1"].(*credential.AzureCredential).TenantID)
	assert.Equal("00000000-0000-0000-0000-000000000000", *flowpipeConfig.Credentials["azure.sp1_azure_2"].(*credential.AzureCredential).ClientID)
	assert.Equal("~dummy@3password", *flowpipeConfig.Credentials["azure.sp1_azure_2"].(*credential.AzureCredential).ClientSecret)
	assert.Equal("AZUREUSGOVERNMENTCLOUD", *flowpipeConfig.Credentials["azure.sp1_azure_2"].(*credential.AzureCredential).Environment)
	assert.Equal("00000000-0000-0000-0000-000000000000", *flowpipeConfig.Credentials["azure.sp1_azure_2"].(*credential.AzureCredential).TenantID)

	// Bitbucket
	assert.Equal("steampipe_bitbucket", flowpipeConfig.CredentialImports["steampipe_bitbucket"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_bitbucket"].Prefix)
	assert.Equal("bitbucket.sp1_bitbucket_1", flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_1"].GetHclResourceImpl().FullName)
	assert.Equal("bitbucket.sp1_bitbucket_2", flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_2"].GetHclResourceImpl().FullName)
	assert.Equal("https://api.bitbucket.org/2.0", *flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_1"].(*credential.BitbucketCredential).BaseURL)
	assert.Equal("blHdmvlkFakeToken1", *flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_1"].(*credential.BitbucketCredential).Password)
	assert.Equal("MyUsername1", *flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_1"].(*credential.BitbucketCredential).Username)
	assert.Equal("https://api.bitbucket.org/2.0", *flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_2"].(*credential.BitbucketCredential).BaseURL)
	assert.Equal("blHdmvlkFakeToken2", *flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_2"].(*credential.BitbucketCredential).Password)
	assert.Equal("MyUsername2", *flowpipeConfig.Credentials["bitbucket.sp1_bitbucket_2"].(*credential.BitbucketCredential).Username)

	// ClickUp
	assert.Equal("steampipe_clickup", flowpipeConfig.CredentialImports["steampipe_clickup"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_clickup"].Prefix)
	assert.Equal("clickup.sp1_clickup_1", flowpipeConfig.Credentials["clickup.sp1_clickup_1"].GetHclResourceImpl().FullName)
	assert.Equal("clickup.sp1_clickup_2", flowpipeConfig.Credentials["clickup.sp1_clickup_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["clickup.sp1_clickup_1"].(*credential.ClickUpCredential).Token)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["clickup.sp1_clickup_2"].(*credential.ClickUpCredential).Token)

	// Datadog
	assert.Equal("steampipe_datadog", flowpipeConfig.CredentialImports["steampipe_datadog"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_datadog"].Prefix)
	assert.Equal("datadog.sp1_datadog_1", flowpipeConfig.Credentials["datadog.sp1_datadog_1"].GetHclResourceImpl().FullName)
	assert.Equal("datadog.sp1_datadog_2", flowpipeConfig.Credentials["datadog.sp1_datadog_2"].GetHclResourceImpl().FullName)
	assert.Equal("1a2345bc6d78e9d98fa7bcd6e5ef56a7", *flowpipeConfig.Credentials["datadog.sp1_datadog_1"].(*credential.DatadogCredential).APIKey)
	assert.Equal("https://api.datadoghq.com/", *flowpipeConfig.Credentials["datadog.sp1_datadog_1"].(*credential.DatadogCredential).APIUrl)
	assert.Equal("b1cf234c0ed4c567890b524a3b42f1bd91c111a1", *flowpipeConfig.Credentials["datadog.sp1_datadog_1"].(*credential.DatadogCredential).AppKey)
	assert.Equal("1a2345bc6d78e9d98fa7bcd6e5ef57b8", *flowpipeConfig.Credentials["datadog.sp1_datadog_2"].(*credential.DatadogCredential).APIKey)
	assert.Equal("https://api.datadoghq.com/", *flowpipeConfig.Credentials["datadog.sp1_datadog_2"].(*credential.DatadogCredential).APIUrl)
	assert.Equal("b1cf234c0ed4c567890b524a3b42f1bd91c222b2", *flowpipeConfig.Credentials["datadog.sp1_datadog_2"].(*credential.DatadogCredential).AppKey)

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

	// GCP
	assert.Equal("steampipe_gcp", flowpipeConfig.CredentialImports["steampipe_gcp"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_gcp"].Prefix)
	assert.Equal("gcp.sp1_gcp_1", flowpipeConfig.Credentials["gcp.sp1_gcp_1"].GetHclResourceImpl().FullName)
	assert.Equal("gcp.sp1_gcp_2", flowpipeConfig.Credentials["gcp.sp1_gcp_2"].GetHclResourceImpl().FullName)
	assert.Equal("/home/me/my-service-account-creds-for-project-aaa.json", *flowpipeConfig.Credentials["gcp.sp1_gcp_1"].(*credential.GcpCredential).Credentials)
	assert.Equal("/home/me/my-service-account-creds-for-project-bbb.json", *flowpipeConfig.Credentials["gcp.sp1_gcp_2"].(*credential.GcpCredential).Credentials)

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

	// Guardrails
	assert.Equal("steampipe_guardrails", flowpipeConfig.CredentialImports["steampipe_guardrails"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_guardrails"].Prefix)
	assert.Equal("guardrails.sp1_guardrails_1", flowpipeConfig.Credentials["guardrails.sp1_guardrails_1"].GetHclResourceImpl().FullName)
	assert.Equal("guardrails.sp1_guardrails_2", flowpipeConfig.Credentials["guardrails.sp1_guardrails_2"].GetHclResourceImpl().FullName)
	assert.Equal("c8e2c2ed-1ca8-429b-b369-010e3cf75aac", *flowpipeConfig.Credentials["guardrails.sp1_guardrails_1"].(*credential.GuardrailsCredential).AccessKey)
	assert.Equal("a3d8385d-47f7-40c5-a90c-bfdf5b43c8dd", *flowpipeConfig.Credentials["guardrails.sp1_guardrails_1"].(*credential.GuardrailsCredential).SecretKey)
	assert.Equal("https://turbot-acme.cloud.turbot.com/", *flowpipeConfig.Credentials["guardrails.sp1_guardrails_1"].(*credential.GuardrailsCredential).Workspace)
	assert.Equal("c8e2c2ed-1ca8-429b-b369-010e3cf75aac", *flowpipeConfig.Credentials["guardrails.sp1_guardrails_2"].(*credential.GuardrailsCredential).AccessKey)
	assert.Equal("a3d8385d-47f7-40c5-a90c-bfdf5b43c8dd", *flowpipeConfig.Credentials["guardrails.sp1_guardrails_2"].(*credential.GuardrailsCredential).SecretKey)
	assert.Equal("https://turbot-acme.cloud.turbot.com/", *flowpipeConfig.Credentials["guardrails.sp1_guardrails_2"].(*credential.GuardrailsCredential).Workspace)

	// IP2LocationIO
	assert.Equal("steampipe_ip2locationio", flowpipeConfig.CredentialImports["steampipe_ip2locationio"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_ip2locationio"].Prefix)
	assert.Equal("ip2locationio.sp1_ip2locationio_1", flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_1"].GetHclResourceImpl().FullName)
	assert.Equal("ip2locationio.sp1_ip2locationio_2", flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_1"].(*credential.IP2LocationIOCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["ip2locationio.sp1_ip2locationio_2"].(*credential.IP2LocationIOCredential).APIKey)

	// IPstack
	assert.Equal("steampipe_ipstack", flowpipeConfig.CredentialImports["steampipe_ipstack"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_ipstack"].Prefix)
	assert.Equal("ipstack.sp1_ipstack_1", flowpipeConfig.Credentials["ipstack.sp1_ipstack_1"].GetHclResourceImpl().FullName)
	assert.Equal("ipstack.sp1_ipstack_2", flowpipeConfig.Credentials["ipstack.sp1_ipstack_2"].GetHclResourceImpl().FullName)
	assert.Equal("e0067f483763d6132d934864f8a6de22", *flowpipeConfig.Credentials["ipstack.sp1_ipstack_1"].(*credential.IPstackCredential).AccessKey)
	assert.Equal("e0067f483763d6132d934864f8a6de22", *flowpipeConfig.Credentials["ipstack.sp1_ipstack_2"].(*credential.IPstackCredential).AccessKey)

	// Jira
	assert.Equal("steampipe_jira", flowpipeConfig.CredentialImports["steampipe_jira"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_jira"].Prefix)
	assert.Equal("jira.sp1_jira_1", flowpipeConfig.Credentials["jira.sp1_jira_1"].GetHclResourceImpl().FullName)
	assert.Equal("jira.sp1_jira_2", flowpipeConfig.Credentials["jira.sp1_jira_2"].GetHclResourceImpl().FullName)
	assert.Equal("jira.sp1_jira_3", flowpipeConfig.Credentials["jira.sp1_jira_3"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["jira.sp1_jira_1"].(*credential.JiraCredential).APIToken)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["jira.sp1_jira_2"].(*credential.JiraCredential).APIToken)
	assert.Equal("abcdefgj", *flowpipeConfig.Credentials["jira.sp1_jira_3"].(*credential.JiraCredential).APIToken)

	// JumpCloud
	assert.Equal("steampipe_jumpcloud", flowpipeConfig.CredentialImports["steampipe_jumpcloud"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_jumpcloud"].Prefix)
	assert.Equal("jumpcloud.sp1_jumpcloud_1", flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_1"].GetHclResourceImpl().FullName)
	assert.Equal("jumpcloud.sp1_jumpcloud_2", flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_1"].(*credential.JumpCloudCredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["jumpcloud.sp1_jumpcloud_2"].(*credential.JumpCloudCredential).APIKey)

	// Mastodon
	assert.Equal("steampipe_mastodon", flowpipeConfig.CredentialImports["steampipe_mastodon"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_mastodon"].Prefix)
	assert.Equal("mastodon.sp1_mastodon_1", flowpipeConfig.Credentials["mastodon.sp1_mastodon_1"].GetHclResourceImpl().FullName)
	assert.Equal("mastodon.sp1_mastodon_2", flowpipeConfig.Credentials["mastodon.sp1_mastodon_2"].GetHclResourceImpl().FullName)
	assert.Equal("FK1_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU", *flowpipeConfig.Credentials["mastodon.sp1_mastodon_1"].(*credential.MastodonCredential).AccessToken)
	assert.Equal("https://myserver.social", *flowpipeConfig.Credentials["mastodon.sp1_mastodon_1"].(*credential.MastodonCredential).Server)
	assert.Equal("FK2_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU", *flowpipeConfig.Credentials["mastodon.sp1_mastodon_2"].(*credential.MastodonCredential).AccessToken)
	assert.Equal("https://myserver.social", *flowpipeConfig.Credentials["mastodon.sp1_mastodon_2"].(*credential.MastodonCredential).Server)

	// Microsoft Teams
	assert.Equal("steampipe_teams", flowpipeConfig.CredentialImports["steampipe_teams"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_teams"].Prefix)
	assert.Equal("teams.sp1_teams_1", flowpipeConfig.Credentials["teams.sp1_teams_1"].GetHclResourceImpl().FullName)
	assert.Equal("teams.sp1_teams_2", flowpipeConfig.Credentials["teams.sp1_teams_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["teams.sp1_teams_1"].(*credential.MicrosoftTeamsCredential).AccessToken)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["teams.sp1_teams_2"].(*credential.MicrosoftTeamsCredential).AccessToken)

	// Okta
	assert.Equal("steampipe_okta", flowpipeConfig.CredentialImports["steampipe_okta"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_okta"].Prefix)
	assert.Equal("okta.sp1_okta_1", flowpipeConfig.Credentials["okta.sp1_okta_1"].GetHclResourceImpl().FullName)
	assert.Equal("okta.sp1_okta_2", flowpipeConfig.Credentials["okta.sp1_okta_2"].GetHclResourceImpl().FullName)
	assert.Equal("https://test1.okta.com", *flowpipeConfig.Credentials["okta.sp1_okta_1"].(*credential.OktaCredential).Domain)
	assert.Equal("testtoken", *flowpipeConfig.Credentials["okta.sp1_okta_1"].(*credential.OktaCredential).Token)
	assert.Equal("https://test2.okta.com", *flowpipeConfig.Credentials["okta.sp1_okta_2"].(*credential.OktaCredential).Domain)
	assert.Equal("testtoken", *flowpipeConfig.Credentials["okta.sp1_okta_2"].(*credential.OktaCredential).Token)

	// OpenAI
	assert.Equal("steampipe_openai", flowpipeConfig.CredentialImports["steampipe_openai"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_openai"].Prefix)
	assert.Equal("openai.sp1_openai_1", flowpipeConfig.Credentials["openai.sp1_openai_1"].GetHclResourceImpl().FullName)
	assert.Equal("openai.sp1_openai_2", flowpipeConfig.Credentials["openai.sp1_openai_2"].GetHclResourceImpl().FullName)
	assert.Equal("abcdefgh", *flowpipeConfig.Credentials["openai.sp1_openai_1"].(*credential.OpenAICredential).APIKey)
	assert.Equal("abcdefgi", *flowpipeConfig.Credentials["openai.sp1_openai_2"].(*credential.OpenAICredential).APIKey)

	// Opsgenie
	assert.Equal("steampipe_opsgenie", flowpipeConfig.CredentialImports["steampipe_opsgenie"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_opsgenie"].Prefix)
	assert.Equal("opsgenie.sp1_opsgenie_1", flowpipeConfig.Credentials["opsgenie.sp1_opsgenie_1"].GetHclResourceImpl().FullName)
	assert.Equal("opsgenie.sp1_opsgenie_2", flowpipeConfig.Credentials["opsgenie.sp1_opsgenie_2"].GetHclResourceImpl().FullName)
	assert.Equal("alertapikey1", *flowpipeConfig.Credentials["opsgenie.sp1_opsgenie_1"].(*credential.OpsgenieCredential).AlertAPIKey)
	assert.Equal("incidentapikey1", *flowpipeConfig.Credentials["opsgenie.sp1_opsgenie_1"].(*credential.OpsgenieCredential).IncidentAPIKey)
	assert.Equal("alertapikey2", *flowpipeConfig.Credentials["opsgenie.sp1_opsgenie_2"].(*credential.OpsgenieCredential).AlertAPIKey)
	assert.Equal("incidentapikey2", *flowpipeConfig.Credentials["opsgenie.sp1_opsgenie_2"].(*credential.OpsgenieCredential).IncidentAPIKey)

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

	// Zendesk
	assert.Equal("steampipe_zendesk", flowpipeConfig.CredentialImports["steampipe_zendesk"].FullName)
	assert.Equal("sp1_", *flowpipeConfig.CredentialImports["steampipe_zendesk"].Prefix)
	assert.Equal("zendesk.sp1_zendesk_1", flowpipeConfig.Credentials["zendesk.sp1_zendesk_1"].GetHclResourceImpl().FullName)
	assert.Equal("zendesk.sp1_zendesk_2", flowpipeConfig.Credentials["zendesk.sp1_zendesk_2"].GetHclResourceImpl().FullName)
	assert.Equal("pam@dmi.com", *flowpipeConfig.Credentials["zendesk.sp1_zendesk_1"].(*credential.ZendeskCredential).Email)
	assert.Equal("dmi", *flowpipeConfig.Credentials["zendesk.sp1_zendesk_1"].(*credential.ZendeskCredential).Subdomain)
	assert.Equal("17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4", *flowpipeConfig.Credentials["zendesk.sp1_zendesk_1"].(*credential.ZendeskCredential).Token)
	assert.Equal("pam@dmj.com", *flowpipeConfig.Credentials["zendesk.sp1_zendesk_2"].(*credential.ZendeskCredential).Email)
	assert.Equal("dmj", *flowpipeConfig.Credentials["zendesk.sp1_zendesk_2"].(*credential.ZendeskCredential).Subdomain)
	assert.Equal("17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4", *flowpipeConfig.Credentials["zendesk.sp1_zendesk_2"].(*credential.ZendeskCredential).Token)
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
	assert.Equal("my slack app in config_dir with description", *flowpipeConfig.Integrations["slack.my_slack_app"].GetHclResourceImpl().Description)

	// ensure that the default integration exist
	assert.Equal("http.default", flowpipeConfig.Integrations["http.default"].GetHclResourceImpl().FullName)

	assert.Equal(4, len(flowpipeConfig.Notifiers))

	notifierWithDefaultIntegration := flowpipeConfig.Notifiers["with_default_integration"]
	if notifierWithDefaultIntegration == nil {
		assert.Fail("notifier with_default_integration not found")
		return
	}

	assert.Equal("with_default_integration", notifierWithDefaultIntegration.GetHclResourceImpl().FullName)
	assert.Equal(1, len(notifierWithDefaultIntegration.GetNotifies()))
	assert.Equal("http.default", notifierWithDefaultIntegration.GetNotifies()[0].Integration.(*modconfig.HttpIntegration).FullName)

	// ensure that default notifier exist
	assert.Equal("default", flowpipeConfig.Notifiers["default"].GetHclResourceImpl().FullName)
	assert.Equal(1, len(flowpipeConfig.Notifiers["default"].GetNotifies()))

	// TODO: test this when we have http up and running
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

	pipeline = pipelines["mod_with_integration.pipeline.approval_with_override_in_step"]
	if pipeline == nil {
		assert.Fail("pipeline approval_with_override_in_step not found")
		return
	}

	step, ok = pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Step is not an input step")
		return
	}

	assert.Equal("this subject is in step", *step.Subject)
	assert.Equal("this channel is in step override", *step.Channel)

	assert.True(helpers.StringSliceEqualIgnoreOrder(step.To, []string{"foo", "bar", "baz override"}))
	assert.True(helpers.StringSliceEqualIgnoreOrder(step.Cc, []string{"foo", "bar", "baz cc"}))
	assert.True(helpers.StringSliceEqualIgnoreOrder(step.Bcc, []string{"foo bb", "bar", "baz override"}))
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

func (suite *FlowpipeModTestSuite) TestModMessageStep() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_message_step"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_message_step", workspace.WithCredentials(flowpipeConfig.Credentials),
		workspace.WithIntegrations(flowpipeConfig.Integrations), workspace.WithNotifiers(flowpipeConfig.Notifiers))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipeline := mod.ResourceMaps.Pipelines["mod_message_step.pipeline.message_step_one"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	messageStepInterface := pipeline.Steps[0]
	if messageStepInterface == nil {
		assert.Fail("message step not found")
		return
	}

	messageStep, ok := messageStepInterface.(*modconfig.PipelineStepMessage)
	if !ok {
		assert.Fail("message step is not of type PipelineStepMessage")
		return
	}

	assert.Equal("Hello World", messageStep.Text)

	pipeline = mod.ResourceMaps.Pipelines["mod_message_step.pipeline.message_step_with_overrides"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	messageStepInterface = pipeline.Steps[0]
	if messageStepInterface == nil {
		assert.Fail("message step not found")
		return
	}

	messageStep, ok = messageStepInterface.(*modconfig.PipelineStepMessage)
	if !ok {
		assert.Fail("message step is not of type PipelineStepMessage")
		return
	}

	assert.Equal("Hello World 2", messageStep.Text)
	assert.Equal("channel override", *messageStep.Channel)
	assert.True(helpers.StringSliceEqualIgnoreOrder([]string{"foo", "baz"}, messageStep.Cc))
	assert.True(helpers.StringSliceEqualIgnoreOrder([]string{"bar"}, messageStep.Bcc))
}

func (suite *FlowpipeModTestSuite) TestModDynamicPipeRef() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_dynamic_pipeline_ref"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_dynamic_pipeline_ref", workspace.WithCredentials(flowpipeConfig.Credentials),
		workspace.WithIntegrations(flowpipeConfig.Integrations), workspace.WithNotifiers(flowpipeConfig.Notifiers))

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipeline := mod.ResourceMaps.Pipelines["dynamic_pipe_ref.pipeline.top_dynamic"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	steps := pipeline.Steps
	assert.Equal("middle_dynamic_static_to_a", steps[0].GetName())

	// the second step has a dynamic pipeline reference
	assert.NotNil(steps[1].GetUnresolvedAttributes()["pipeline"])
}

func (suite *FlowpipeModTestSuite) TestModTryFunction() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_try_function"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_try_function", workspace.WithCredentials(flowpipeConfig.Credentials), workspace.WithNotifiers(flowpipeConfig.Notifiers))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	pipeline := w.Mod.ResourceMaps.Pipelines["test.pipeline.try_function"]
	assert.NotNil(pipeline)

	assert.NotNil(pipeline.Steps[0].GetUnresolvedAttributes()["value"])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.max_function"]
	assert.NotNil(pipeline)

	assert.NotNil(pipeline.Steps[0].GetUnresolvedAttributes()["value"])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.try_function_no_for_each"]
	assert.NotNil(pipeline)
	assert.NotNil(pipeline.Steps[0].GetUnresolvedAttributes()["value"])
	assert.NotNil(pipeline.Steps[1].GetUnresolvedAttributes()["value"])
	assert.Equal("transform.first", pipeline.Steps[1].GetDependsOn()[0])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.try_function_no_for_each_combination_1"]
	assert.NotNil(pipeline)
	assert.NotNil(pipeline.Steps[0].GetUnresolvedAttributes()["value"])
	assert.NotNil(pipeline.Steps[1].GetUnresolvedAttributes()["value"])
	assert.Equal("transform.first", pipeline.Steps[1].GetDependsOn()[0])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.try_function_no_for_each_combination_2"]
	assert.NotNil(pipeline)
	assert.NotNil(pipeline.Steps[0].GetUnresolvedAttributes()["value"])
	// the second step (number) should not have any unresolved attributes
	assert.Nil(pipeline.Steps[1].GetUnresolvedAttributes()["value"])

	assert.NotNil(pipeline.Steps[2].GetUnresolvedAttributes()["value"])
	assert.Equal("transform.first", pipeline.Steps[2].GetDependsOn()[0])
	assert.Equal("transform.number", pipeline.Steps[2].GetDependsOn()[1])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.try_function_within_json_encode"]
	assert.NotNil(pipeline)
	// step 0 -> transform.nexus
	// step 1 -> the http step
	assert.NotNil(pipeline.Steps[1].GetUnresolvedAttributes()["request_body"])
	assert.Equal("transform.nexus", pipeline.Steps[1].GetDependsOn()[0])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.try_function_from_param"]
	assert.NotNil(pipeline)
	assert.NotNil(pipeline.Steps[0].GetUnresolvedAttributes()["value"])
}

func (suite *FlowpipeModTestSuite) TestInputStepWithThrow() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./input_step_with_throw"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./input_step_with_throw", workspace.WithCredentials(flowpipeConfig.Credentials), workspace.WithNotifiers(flowpipeConfig.Notifiers))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)
}

func (suite *FlowpipeModTestSuite) TestInputStepWithLoop() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./input_step_with_loop"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./input_step_with_loop", workspace.WithCredentials(flowpipeConfig.Credentials), workspace.WithNotifiers(flowpipeConfig.Notifiers))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)
}

func (suite *FlowpipeModTestSuite) TestLoopVarious() {
	assert := assert.New(suite.T())

	flowpipeConfig, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./mod_loop_various"})
	assert.Nil(err.Error)

	w, errorAndWarning := workspace.Load(suite.ctx, "./mod_loop_various", workspace.WithCredentials(flowpipeConfig.Credentials), workspace.WithNotifiers(flowpipeConfig.Notifiers))
	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	pipeline := w.Mod.ResourceMaps.Pipelines["test.pipeline.sleep"]
	assert.NotNil(pipeline)
	step := pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.sleep_2"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Equal("10s", *step.GetLoopConfig().(*modconfig.LoopSleepStep).Duration)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.sleep_3"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeDuration])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.sleep_4"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeDuration])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.http"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Equal("https://bar", *step.GetLoopConfig().(*modconfig.LoopHttpStep).URL)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.http_2"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUrl])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.container"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeMemory])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.container_2"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeMemory])
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopContainerStep).Cmd)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.container_3"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeMemory])
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopContainerStep).Cmd)
	assert.Equal([]string{"1", "2"}, *step.GetLoopConfig().(*modconfig.LoopContainerStep).Entrypoint)
	assert.Equal(int64(4), *step.GetLoopConfig().(*modconfig.LoopContainerStep).CpuShares)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.container_4"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeMemory])
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopContainerStep).Cmd)
	assert.Equal([]string{"1", "2"}, *step.GetLoopConfig().(*modconfig.LoopContainerStep).Entrypoint)
	assert.Equal(int64(4), *step.GetLoopConfig().(*modconfig.LoopContainerStep).CpuShares)
	assert.Equal(map[string]string{"bar": "baz"}, *step.GetLoopConfig().(*modconfig.LoopContainerStep).Env)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.pipeline"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeArgs])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.pipeline_2"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Nil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeArgs])
	assert.Equal(map[string]interface{}{"a": "foo_10", "c": 44}, step.GetLoopConfig().(*modconfig.LoopPipelineStep).Args.(map[string]interface{}))

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.pipeline_3"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Nil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeArgs])
	assert.Equal(map[string]interface{}{"a": "foo_10", "c": 44}, step.GetLoopConfig().(*modconfig.LoopPipelineStep).Args.(map[string]interface{}))

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.query"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Nil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeArgs])
	assert.Equal([]interface{}{"bar"}, *step.GetLoopConfig().(*modconfig.LoopQueryStep).Args)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.query_2"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeArgs])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.message"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Equal("I'm a sample message two", *step.GetLoopConfig().(*modconfig.LoopMessageStep).Text)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.message_2"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Equal("I'm a sample message two", *step.GetLoopConfig().(*modconfig.LoopMessageStep).Text)
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopMessageStep).To)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.message_3"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Equal("I'm a sample message two", *step.GetLoopConfig().(*modconfig.LoopMessageStep).Text)
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopMessageStep).To)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.message_4"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeBcc])
	assert.Equal("I'm a sample message two", *step.GetLoopConfig().(*modconfig.LoopMessageStep).Text)
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopMessageStep).To)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.message_5"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeBcc])
	assert.Equal("I'm a sample message two", *step.GetLoopConfig().(*modconfig.LoopMessageStep).Text)
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopMessageStep).To)
	assert.Equal("new", step.GetLoopConfig().(*modconfig.LoopMessageStep).Notifier.GetHclResourceImpl().FullName)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.message_6"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeNotifier])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeBcc])
	assert.Equal("I'm a sample message two", *step.GetLoopConfig().(*modconfig.LoopMessageStep).Text)
	assert.Equal([]string{"a", "b", "c"}, *step.GetLoopConfig().(*modconfig.LoopMessageStep).To)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.input"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Equal("Shall we play a game 2?", *step.GetLoopConfig().(*modconfig.LoopInputStep).Prompt)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.input_2"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeNotifier])
	assert.Equal("Shall we play a game 2?", *step.GetLoopConfig().(*modconfig.LoopInputStep).Prompt)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.function"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.function_3"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.Equal(map[string]string{"restrictedActions": "def", "foo": "bar"}, *step.GetLoopConfig().(*modconfig.LoopFunctionStep).Env)
	assert.Equal(map[string]interface{}{"a": "c", "c": 44}, *step.GetLoopConfig().(*modconfig.LoopFunctionStep).Event)

	pipeline = w.Mod.ResourceMaps.Pipelines["test.pipeline.function_4"]
	assert.NotNil(pipeline)
	step = pipeline.Steps[0]
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeUntil])
	assert.NotNil(step.GetLoopConfig().GetUnresolvedAttributes()[schema.AttributeTypeEvent])
	assert.Equal(map[string]string{"restrictedActions": "def", "foo": "bar"}, *step.GetLoopConfig().(*modconfig.LoopFunctionStep).Env)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeModTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeModTestSuite))
}
