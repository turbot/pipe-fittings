package pipeline_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/turbot/pipe-fittings/flowpipeconfig"
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
	title       string
	description string
	base        string
	compare     string
	equal       bool
}

var modEqualityTestCases = []modEqualityTestCase{
	// {
	// 	title:   "base_a == base_a",
	// 	base:    "./base_a",
	// 	compare: "./base_a",
	// 	equal:   true,
	// },
	// {
	// 	title:   "base_a != base_b",
	// 	base:    "./base_a",
	// 	compare: "./base_b",
	// 	equal:   false,
	// },
	// {
	// 	title:   "http_step_with_config == http_step_with_config",
	// 	base:    "./http_step_with_config",
	// 	compare: "./http_step_with_config",
	// 	equal:   true,
	// },
	// {
	// 	title:   "http_step_with_config == http_step_with_config_line_change",
	// 	base:    "./http_step_with_config",
	// 	compare: "./http_step_with_config_line_change",
	// 	equal:   true,
	// },
	// {
	// 	title:   "http_step_with_config == http_step_with_config_b",
	// 	base:    "./http_step_with_config",
	// 	compare: "./http_step_with_config_b",
	// 	equal:   false,
	// },
	// {
	// 	title:   "http_step_with_config == http_step_with_config_c",
	// 	base:    "./http_step_with_config",
	// 	compare: "./http_step_with_config_c",
	// 	equal:   false,
	// },
	// {
	// 	title:   "input_step_a == input_step_a",
	// 	base:    "./input_step_a",
	// 	compare: "./input_step_a",
	// 	equal:   true,
	// },
	{
		title:   "input_step_a != input_step_b",
		base:    "./input_step_a",
		compare: "./input_step_b",
		equal:   false,
	},
	// {
	// 	title:   "input_step_b == input_step_b",
	// 	base:    "./input_step_b",
	// 	compare: "./input_step_b",
	// 	equal:   true,
	// },
	// {
	// 	title:   "input_step_a != input_step_c",
	// 	base:    "./input_step_a",
	// 	compare: "./input_step_c",
	// 	equal:   false,
	// },
	// {
	// 	title:   "input_step_c == input_step_c",
	// 	base:    "./input_step_c",
	// 	compare: "./input_step_c",
	// 	equal:   true,
	// },
	// {
	// 	title:   "input_step_d != input_step_d",
	// 	base:    "./input_step_d",
	// 	compare: "./input_step_d",
	// 	equal:   true,
	// },
	// {
	// 	title:   "input_step_d != input_step_d_line_change",
	// 	base:    "./input_step_d",
	// 	compare: "./input_step_d_line_change",
	// 	equal:   true,
	// },
	// {
	// 	title:   "input_step_d != input_step_e",
	// 	base:    "./input_step_d",
	// 	compare: "./input_step_e",
	// 	equal:   false,
	// },
	// {
	// 	title:   "container_a == container_a",
	// 	base:    "./container_a",
	// 	compare: "./container_a",
	// 	equal:   true,
	// },
	// {
	// 	title:   "container_a == container_a_line_change",
	// 	base:    "./container_a",
	// 	compare: "./container_a_line_change",
	// 	equal:   true,
	// },
	// {
	// 	title:   "container_a != container_b",
	// 	base:    "./container_a",
	// 	compare: "./container_b",
	// 	equal:   false,
	// },
	// {
	// 	title:   "container_c == container_c",
	// 	base:    "./container_c",
	// 	compare: "./container_c",
	// 	equal:   true,
	// },
	// {
	// 	title:   "container_c != container_d",
	// 	base:    "./container_c",
	// 	compare: "./container_d",
	// 	equal:   false,
	// },
	// {
	// 	title:   "container_d == container_d",
	// 	base:    "./container_d",
	// 	compare: "./container_d",
	// 	equal:   true,
	// },
	// {
	// 	title:       "container_d != container_e",
	// 	description: "cmd attribute has different map values, runtime reference",
	// 	base:        "./container_d",
	// 	compare:     "./container_e",
	// 	equal:       false,
	// },
	// {
	// 	title:   "container_f == container_f",
	// 	base:    "./container_f",
	// 	compare: "./container_f",
	// 	equal:   true,
	// },
	// {
	// 	title:       "container_f != container_g",
	// 	description: "cmd attribute has different map values, not runtime reference",
	// 	base:        "./container_f",
	// 	compare:     "./container_g",
	// 	equal:       false,
	// },
	// {
	// 	title:   "param_a == param_a",
	// 	base:    "./param_a",
	// 	compare: "./param_a",
	// 	equal:   true,
	// },
	// {
	// 	title:   "param_a == param_a_line_change",
	// 	base:    "./param_a",
	// 	compare: "./param_a_line_change",
	// 	equal:   true,
	// },
	// {
	// 	title:       "param_a != param_b",
	// 	description: "param b has a param with a different default value, same name",
	// 	base:        "./param_a",
	// 	compare:     "./param_b",
	// 	equal:       false,
	// },
	// {
	// 	title:   "param_c == param_c",
	// 	base:    "./param_c",
	// 	compare: "./param_c",
	// 	equal:   true,
	// },
	// {
	// 	title:       "param_c != param_d",
	// 	description: "param d has a param with a different type",
	// 	base:        "./param_c",
	// 	compare:     "./param_d",
	// 	equal:       false,
	// },
	// {
	// 	title:   "foreach_a == foreach_a",
	// 	base:    "./foreach_a",
	// 	compare: "./foreach_a",
	// 	equal:   true,
	// },
	// {
	// 	title:   "foreach_a == foreach_a_line_change",
	// 	base:    "./foreach_a",
	// 	compare: "./foreach_a_line_change",
	// 	equal:   true,
	// },
	// {
	// 	title:       "foreach_a != foreach_b",
	// 	description: "different element in for_each, same length",
	// 	base:        "./foreach_a",
	// 	compare:     "./foreach_b",
	// 	equal:       false,
	// },
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

			flowpipeConfigA, err := flowpipeconfig.LoadFlowpipeConfig([]string{TARGET_DIR})
			if err.Error != nil {
				assert.FailNow(err.Error.Error())
				return
			}

			wA, errorAndWarning := workspace.Load(suite.ctx, TARGET_DIR, workspace.WithCredentials(flowpipeConfigA.Credentials), workspace.WithIntegrations(flowpipeConfigA.Integrations), workspace.WithNotifiers(flowpipeConfigA.Notifiers))
			assert.NotNil(wA)
			assert.Nil(errorAndWarning.Error)
			assert.Equal(0, len(errorAndWarning.Warnings))

			utils.EmptyDir(TARGET_DIR)            //nolint:errcheck // test only
			utils.CopyDir(tc.compare, TARGET_DIR) //nolint:errcheck // test only

			flowpipeConfigB, err := flowpipeconfig.LoadFlowpipeConfig([]string{TARGET_DIR})
			if err.Error != nil {
				assert.FailNow(err.Error.Error())
				return
			}

			wB, errorAndWarning := workspace.Load(suite.ctx, TARGET_DIR, workspace.WithCredentials(flowpipeConfigB.Credentials), workspace.WithIntegrations(flowpipeConfigB.Integrations), workspace.WithNotifiers(flowpipeConfigB.Notifiers))
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
