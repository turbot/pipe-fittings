package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestInputStepInvalidType(t *testing.T) {
	assert := assert.New(t)

	_, err := load_mod.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/invalid_input_type.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), "Attribute type specified with invalid value not_valid")
}
