package pipeline_test

import (
	"context"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/turbot/pipe-fittings/tests/test_init"

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

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./good_mod", map[string]modconfig.Credential{}, ".fp")

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

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_references", map[string]modconfig.Credential{}, ".fp")

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

	os.Setenv("TOKEN", "foobarbaz")
	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_with_creds", map[string]modconfig.Credential{}, ".fp")

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
	os.Unsetenv("TOKEN")
}

func (suite *FlowpipeModTestSuite) TestModWithCredsNoEnvVarSet() {
	assert := assert.New(suite.T())

	// This is the same test with TestModWithCreds but with no TOKEN env var set, the value for the second step should be nil
	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_with_creds", map[string]modconfig.Credential{}, ".fp")

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

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_with_dynamic_creds", map[string]modconfig.Credential{}, ".fp")

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

func (suite *FlowpipeModTestSuite) TestStepOutputParsing() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_with_step_output", map[string]modconfig.Credential{}, ".fp")

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

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_dep_one", map[string]modconfig.Credential{}, ".fp")

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

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_dep_simple", map[string]modconfig.Credential{}, ".fp")

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

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_variable", map[string]modconfig.Credential{}, ".fp")

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

	reportTriggersWithIntervalVarWithDefaultValue := triggers["test_mod.trigger.interval.report_triggers_with_interval_var_with_default_value"]
	if reportTriggersWithIntervalVarWithDefaultValue == nil {
		assert.Fail("report_triggers_with_interval_var_with_default_value not found")
		return
	}

	intervalSchedule := reportTriggersWithIntervalVarWithDefaultValue.Config.(*modconfig.TriggerInterval)
	assert.Equal("weekly", intervalSchedule.Schedule)

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeModTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeModTestSuite))
}
