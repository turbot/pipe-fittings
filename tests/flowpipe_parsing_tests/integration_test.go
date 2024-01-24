package pipeline_test

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestIntegrations(t *testing.T) {
	assert := assert.New(t)

	mod, err := load_mod.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/integration.fp")
	assert.Nil(err)
	assert.NotNil(mod)

	integrations := mod.ResourceMaps.Integrations

	if len(integrations) == 0 {
		assert.Fail("unable to load integrations")
	}

	integration := integrations["local.integration.slack.slack_with_token"]
	if integration == nil {
		assert.Fail("integration local.integration.slack.slack_with_token not found")
	}
	assert.Equal("slack", integration.GetIntegrationType())
	assert.Equal("xoxp-abc123", *integration.(*modconfig.SlackIntegration).Token)
	assert.Nil(integration.(*modconfig.SlackIntegration).SigningSecret)
	assert.Nil(integration.(*modconfig.SlackIntegration).WebhookUrl)
	if _, ok := integration.(*modconfig.EmailIntegration); ok {
		assert.Fail("integration local.integration.slack.slack_with_token matched email integration type")
	}

	integration = integrations["local.integration.slack.slack_with_token_and_secret"]
	if integration == nil {
		assert.Fail("integration local.integration.slack.slack_with_token_and_secret not found")
	}
	assert.Equal("slack", integration.GetIntegrationType())
	assert.Equal("xoxp-abc123", *integration.(*modconfig.SlackIntegration).Token)
	assert.Equal("W&EYrf78rqwf", *integration.(*modconfig.SlackIntegration).SigningSecret)
	assert.Nil(integration.(*modconfig.SlackIntegration).WebhookUrl)
	if _, ok := integration.(*modconfig.EmailIntegration); ok {
		assert.Fail("integration local.integration.slack.slack_with_token_and_secret matched email integration type")
	}

	integration = integrations["local.integration.slack.slack_with_webhook"]
	if integration == nil {
		assert.Fail("integration local.integration.slack.slack_with_webhook not found")
	}
	assert.Equal("slack", integration.GetIntegrationType())
	assert.Nil(integration.(*modconfig.SlackIntegration).Token)
	assert.Nil(integration.(*modconfig.SlackIntegration).SigningSecret)
	assert.Contains(*integration.(*modconfig.SlackIntegration).WebhookUrl, "hooks.slack.com")
	if _, ok := integration.(*modconfig.EmailIntegration); ok {
		assert.Fail("integration local.integration.slack.slack_with_webhook matched email integration type")
	}

	integration = integrations["local.integration.email.email_min"]
	if integration == nil {
		assert.Fail("integration local.integration.email.email_min not found")
	}
	assert.Equal("email", integration.GetIntegrationType())
	assert.Equal("smtp.host.tld", *integration.(*modconfig.EmailIntegration).SmtpHost)
	assert.Nil(integration.(*modconfig.EmailIntegration).SmtpPort)
	assert.Nil(integration.(*modconfig.EmailIntegration).SmtpsPort)
	assert.Nil(integration.(*modconfig.EmailIntegration).SmtpTls)
	assert.Nil(integration.(*modconfig.EmailIntegration).SmtpUsername)
	assert.Nil(integration.(*modconfig.EmailIntegration).SmtpPassword)
	assert.Equal("turbie@flowpipe.io", *integration.(*modconfig.EmailIntegration).From)
	assert.Nil(integration.(*modconfig.EmailIntegration).DefaultRecipient)
	assert.Nil(integration.(*modconfig.EmailIntegration).DefaultSubject)
	if _, ok := integration.(*modconfig.SlackIntegration); ok {
		assert.Fail("integration local.integration.email.email_min matched slack integration type")
	}

	integration = integrations["local.integration.email.email_with_all"]
	if integration == nil {
		assert.Fail("integration local.integration.email.email_with_all not found")
	}
	assert.Equal("email", integration.GetIntegrationType())
	assert.Equal("123.456.789.000", *integration.(*modconfig.EmailIntegration).SmtpHost)
	assert.Equal(25, *integration.(*modconfig.EmailIntegration).SmtpPort)
	assert.Equal(587, *integration.(*modconfig.EmailIntegration).SmtpsPort)
	assert.Equal("auto", *integration.(*modconfig.EmailIntegration).SmtpTls)
	assert.Equal("turbie", *integration.(*modconfig.EmailIntegration).SmtpUsername)
	assert.Equal("some_password_here", *integration.(*modconfig.EmailIntegration).SmtpPassword)
	assert.Equal("turbie@flowpipe.io", *integration.(*modconfig.EmailIntegration).From)
	assert.Equal("user@test.tld", *integration.(*modconfig.EmailIntegration).DefaultRecipient)
	assert.Equal("Flowpipe: Action Required", *integration.(*modconfig.EmailIntegration).DefaultSubject)
	if _, ok := integration.(*modconfig.SlackIntegration); ok {
		assert.Fail("integration local.integration.email.email_with_all matched slack integration type")
	}
}
