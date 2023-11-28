package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestLoop(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/loop.fp")
	assert.Nil(err, "error found")

	if pipelines["local.pipeline.simple_loop"] == nil {
		assert.Fail("simple_loop pipeline not found")
		return
	}

	// we should have one unresolved body for the loop
	pipeline := pipelines["local.pipeline.simple_loop"]
	assert.NotNil(pipeline.Steps[0].GetUnresolvedBodies()["loop"])

	pipeline = pipelines["local.pipeline.loop_depeneds_on_another_step"]

	if pipeline == nil {
		assert.Fail("loop_depeneds_on_another_step not found")
		return
	}

	// the second step (the one that has the loop) depends on the first one
	assert.Equal("transform.base", pipeline.Steps[1].GetDependsOn()[0])

	pipeline = pipelines["local.pipeline.simple_http_loop"]
	assert.NotNil(pipeline.Steps[0].GetUnresolvedBodies()["loop"])

}
