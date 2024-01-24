package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestCredentialImport(t *testing.T) {
	assert := assert.New(t)

	mod, err := load_mod.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/credential_import.fp")
	assert.Nil(err)
	assert.NotNil(mod)
	if mod == nil {
		return
	}

	// credential := mod.ResourceMaps.Credentials["local.credential.aws.aws_static"]
	// if credential == nil {
	// 	assert.Fail("Credential not found")
	// 	return
	// }

	// assert.Equal("credential.aws.aws_static", credential.GetUnqualifiedName())
	// assert.Equal("aws", credential.GetCredentialType())

	// awsCred := credential.(*modconfig.AwsCredential)
	// assert.Equal("ASIAQGDFAKEKGUI5MCEU", *awsCred.AccessKey)
}
