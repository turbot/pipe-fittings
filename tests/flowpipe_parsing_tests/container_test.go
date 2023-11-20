package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestContainerStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/container.fp")
	assert.Nil(err, "error found")
	assert.Equal(2, len(pipelines), "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_step_container"] == nil {
		assert.Fail("pipeline_step_container pipeline not found")
		return
	}

	step := pipelines["local.pipeline.pipeline_step_container"].GetStep("container.container_test1")
	if step == nil {
		assert.Fail("container step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("container_test1", inputs[schema.AttributeTypeName])
	assert.Equal("test/image", inputs[schema.AttributeTypeImage])
	assert.Equal(int64(60), inputs[schema.AttributeTypeTimeout])

	if _, ok := inputs[schema.AttributeTypeCmd].([]string); !ok {
		assert.Fail("attribute cmd should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeCmd].([]string)))

	if _, ok := inputs[schema.AttributeTypeEnv].(map[string]string); !ok {
		assert.Fail("env block is not defined correctly")
	}
	env := inputs[schema.AttributeTypeEnv].(map[string]string)
	assert.Equal("hello world", env["ENV_TEST"])

	// Pipeline 2

	if pipelines["local.pipeline.pipeline_step_with_param"] == nil {
		assert.Fail("pipeline_step_with_param pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_step_with_param"].GetStep("container.container_test1")
	if step == nil {
		assert.Fail("container step not found")
		return
	}

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"region":  cty.StringVal("ap-south-1"),
		"image":   cty.StringVal("test/image"),
		"timeout": cty.NumberIntVal(120),
		"cmd": cty.ListVal([]cty.Value{
			cty.StringVal("foo"),
			cty.StringVal("bar"),
		}),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("container_test1", inputs[schema.AttributeTypeName])
	assert.Equal("test/image", inputs[schema.AttributeTypeImage])
	assert.Equal(int64(120), inputs[schema.AttributeTypeTimeout])

	if _, ok := inputs[schema.AttributeTypeCmd].([]string); !ok {
		assert.Fail("attribute cmd should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeCmd].([]string)))

	if _, ok := inputs[schema.AttributeTypeEnv].(map[string]string); !ok {
		assert.Fail("env block is not defined correctly")
	}
	env = inputs[schema.AttributeTypeEnv].(map[string]string)
	assert.Equal("ap-south-1", env["REGION"])
}
