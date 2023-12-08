//nolint:forbidigo // Test case: it's OK to use fmt.Println()
package flowpipe_invalid_tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

type testSetup struct {
	title         string
	file          string
	containsError string
}

var tests = []testSetup{
	{
		title:         "bad output reference",
		file:          "./pipelines/bad_output_reference.fp",
		containsError: `invalid depends_on 'transform.does_not_exist' - does not exist for pipeline local.pipeline`,
	},
	{
		title:         "duplicate pipeline",
		file:          "./pipelines/duplicate_pipelines.fp",
		containsError: "Mod defines more than one resource named 'local.pipeline.pipeline_007'",
	},
	{
		title:         "duplicate triggers - different pipeline",
		file:          "./pipelines/duplicate_triggers_diff_pipeline.fp",
		containsError: "Mod defines more than one resource named 'local.trigger.schedule.my_hourly_trigger'",
	},
	{
		title:         "duplicate triggers",
		file:          "./pipelines/duplicate_triggers.fp",
		containsError: "duplicate unresolved block name 'trigger.my_hourly_trigger'",
	},
	{
		title:         "invalid http trigger",
		file:          "./pipelines/invalid_http_trigger.fp",
		containsError: `Unsupported argument: An argument named "if" is not expected here.`,
	},
	{
		title:         "invalid step attribute (transform)",
		file:          "./pipelines/invalid_step_attribute.fp",
		containsError: `Unsupported argument: An argument named "abc" is not expected here.`,
	},
	{
		title:         "invalid param",
		file:          "./pipelines/invalid_params.fp",
		containsError: `invalid property path: params.message_retention_duration`,
	},
	{
		title:         "invalid depends",
		file:          "./pipelines/invalid_depends.fp",
		containsError: "Failed to decode mod:\ninvalid depends_on 'http.my_step_1' - step 'sleep.sleep_1' does not exist for pipeline local.pipeline.invalid_depends",
	},
	{
		title:         "invalid email port",
		file:          "./pipelines/invalid_email_port.fp",
		containsError: "Failed to decode mod:\nUnable to convert port into integer\n",
	},
	{
		title:         "invalid email recipient",
		file:          "./pipelines/invalid_email_recipient.fp",
		containsError: "Unable to parse to attribute to string slice: Bad Request: expected string type, but got number\n",
	},
	{
		title:         "invalid trigger",
		file:          "./pipelines/invalid_trigger.fp",
		containsError: "Failed to decode mod:\nMissing required argument: The argument \"pipeline\" is required, but no definition was found.",
	},
	{
		title:         "invalid approval - notify and notifies specified",
		file:          "./pipelines/approval_notify_and_notifies.fp",
		containsError: "Notify and Notifies attributes are mutually exclusive: input.input",
	},
	{
		title:         "invalid approval - slack notify missing channel",
		file:          "./pipelines/approval_invalid_notify_slack.fp",
		containsError: "channel must be specified for slack integration",
	},
	{
		title:         "invalid approval - email notify missing to",
		file:          "./pipelines/approval_invalid_notify_email.fp",
		containsError: "to must be specified for email integration",
	},
	{
		title:         "invalid approval - slack integration invalid attribute",
		file:          "./pipelines/approval_invalid_integration_slack_attribute.fp",
		containsError: "Unsupported attribute: 'from' not expected here.",
	},
	{
		title:         "invalid approval - email integration invalid attribute",
		file:          "./pipelines/approval_invalid_integration_email_attribute.fp",
		containsError: "Unsupported attribute: 'token' not expected here.",
	},
	{
		title:         "invalid approval - notify with missing integration attribute",
		file:          "./pipelines/approval_invalid_notify_missing_integration.fp",
		containsError: "Missing required argument: The argument \"integration\" is required, but no definition was found.",
	},
	{
		title:         "invalid approval - notify with invalid integration that does not exist",
		file:          "./pipelines/approval_invalid_notify_invalid_integration.fp",
		containsError: "MISSING: integration.slack.missing_slack_integration",
	},
	{
		title:         "invalid approval - step with multiple notify block with invalid slack attribute",
		file:          "./pipelines/approval_invalid_multiple_notify_slack.fp",
		containsError: "channel must be specified for slack integration",
	},
	{
		title:         "invalid approval - step with multiple notify block with invalid email attribute",
		file:          "./pipelines/approval_invalid_multiple_notify_email.fp",
		containsError: "to must be specified for email integration",
	},
	{
		title:         "invalid approval - multiple notify and notifies specified",
		file:          "./pipelines/approval_multiple_notify_and_notifies.fp",
		containsError: "Notify and Notifies attributes are mutually exclusive: input.input",
	},
	{
		title:         "invalid loop - bad definition for echo step loop",
		file:          "./pipelines/loop_invalid_transform.fp",
		containsError: "An argument named \"baz\" is not expected here",
	},
	{
		title:         "invalid loop - no if",
		file:          "./pipelines/loop_no_if.fp",
		containsError: "The argument \"until\" is required, but no definition was found",
	},
	{
		title:         "retry - multiple retry blocks",
		file:          "./pipelines/retry_multiple_retry_blocks.fp",
		containsError: "Only one retry block is allowed per step",
	},
	{
		title:         "retry - invalid attribute",
		file:          "./pipelines/retry_invalid_attribute.fp",
		containsError: "Unsupported attribute except in retry block",
	},
	{
		title:         "retry - invalid attribute value",
		file:          "./pipelines/retry_invalid_attribute_value.fp",
		containsError: "Unsuitable value: a number is required",
	},
	{
		title:         "retry - invalid attribute value for strategy",
		file:          "./pipelines/retry_invalid_value_for_strategy.fp",
		containsError: "Invalid retry strategy: Valid values are constant, exponential or linear",
	},
	{
		title:         "throw - invalid attribute",
		file:          "./pipelines/throw_invalid_attribute.fp",
		containsError: "An argument named \"foo\" is not expected here",
	},
	{
		title:         "throw - missing if",
		file:          "./pipelines/throw_missing_if.fp",
		containsError: "The argument \"if\" is required, but no definition was found",
	},
	{
		title:         "invalid container step attribute - source",
		file:          "./pipelines/container_step_invalid_attribute.fp",
		containsError: "Source is not yet implemented",
	},
	{
		title:         "invalid pipeline output attribute - sensitive",
		file:          "./pipelines/output_invalid_attribute.fp",
		containsError: "Unsupported argument: An argument named \"sensitive\" is not expected here.",
	},
	{
		title:         "invalid error block attribute - ignored",
		file:          "./pipelines/invalid_error_attribute.fp",
		containsError: "Unsupported attribute 'ignored' provided for block type error",
	},
}

// Simple invalid test. Only single file resources can be evaluated here. This test is unable to test
// more complex error message expectations or complex structure such as mod & var
func TestSimpleInvalidResources(t *testing.T) {
	assert := assert.New(t)

	ctx := context.TODO()

	for _, test := range tests {
		if test.title == "" {
			assert.Fail("Test must contain a title")
			continue
		}
		if test.containsError == "" {
			assert.Fail("Test: " + test.title + " does not have containsError test")
			continue
		}

		fmt.Println("Running test: " + test.title)

		_, _, err := load_mod.LoadPipelines(ctx, test.file)
		if err == nil {
			assert.Fail("Test: " + test.title + " did not return an error")
			continue
		}
		assert.Contains(err.Error(), test.containsError)
	}
}
