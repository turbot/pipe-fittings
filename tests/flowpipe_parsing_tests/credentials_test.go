package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestCredentials(t *testing.T) {
	assert := assert.New(t)

	mod, err := load_mod.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/credentials.fp")
	assert.Nil(err)
	assert.NotNil(mod)
	if mod == nil {
		return
	}

	credential := mod.ResourceMaps.Credentials["local.credential.aws.aws_static"]
	if credential == nil {
		assert.Fail("Credential not found")
		return
	}
}
