package modconfig

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAwsCredential(t *testing.T) {

	assert := assert.New(t)

	awsCred := AwsCredential{}

	os.Setenv("AWS_ACCESS_KEY_ID", "foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "bar")

	newCreds, err := awsCred.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newCreds)

	newAwsCreds := newCreds.(*AwsCredential)

	assert.Equal("foo", *newAwsCreds.AccessKey)
	assert.Equal("bar", *newAwsCreds.SecretKey)
	assert.Nil(newAwsCreds.SessionToken)
}

func XTestAwsCredentialRole(t *testing.T) {

	assert := assert.New(t)

	awsCred := AwsCredential{}

	newCreds, err := awsCred.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newCreds)

	newAwsCreds := newCreds.(*AwsCredential)

	assert.NotNil(newAwsCreds.SessionToken)
	assert.NotEqual("", newAwsCreds.SessionToken)
}
