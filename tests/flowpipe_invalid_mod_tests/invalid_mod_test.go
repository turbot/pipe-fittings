//nolint:forbidigo // Test case, it's OK to use fmt.Println()
package invalid_mod_tests

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeSimpleInvalidModTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeSimpleInvalidModTestSuite) SetupSuite() {

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
func (suite *FlowpipeSimpleInvalidModTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

type testSetup struct {
	title         string
	modDir        string
	containsError string
	errorType     string
}

var tests = []testSetup{
	{
		title:         "Missing var",
		modDir:        "./mods/mod_missing_var",
		containsError: "Unresolved blocks:\n   integration.slack.slack_app_from_var -> var.slack_signing_secret\n     MISSING: var.slack_signing_secret",
	},
	{
		title:         "Missing var trigger",
		modDir:        "./mods/mod_missing_var_trigger",
		containsError: "Unresolved blocks:\n   trigger.my_hourly_trigger -> var.trigger_schedule",
	},
	{
		title:         "Bad step pipeline reference",
		modDir:        "./mods/mod_bad_step_pipeline_reference",
		containsError: "Unresolved blocks:\n   pipeline.foo -> pipeline.foo_two_invalid",
	},
	{
		title:         "Bad step reference",
		modDir:        "./mods/bad_step_reference",
		containsError: "invalid depends_on 'echozzzz.bar' - step 'echo.baz' does not exist for pipeline pipeline_with_references.pipeline.foo",
	},
	{
		title:         "Bad step reference 2",
		modDir:        "./mods/bad_step_reference_two",
		containsError: "invalid depends_on 'echo.barrs' - step 'echo.baz' does not exist for pipeline pipeline_with_references.pipeline.foo",
	},
	{
		title:         "Bad trigger reference to pipeline",
		modDir:        "./mods/bad_trigger_reference",
		containsError: "Unresolved blocks:\n   trigger.my_hourly_trigger -> pipeline.simple_with_trigger\n     MISSING: pipeline.simple_with_trigger",
		errorType:     perr.ErrorCodeDependencyFailure,
	},
}

func (suite *FlowpipeSimpleInvalidModTestSuite) TestSimpleInvalidMods() {
	assert := assert.New(suite.T())

	for _, test := range tests {
		if test.title == "" {
			assert.Fail("Test must have title")
			continue
		}
		if test.containsError == "" {
			assert.Fail("Test " + test.title + " does not have containsError")
			continue
		}

		fmt.Println("Running test " + test.title)

		_, errorAndWarning := workspace.LoadWithParams(suite.ctx, test.modDir, []string{".hcl", ".sp"})
		assert.NotNil(errorAndWarning.Error)
		assert.Contains(errorAndWarning.Error.Error(), test.containsError)

		if test.errorType != "" {
			err, ok := errorAndWarning.Error.(perr.ErrorModel)
			if !ok {
				assert.Fail("should be a pcerr.ErrorModel")
				return
			}

			assert.Equal(test.errorType, err.Type, "wrong error type")
		}
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeInvalidTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeSimpleInvalidModTestSuite))
}
