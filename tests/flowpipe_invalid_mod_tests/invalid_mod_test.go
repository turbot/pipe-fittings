package invalid_mod_tests

import (
	"context"
	"errors"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/tests/test_init"
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

	// set app specific constants
	test_init.SetAppSpecificConstants()

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeSimpleInvalidModTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

type testSetup struct {
	title            string
	modDir           string
	containsError    string
	errorType        string
	expectedContains []string
}

var tests = []testSetup{
	{
		title:  "Missing var",
		modDir: "./mods/mod_missing_var",
		// Testing variable not set, if message is different then the variable prompt may not work
		containsError: "missing 1 variable value:\n\tslack_token not set",
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
		containsError: "invalid depends_on 'echozzzz.bar', step 'echozzzz.bar' does not exist in pipeline pipeline_with_references.pipeline.foo",
	},
	{
		title:         "Bad step reference 2",
		modDir:        "./mods/bad_step_reference_two",
		containsError: "invalid depends_on 'transform.barrs', step 'transform.barrs' does not exist in pipeline pipeline_with_references.pipeline.foo",
	},
	{
		title:         "Bad trigger reference to pipeline",
		modDir:        "./mods/bad_trigger_reference",
		containsError: "Unresolved blocks:\n   trigger.my_hourly_trigger -> pipeline.simple_with_trigger\n     MISSING: pipeline.simple_with_trigger",
		errorType:     perr.ErrorCodeDependencyFailure,
	},
	{
		title:         "Invalid credential reference",
		modDir:        "./mods/invalid_creds_reference",
		containsError: "invalid depends_on 'aws.abc', credential does not exist in pipeline mod_with_creds.pipeline.with_creds",
	},
	{
		title:         "Invalid credential type reference - dynamic",
		modDir:        "./mods/invalid_cred_types_dynamic",
		containsError: "invalid depends_on 'foo.<dynamic>', credential type 'foo' not supported in pipeline invalid_cred_types_dynamic.pipeline.with_invalid_cred_type_dynamic",
	},
	{
		title:         "Invalid credential type reference - static",
		modDir:        "./mods/invalid_cred_types_static",
		containsError: "invalid depends_on 'foo.default', credential does not exist in pipeline invalid_cred_types_static.pipeline.with_invalid_cred_type_static",
		expectedContains: []string{
			"invalid_cred_types_static/mod.fp:",
		},
	},
	{
		title:         "Number as string in retry block",
		modDir:        "./mods/number_as_string_retry_block",
		containsError: "Failed to decode mod: Unable to parse min_interval attribute to integer",
	},
	{
		title:         "Bool as string in error block",
		modDir:        "./mods/bool_as_string_error_block",
		containsError: "Failed to decode mod: Unable to parse ignore attribute to bool",
	},
	{
		title:         "Bool as number in error block",
		modDir:        "./mods/bool_as_number_error_block",
		containsError: "Failed to decode mod: Unable to parse ignore attribute to bool",
	},
	{
		title:         "Input step no label",
		modDir:        "./mods/input_step_no_label",
		containsError: "Missing name for option: All option blocks must have 1 labels (name).",
	},
	{
		title:         "Bad reference to another step",
		modDir:        "./mods/bad_step_reference_from_another_step",
		containsError: "invalid depends_on 'transform.onex', step 'transform.onex' does not exist in pipeline test.pipeline.bad_step_ref",
		expectedContains: []string{
			"bad_step_reference_from_another_step/mod.fp",
		},
	},
	{
		title:  "Bad reference to another step from output block",
		modDir: "./mods/bad_step_reference_from_output",
		expectedContains: []string{
			"invalid depends_on 'input.approve' in output block, 'input.approve' does not exist in pipeline test.pipeline.bad_step_ref",
			"bad_step_reference_from_output/mod.fp:",
		},
	},
}

func (suite *FlowpipeSimpleInvalidModTestSuite) TestSimpleInvalidMods() {

	for _, test := range tests {
		suite.T().Run(test.title, func(t *testing.T) {
			assert := assert.New(t)

			if test.title == "" {
				assert.Fail("Test must have title")
				return
			}
			if test.containsError == "" && len(test.expectedContains) == 0 {
				assert.Fail("Test " + test.title + " does not have expected error")
				return
			}

			_, errorAndWarning := workspace.Load(suite.ctx, test.modDir, workspace.WithCredentials(map[string]credential.Credential{}))
			assert.NotNil(errorAndWarning.Error)
			if errorAndWarning.Error != nil {
				assert.Contains(errorAndWarning.Error.Error(), test.containsError)

				for _, expectedContains := range test.expectedContains {
					assert.Contains(errorAndWarning.Error.Error(), expectedContains)
				}
			}

			if test.errorType != "" {
				var err perr.ErrorModel
				ok := errors.As(errorAndWarning.Error, &err)
				if !ok {
					assert.Fail("should be a pcerr.ErrorModel")
					return
				}

				assert.Equal(test.errorType, err.Type, "wrong error type")
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeInvalidTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeSimpleInvalidModTestSuite))
}
