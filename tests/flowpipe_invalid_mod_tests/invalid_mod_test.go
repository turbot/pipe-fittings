//nolint:forbidigo // Test case, it's OK to use fmt.Println()
package invalid_mod_tests

import (
	"context"
	"errors"
	"fmt"
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
	title         string
	modDir        string
	containsError string
	errorType     string
}

var tests = []testSetup{
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
		containsError: "invalid depends_on 'echozzzz.bar' - step 'transform.baz' does not exist for pipeline pipeline_with_references.pipeline.foo",
	},
	{
		title:         "Bad step reference 2",
		modDir:        "./mods/bad_step_reference_two",
		containsError: "invalid depends_on 'transform.barrs' - step 'transform.baz' does not exist for pipeline pipeline_with_references.pipeline.foo",
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
		containsError: "invalid depends_on 'aws.abc' - credential does not exist for pipeline mod_with_creds.pipeline.with_creds",
	},
	{
		title:         "Invalid credential type reference - dynamic",
		modDir:        "./mods/invalid_cred_types_dynamic",
		containsError: "invalid depends_on 'foo.<dynamic>' - credential type 'foo' not supported for pipeline invalid_cred_types_dynamic.pipeline.with_invalid_cred_type_dynamic",
	},
	{
		title:         "Invalid credential type reference - static",
		modDir:        "./mods/invalid_cred_types_static",
		containsError: "invalid depends_on 'foo.default' - credential does not exist for pipeline invalid_cred_types_static.pipeline.with_invalid_cred_type_static",
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

		_, errorAndWarning := workspace.Load(suite.ctx, test.modDir, workspace.WithCredentials(map[string]credential.Credential{}))
		assert.NotNil(errorAndWarning.Error)
		if errorAndWarning.Error != nil {
			assert.Contains(errorAndWarning.Error.Error(), test.containsError)
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
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeInvalidTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeSimpleInvalidModTestSuite))
}
