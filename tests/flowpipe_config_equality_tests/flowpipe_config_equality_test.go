package pipeline_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/tests/test_init"
	"github.com/turbot/pipe-fittings/utils"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlowpipeConfigEqualityTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeConfigEqualityTestSuite) SetupSuite() {

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

func (suite *FlowpipeConfigEqualityTestSuite) TestFlowpipeConfigEquality() {
	assert := assert.New(suite.T())

	utils.EmptyDir("./config_notifier_target")                          //nolint:errcheck // test only
	utils.CopyDir("./config_notifier_base", "./config_notifier_target") //nolint:errcheck // test only

	flowpipeConfigA, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	flowpipeConfigB, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	// First test - A and B reading the same file .. should return true
	assert.True(flowpipeConfigA.Equals(flowpipeConfigB))

	// Second test - add integration (in config_notifier_base_b)

	utils.EmptyDir("./config_notifier_target")                            //nolint:errcheck // test only
	utils.CopyDir("./config_notifier_base_b", "./config_notifier_target") //nolint:errcheck // test only

	flowpipeConfigB, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	assert.False(flowpipeConfigA.Equals(flowpipeConfigB))

	// Third test - reset A and the equality should be true
	flowpipeConfigA, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)
	assert.True(flowpipeConfigA.Equals(flowpipeConfigB))

	// Fourth test - change the notifier to add a new notify block (that's coming from base_c)
	// the euqality should be false
	utils.EmptyDir("./config_notifier_target")                            //nolint:errcheck // test only
	utils.CopyDir("./config_notifier_base_c", "./config_notifier_target") //nolint:errcheck // test only

	flowpipeConfigB, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	assert.False(flowpipeConfigA.Equals(flowpipeConfigB))

	// Fifth test - reset A and the equality should be true
	flowpipeConfigA, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)
	assert.True(flowpipeConfigA.Equals(flowpipeConfigB))

	// Sixth test - update one of the notify block within the notifier
	utils.EmptyDir("./config_notifier_target")                            //nolint:errcheck // test only
	utils.CopyDir("./config_notifier_base_d", "./config_notifier_target") //nolint:errcheck // test only

	flowpipeConfigB, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	assert.False(flowpipeConfigA.Equals(flowpipeConfigB))
}

func (suite *FlowpipeConfigEqualityTestSuite) TestFlowpipeConfigEqualityTwo() {
	assert := assert.New(suite.T())

	utils.EmptyDir("./config_notifier_target")                            //nolint:errcheck // test only
	utils.CopyDir("./config_notifier_base_e", "./config_notifier_target") //nolint:errcheck // test only

	flowpipeConfigA, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	flowpipeConfigB, err := flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	// First test - A and B reading the same file .. should return true
	assert.True(flowpipeConfigA.Equals(flowpipeConfigB))

	// Second test, update notifier's to list
	utils.EmptyDir("./config_notifier_target")                            //nolint:errcheck // test only
	utils.CopyDir("./config_notifier_base_f", "./config_notifier_target") //nolint:errcheck // test only

	flowpipeConfigB, err = flowpipeconfig.LoadFlowpipeConfig([]string{"./config_notifier_target"})
	assert.Nil(err.Error)

	assert.False(flowpipeConfigA.Equals(flowpipeConfigB))

}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeConfigEqualityTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeConfigEqualityTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeConfigEqualityTestSuite))
}
