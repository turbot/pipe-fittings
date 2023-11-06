package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestTransformStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./pipelines/transform.fp")
	assert.Nil(err, "error found")
	assert.Equal(1, len(pipelines), "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_transform_step"] == nil {
		assert.Fail("pipeline_with_transform_step pipeline not found")
		return
	}

	step := pipelines["local.pipeline.pipeline_with_transform_step"].GetStep("transform.transform_test")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("hello world", inputs["value"])
}
