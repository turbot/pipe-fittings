package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestDemoPipeline(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	pipelines, _, err := misc.LoadPipelines(ctx, "./pipelines/demo.fp")
	assert.Nil(err, "error found")
	assert.NotNil(pipelines)
	assert.NotNil(pipelines["local.pipeline.complex_one"])

	// TODO: check pipeline definition
}
