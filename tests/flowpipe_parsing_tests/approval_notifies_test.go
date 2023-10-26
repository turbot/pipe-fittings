package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func TestApprovalNotifies(t *testing.T) {
	assert := assert.New(t)

	mod, err := misc.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/approval_notifies.fp")
	assert.Nil(err)
	assert.NotNil(mod)
	if mod == nil {
		return
	}

	assert.Equal(2, len(mod.ResourceMaps.Integrations))

	integration := mod.ResourceMaps.Integrations["local.integration.slack.my_slack_app"]
	if integration == nil {
		assert.Fail("Integration not found")
		return
	}

	pipeline := mod.ResourceMaps.Pipelines["local.pipeline.approval_with_notifies"]
	if pipeline == nil {
		assert.Fail("Pipeline not found")
		return
	}

	inputs, _ := pipeline.Steps[0].GetInputs(nil)
	assert.Nil(inputs)

	inputStep, ok := pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Pipeline step not found")
		return
	}

	assert.Equal("input", inputStep.Name)
	assert.Nil(inputStep.Notify)

	notifies := inputStep.Notifies

	if notifies == cty.NilVal {
		assert.Fail("Notifies is nil")
		return
	}

	notifyList := notifies.AsValueSlice()
	assert.Equal(2, len(notifyList))
	assert.Equal("foo", notifyList[0].AsValueMap()["channel"].AsString())
	assert.Equal("xoxp-111111", notifyList[0].AsValueMap()["integration"].AsValueMap()["token"].AsString())

	assert.Equal("bob.loblaw@bobloblawlaw.com", notifyList[1].AsValueMap()["to"].AsString())

	pipeline = mod.ResourceMaps.Pipelines["local.pipeline.approval_with_notifies_depend_another_step"]
	if pipeline == nil {
		assert.Fail("approval_with_notifies_depend_another_step pipeline not found")
		return
	}

	assert.Equal(2, len(pipeline.Steps))

	assert.Equal("echo", pipeline.Steps[0].GetName())
	assert.Equal("input", pipeline.Steps[1].GetName())

	assert.Equal("echo.echo", pipeline.Steps[1].GetDependsOn()[0])
	unresolvedAttribute := pipeline.Steps[1].GetUnresolvedAttributes()["notifies"]
	assert.NotNil(unresolvedAttribute)
	assert.True(len(unresolvedAttribute.Variables()) > 0)
}
