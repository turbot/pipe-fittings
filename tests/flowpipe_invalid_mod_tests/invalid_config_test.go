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
	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/tests/test_init"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeSimpleInvalidConfigTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeSimpleInvalidConfigTestSuite) SetupSuite() {

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
func (suite *FlowpipeSimpleInvalidConfigTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

type invalidConfigTestSetup struct {
	title         string
	modDir        string
	configDirs    []string
	containsError string
	errorType     string
}

var invalidConfigTests = []invalidConfigTestSetup{
	{
		title:         "Invalid (unsupported) credential type",
		modDir:        "",
		configDirs:    []string{"./mods/invalid_cred"},
		containsError: "Invalid credential type slacks",
	},
}

func (suite *FlowpipeSimpleInvalidConfigTestSuite) TestSimpleInvalidMods() {
	assert := assert.New(suite.T())

	for _, test := range invalidConfigTests {
		if test.title == "" {
			assert.Fail("Test must have title")
			continue
		}
		if test.containsError == "" {
			assert.Fail("Test " + test.title + " does not have containsError")
			continue
		}

		fmt.Println("Running test " + test.title)

		_, errorAndWarning := flowpipeconfig.LoadFlowpipeConfig(test.configDirs)
		if errorAndWarning.Error == nil {
			assert.FailNow("Expected error")
			return
		}

		assert.Contains(errorAndWarning.Error.Error(), test.containsError)

		if test.modDir != "" {
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
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeInvalidConfigTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeSimpleInvalidConfigTestSuite))
}
