package flowpipe_mod_load_tests

import (
	"context"
	"github.com/turbot/pipe-fittings/app_specific"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeModLoadTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeModLoadTestSuite) SetupSuite() {

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

	app_specific.WorkspaceDataDir = ".flowpipe"
	app_specific.ModFileName = "mod.hcl"
	app_specific.DefaultVarsFileName = "flowpipe.pvars"
	app_specific.DefaultInstallDir = "~/.flowpipe"

	app_specific.ModDataExtension = ".hcl"
	app_specific.VariablesExtension = ".pvars"
	app_specific.AutoVariablesExtension = ".auto.pvars"
	app_specific.EnvInputVarPrefix = "P_VAR_"
	app_specific.AppName = "flowpipe"

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeModLoadTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

func (suite *FlowpipeModLoadTestSuite) TestInputStepContainsNotifyBlockThatHasVarOnIt() {
	assert := assert.New(suite.T())

	workspace, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mods/step_with_notify_and_var_default", []string{".hcl", ".sp"})
	assert.Nil(errorAndWarning.Error)

	mod := workspace.Mod

	pipeline := mod.ResourceMaps.Pipelines["local.pipeline.approval_with_variables"]

	if pipeline == nil {
		assert.Fail("Pipeline approval_with_variables not found")
		return
	}
	step, ok := pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Step is not an input step")
		return
	}
	assert.Equal("input", step.Name)

	if step.Notify == nil {
		assert.Fail("notify block is nil")
		return
	}

	assert.Equal("bar", *step.Notify.Channel, "this value - bar - is set from the default of the variable")
	assert.Equal("this value is from pvar file", step.Notify.Integration.AsValueMap()["token"].AsString())
}

func (suite *FlowpipeModLoadTestSuite) TestNotifyDependsAnotherStep() {
	assert := assert.New(suite.T())

	workspace, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mods/notify_depends_another_step", []string{".hcl", ".sp"})
	assert.Nil(errorAndWarning.Error)

	mod := workspace.Mod
	pipeline := mod.ResourceMaps.Pipelines["local.pipeline.approval_with_depends_on"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.Equal("pipeline.get_integration", pipeline.Steps[1].GetDependsOn()[0], "the second step (input step) has a dependency to pipeline.get_integration step")
}

func (suite *FlowpipeModLoadTestSuite) TestNotifyWithRuntimeParam() {
	assert := assert.New(suite.T())

	workspace, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mods/notify_with_runtime_param", []string{".hcl", ".sp"})
	assert.Nil(errorAndWarning.Error)

	mod := workspace.Mod
	pipeline := mod.ResourceMaps.Pipelines["local.pipeline.notify_with_runtime_param"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.Equal("pipeline.get_integration", pipeline.Steps[1].GetDependsOn()[0], "the second step (input step) has a dependency to pipeline.get_integration step")

	unresolvedBodies := pipeline.Steps[1].GetUnresolvedBodies()
	assert.NotNil(unresolvedBodies["notify"])
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeModLoadTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeModLoadTestSuite))
}
