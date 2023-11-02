//nolint:forbidigo // Test case: it's OK to use fmt.Println()
package flowpipe_invalid_tests

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/load_mod"
	"testing"

	"github.com/stretchr/testify/assert"
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
		containsError: `invalid depends_on 'echo.does_not_exist' - does not exist for pipeline local.pipeline`,
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
		title:         "invalid step attribute (echo)",
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
		containsError: "Failed to decode mod:\nUnable to convert port into integer\n(pipelines/invalid_email_port.fp:9,5-30)",
	},
	{
		title:         "invalid email recipient",
		file:          "./pipelines/invalid_email_recipient.fp",
		containsError: "Unable to parse to attribute to string slice: Bad Request: expected string type, but got number\n(pipelines/invalid_email_recipient.fp:5,5-82)",
	},
	{
		title:         "invalid trigger",
		file:          "./pipelines/invalid_trigger.fp",
		containsError: "Failed to decode mod:\nMissing required argument: The argument \"pipeline\" is required, but no definition was found.",
	},
	{
		title:         "invalid approval - notify and notifies specified",
		file:          "./pipelines/approval_notify_and_notifies.fp",
		containsError: "Notify and Notifies attributes are mutualy exclusive: input.input",
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
		title:         "invalid loop - bad definition for echo step loop",
		file:          "./pipelines/loop_invalid_echo.fp",
		containsError: "An argument named \"baz\" is not expected here",
	},
	{
		title:         "invalid loop - no if",
		file:          "./pipelines/loop_no_if.fp",
		containsError: "The argument \"if\" is required, but no definition was found",
	},
}

// Simple invalid test. Only single file resources can be evaluated here. This test is unaable to test
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
		assert.NotNil(err)
		assert.Contains(err.Error(), test.containsError)
	}
}
