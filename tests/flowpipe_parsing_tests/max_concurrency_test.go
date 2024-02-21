package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestMaxConcurrency(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/max_concurrency.fp")
	assert.Nil(err, "error found")

	if pipelines["local.pipeline.step_with_max_concurrency"] == nil {
		assert.Fail("step_with_max_concurrency pipeline not found")
		return
	}

	pipeline := pipelines["local.pipeline.step_with_max_concurrency"]
	assert.NotNil(pipeline)
	assert.Equal(15, *pipeline.Steps[0].GetMaxConcurrency(), "max concurrency not set")
	assert.Nil(pipeline.Steps[1].GetMaxConcurrency())

	pipeline = pipelines["local.pipeline.pipeline_with_max_concurrency"]
	assert.NotNil(pipeline)
}
