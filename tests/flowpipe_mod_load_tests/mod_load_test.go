package flowpipe_mod_load_tests

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/filepaths"
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

	filepaths.PipesComponentWorkspaceDataDir = ".flowpipe"
	filepaths.PipesComponentModsFileName = "mod.hcl"
	filepaths.PipesComponentDefaultVarsFileName = "flowpipe.pvars"
	filepaths.PipesComponentDefaultInstallDir = "~/.flowpipe"

	constants.PipesComponentModDataExtension = ".hcl"
	constants.PipesComponentVariablesExtension = ".pvars"
	constants.PipesComponentAutoVariablesExtension = ".auto.pvars"
	constants.PipesComponentEnvInputVarPrefix = "P_VAR_"
	constants.PipesComponentAppName = "flowpipe"

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

	if step.NotifyList == nil {
		assert.Fail("notify block is nil")
		return
	}
	assert.Equal(1, len(step.NotifyList))

	assert.Equal("bar", *step.NotifyList[0].Channel, "this value - bar - is set from the default of the variable")
	assert.Equal("this value is from pvar file", step.NotifyList[0].Integration.AsValueMap()["token"].AsString())
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
