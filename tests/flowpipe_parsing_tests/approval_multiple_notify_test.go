package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestApprovalMultipleNotify(t *testing.T) {
	assert := assert.New(t)

	mod, err := load_mod.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/approval_multiple_notify.fp")
	assert.Nil(err)
	assert.NotNil(mod)
	assert.Equal(2, len(mod.ResourceMaps.Integrations))

	// Validate slack type integration
	integration := mod.ResourceMaps.Integrations["local.integration.slack.slack_dummy_app"]
	if integration == nil {
		assert.Fail("Integration not found")
		return
	}
	assert.Equal("local.integration.slack.slack_dummy_app", integration.Name())
	assert.Equal("slack", integration.(*modconfig.SlackIntegration).Type)
	assert.Equal("xoxp-fhshf2395723hkhfskh", *integration.(*modconfig.SlackIntegration).Token)

	// Validate email type integration
	integration = mod.ResourceMaps.Integrations["local.integration.email.email_dummy_app"]
	if integration == nil {
		assert.Fail("Integration not found")
		return
	}
	assert.Equal("local.integration.email.email_dummy_app", integration.Name())
	assert.Equal("email", integration.(*modconfig.EmailIntegration).Type)
	assert.Equal("smtp.gmail.com", *integration.(*modconfig.EmailIntegration).SmtpHost)
	assert.Equal(587, *integration.(*modconfig.EmailIntegration).SmtpPort)
	assert.Equal("test@example.com", *integration.(*modconfig.EmailIntegration).From)

	pipeline := mod.ResourceMaps.Pipelines["local.pipeline.approval_multiple_notify_pipeline"]
	if pipeline == nil {
		assert.Fail("Pipeline not found")
		return
	}

	inputStep, ok := pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Pipeline step not found")
		return
	}
	assert.Equal("input_multiple_notify", inputStep.Name)
	assert.NotNil(inputStep.NotifyList)
	assert.Equal(2, len(inputStep.NotifyList))

	// Validate notify block - 1
	notify := inputStep.NotifyList[0]
	assert.Equal("#general", *notify.Channel)

	integrationLink := notify.Integration
	assert.NotNil(integrationLink)
	integrationMap := integrationLink.AsValueMap()
	assert.NotNil(integrationMap)
	assert.Equal("xoxp-fhshf2395723hkhfskh", integrationMap["token"].AsString())

	// Validate notify block - 2
	notify = inputStep.NotifyList[1]
	assert.Equal("test@example.com", *notify.To)

	integrationLink = notify.Integration
	assert.NotNil(integrationLink)
	integrationMap = integrationLink.AsValueMap()
	assert.NotNil(integrationMap)
	assert.Equal("test@example.com", integrationMap[schema.AttributeTypeFrom].AsString())
	assert.Equal("smtp.gmail.com", integrationMap[schema.AttributeTypeSmtpHost].AsString())

	var port *int
	err = gocty.FromCtyValue(integrationMap[schema.AttributeTypeSmtpPort], &port)
	assert.Nil(err)
	assert.Equal(587, *port)
}
