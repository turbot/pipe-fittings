package pipeline_test

import (
	"context"
	"github.com/turbot/pipe-fittings/load_mod"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParamOnEcho(t *testing.T) {
	assert := assert.New(t)

	_, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/param_on_echo.fp")
	assert.Nil(err, "error found")

}
