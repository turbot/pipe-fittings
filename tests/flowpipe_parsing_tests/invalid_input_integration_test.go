package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestNotifiesInvalidIntegration(t *testing.T) {
	assert := assert.New(t)

	_, err := load_mod.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/invalid_input_integration.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), "MISSING: integration.slack.test_app")
}
