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
	assert.NotNil(pipeline.Steps[0].GetRetryConfig(nil))
	retryConfig, diags := pipeline.Steps[0].GetRetryConfig(nil)
	assert.Equal(0, len(diags))
	assert.Equal(2, retryConfig.MaxAttempts)
	assert.Equal("exponential", retryConfig.Strategy)

	pipeline = pipelines["local.pipeline.retry_with_if"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	assert.Nil(pipeline.Steps[0].GetRetryConfig(nil))
	assert.NotNil(pipeline.Steps[0].GetUnresolvedBodies()["retry"])

	pipeline = pipelines["local.pipeline.retry_default"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	retryConfig, diags = pipeline.Steps[0].GetRetryConfig(nil)
	assert.Equal(0, len(diags))
	assert.NotNil(retryConfig)
	assert.Equal(3, retryConfig.MaxAttempts)
	assert.Equal(1000, retryConfig.MinInterval)
	assert.Equal(10000, retryConfig.MaxInterval)
	assert.Equal("constant", retryConfig.Strategy)

}
