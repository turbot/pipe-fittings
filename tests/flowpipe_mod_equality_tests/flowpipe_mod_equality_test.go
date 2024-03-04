package pipeline_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/turbot/pipe-fittings/tests/test_init"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlowpipeModEqualityTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeModEqualityTestSuite) SetupSuite() {

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

	// Create a single, global context for the application
	ctx := context.Background()

	suite.ctx = ctx

	// set app specific constants
	test_init.SetAppSpecificConstants()

	suite.SetupSuiteRunCount++
}

type modEqualityTestCase struct {
	title   string
	base    string
	compare string
	equal   bool
}

var modEqualityTestCases = []modEqualityTestCase{
	{
		title:   "test: base_a == base_a",
		base:    "./base_a",
		compare: "./base_a",
		equal:   true,
	},
	{
		title:   "test: base_a != base_b",
		base:    "./base_a",
		compare: "./base_b",
		equal:   false,
	},
	{
		title:   "test: http_step_with_config == http_step_with_config",
		base:    "./http_step_with_config",
		compare: "./http_step_with_config",
		equal:   true,
	},
	{
		title:   "test: http_step_with_config == http_step_with_config_b",
		base:    "./http_step_with_config",
		compare: "./http_step_with_config_b",
		equal:   false,
	},
	{
		title:   "test: http_step_with_config == http_step_with_config_c",
		base:    "./http_step_with_config",
		compare: "./http_step_with_config_c",
		equal:   false,
	},
}

const (
	TARGET_DIR = "./target_dir"
)

func (suite *FlowpipeModEqualityTestSuite) TestFlowpipeModEquality() {

	for _, tc := range modEqualityTestCases {
		suite.T().Run(tc.title, func(t *testing.T) {
			assert := assert.New(t)
			utils.EmptyDir(TARGET_DIR)         //nolint:errcheck // test only
			utils.CopyDir(tc.base, TARGET_DIR) //nolint:errcheck // test only

			wA, errorAndWarning := workspace.Load(suite.ctx, TARGET_DIR)
			assert.NotNil(wA)
			assert.Nil(errorAndWarning.Error)
			assert.Equal(0, len(errorAndWarning.Warnings))

			utils.EmptyDir(TARGET_DIR)            //nolint:errcheck // test only
			utils.CopyDir(tc.compare, TARGET_DIR) //nolint:errcheck // test only

			wB, errorAndWarning := workspace.Load(suite.ctx, TARGET_DIR)
			assert.NotNil(wB)
			assert.Nil(errorAndWarning.Error)
			assert.Equal(0, len(errorAndWarning.Warnings))

			assert.Equal(tc.equal, wA.GetResourceMaps().Equals(wB.GetResourceMaps()))
		})
	}

}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeModEqualityTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeModEqualityTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeModEqualityTestSuite))
}
