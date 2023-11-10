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

	assert.GreaterOrEqual(len(pipelines), 2, "wrong number of pipelines")

	pipeline := pipelines["local.pipeline.retry_simple"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	assert.NotNil(pipeline.Steps[0].GetRetryConfig())
	assert.Equal(3, pipeline.Steps[0].GetRetryConfig().Retries)

	pipeline = pipelines["local.pipeline.retry_with_if"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	assert.Nil(pipeline.Steps[0].GetRetryConfig())
	assert.NotNil(pipeline.Steps[0].GetUnresolvedBodies()["retry"])
}
