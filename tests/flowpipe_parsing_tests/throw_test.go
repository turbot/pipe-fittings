package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestThrow(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/throw.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.throw_simple_no_unresolved"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.Equal(1, len(pipeline.Steps[0].GetThrowConfig()))
	assert.False(pipeline.Steps[0].GetThrowConfig()[0].Unresolved)
	assert.Equal("foo", *pipeline.Steps[0].GetThrowConfig()[0].Message)
	assert.True(pipeline.Steps[0].GetThrowConfig()[0].If)

	pipeline = pipelines["local.pipeline.throw_simple_unresolved"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.Equal(1, len(pipeline.Steps[0].GetThrowConfig()))
	assert.True(pipeline.Steps[0].GetThrowConfig()[0].Unresolved)

	pipeline = pipelines["local.pipeline.throw_multiple"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.Equal(4, len(pipeline.Steps[0].GetThrowConfig()))
	assert.True(pipeline.Steps[0].GetThrowConfig()[0].Unresolved)
	assert.False(pipeline.Steps[0].GetThrowConfig()[1].Unresolved)
	assert.True(pipeline.Steps[0].GetThrowConfig()[2].Unresolved)
	assert.False(pipeline.Steps[0].GetThrowConfig()[3].Unresolved)
}
