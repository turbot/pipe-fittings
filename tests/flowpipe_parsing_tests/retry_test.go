package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestRetry(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/retry.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.retry_simple"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	assert.NotNil(pipeline.Steps[0].GetRetryConfig())
	assert.Equal(2, pipeline.Steps[0].GetRetryConfig().MaxAttempts)
	assert.Equal("exponential", pipeline.Steps[0].GetRetryConfig().Strategy)

	pipeline = pipelines["local.pipeline.retry_with_if"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	assert.Nil(pipeline.Steps[0].GetRetryConfig())
	assert.NotNil(pipeline.Steps[0].GetUnresolvedBodies()["retry"])

	pipeline = pipelines["local.pipeline.retry_default"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	assert.NotNil(pipeline.Steps[0].GetRetryConfig())
	assert.Equal(3, pipeline.Steps[0].GetRetryConfig().MaxAttempts)
	assert.Equal(1000, pipeline.Steps[0].GetRetryConfig().MinInterval)
	assert.Equal(1000, pipeline.Steps[0].GetRetryConfig().MaxInterval)
	assert.Equal("constant", pipeline.Steps[0].GetRetryConfig().Strategy)

}
