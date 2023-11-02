package pipeline_test

import (
	"context"
	"github.com/turbot/pipe-fittings/load_mod"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/input_step.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_input"] == nil {
		assert.Fail("parent pipeline not found")
		return
	}

	pipelineDefn := pipelines["local.pipeline.pipeline_with_input"]
	assert.Equal("local.pipeline.pipeline_with_input", pipelineDefn.Name(), "wrong pipeline name")
	assert.Equal(1, len(pipelineDefn.Steps), "wrong number of steps")
	assert.Equal("input", pipelineDefn.Steps[0].GetName())
}
