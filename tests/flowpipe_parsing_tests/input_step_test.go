package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestInputStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./pipelines/input_step.fp")
	assert.Nil(err, "error found")
	assert.GreaterOrEqual(len(pipelines), 2, "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_input"] == nil {
		assert.Fail("parent pipeline not found")
		return
	}

	pipelineDefn := pipelines["local.pipeline.pipeline_with_input"]
	assert.Equal("local.pipeline.pipeline_with_input", pipelineDefn.Name(), "wrong pipeline name")
	assert.Equal(1, len(pipelineDefn.Steps), "wrong number of steps")
	assert.Equal("input", pipelineDefn.Steps[0].GetName())

	steps := pipelineDefn.Steps
	assert.GreaterOrEqual(len(steps), 1, "wrong number of steps")

	inputs, err := steps[0].GetInputs(nil)
	assert.Nil(err)
	assert.NotNil(inputs)

	assert.Equal("Choose an option:", inputs[schema.AttributeTypePrompt])

	notifies := inputs[schema.AttributeTypeNotifies].([]map[string]interface{})
	assert.Equal(1, len(notifies))

	channel := notifies[0]["channel"].(string)
	assert.Equal("#general", channel)

	integration := notifies[0][schema.AttributeTypeIntegration].(map[string]interface{})
	assert.Equal("slack", integration[schema.AttributeTypeType])
	assert.Equal("abcde", integration[schema.AttributeTypeToken])
}

func TestInputStepUnresolvedNotify(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./pipelines/input_step.fp")
	assert.Nil(err, "error found")
	assert.GreaterOrEqual(len(pipelines), 2, "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_unresolved_notify"] == nil {
		assert.Fail("parent pipeline not found")
		return
	}

	pipelineDefn := pipelines["local.pipeline.pipeline_with_unresolved_notify"]
	assert.Equal("local.pipeline.pipeline_with_unresolved_notify", pipelineDefn.Name(), "wrong pipeline name")
	assert.Equal(1, len(pipelineDefn.Steps), "wrong number of steps")
	assert.Equal("input", pipelineDefn.Steps[0].GetName())

	steps := pipelineDefn.Steps
	assert.GreaterOrEqual(len(steps), 1, "wrong number of steps")

	integrationVal := cty.ObjectVal(map[string]cty.Value{
		"slack": cty.ObjectVal(map[string]cty.Value{
			"integrated_app": cty.ObjectVal(map[string]cty.Value{
				"token": cty.StringVal("abcde"),
				"type":  cty.StringVal("slack"),
			}),
		}),
	})

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"channel": cty.StringVal("#general"),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["integration"] = integrationVal
	evalContext.Variables["param"] = paramVal

	inputs, err := steps[0].GetInputs(evalContext)
	assert.Nil(err)
	assert.NotNil(inputs)

	assert.Equal("Choose an option:", inputs[schema.AttributeTypePrompt])

	notifies := inputs[schema.AttributeTypeNotifies].([]map[string]interface{})
	assert.Equal(1, len(notifies))

	channel := notifies[0]["channel"].(string)
	assert.Equal("#general", channel)

	integration := notifies[0][schema.AttributeTypeIntegration].(map[string]interface{})
	assert.Equal("slack", integration[schema.AttributeTypeType])
	assert.Equal("abcde", integration[schema.AttributeTypeToken])
}

func TestEmailInputStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./pipelines/input_step.fp")
	assert.Nil(err, "error found")
	assert.GreaterOrEqual(len(pipelines), 2, "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_email_notify"] == nil {
		assert.Fail("parent pipeline not found")
		return
	}

	pipelineDefn := pipelines["local.pipeline.pipeline_with_email_notify"]
	assert.Equal("local.pipeline.pipeline_with_email_notify", pipelineDefn.Name(), "wrong pipeline name")
	assert.Equal(1, len(pipelineDefn.Steps), "wrong number of steps")
	assert.Equal("input", pipelineDefn.Steps[0].GetName())

	steps := pipelineDefn.Steps
	assert.GreaterOrEqual(len(steps), 1, "wrong number of steps")

	integrationVal := cty.ObjectVal(map[string]cty.Value{
		"email": cty.ObjectVal(map[string]cty.Value{
			"email_integration": cty.ObjectVal(map[string]cty.Value{
				"type":          cty.StringVal("email"),
				"smtp_port":     cty.NumberIntVal(587),
				"from":          cty.StringVal("gallant_meitner@example.com"),
				"smtp_host":     cty.StringVal("smtp.gmail.com"),
				"smtp_username": cty.StringVal("awesomebob@blahblah.com"),
				"smtp_password": cty.StringVal("HelloBob@2023"),
			}),
		}),
	})

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"to": cty.StringVal("awesomebob@blahblah.com"),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["integration"] = integrationVal
	evalContext.Variables["param"] = paramVal

	inputs, err := steps[0].GetInputs(evalContext)
	assert.Nil(err)
	assert.NotNil(inputs)

	// assert.Equal("Choose an option:", inputs[schema.AttributeTypePrompt])

	notifies := inputs[schema.AttributeTypeNotifies].([]map[string]interface{})
	assert.Equal(1, len(notifies))

	to := notifies[0]["to"].(string)
	assert.Equal("awesomebob@blahblah.com", to)

	integration := notifies[0][schema.AttributeTypeIntegration].(map[string]interface{})
	assert.Equal("email", integration[schema.AttributeTypeType].(string))
	assert.Equal("smtp.gmail.com", integration[schema.AttributeTypeSmtpHost].(string))
	assert.Equal("gallant_meitner@example.com", integration[schema.AttributeTypeFrom].(string))
	assert.Equal(587, integration[schema.AttributeTypeSmtpPort].(int))
	assert.Equal("awesomebob@blahblah.com", integration[schema.AttributeTypeSmtpUsername].(string))
	assert.Equal("HelloBob@2023", integration[schema.AttributeTypeSmtpPassword].(string))
}
