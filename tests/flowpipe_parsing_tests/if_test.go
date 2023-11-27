package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestIf(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/if.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.if"] == nil {
		assert.Fail("if pipeline not found")
		return
	}

	step := pipelines["local.pipeline.if"].GetStep("transform.text_1")

	if step == nil {
		assert.Fail("transform.text_1 step not found")
		return
	}

	ifExpr := step.GetUnresolvedAttributes()["if"]
	if ifExpr == nil {
		assert.Fail("if expression not found")
		return
	}
}
