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

func TestSlackDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	slackCred := SlackCredential{
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
		},
	}

	newCreds, err := slackCred.Resolve(context.TODO())
	assert.Nil(err)

	newSlackCreds := newCreds.(*SlackCredential)
	assert.Nil(newSlackCreds.Token)

	os.Setenv("SLACK_TOKEN", "foobar")

	newCreds, err = slackCred.Resolve(context.TODO())
	assert.Nil(err)

	newSlackCreds = newCreds.(*SlackCredential)
	assert.Equal("foobar", *newSlackCreds.Token)
}

func TestAbuseIPDBDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	abuseIPDBCred := AbuseIPDBCredential{
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
		},
	}

	newCreds, err := abuseIPDBCred.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBCreds := newCreds.(*AbuseIPDBCredential)
	assert.Nil(newAbuseIPDBCreds.APIKey)

	os.Setenv("ABUSEIPDB_API_KEY", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	newCreds, err = abuseIPDBCred.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBCreds = newCreds.(*AbuseIPDBCredential)
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", *newAbuseIPDBCreds.APIKey)
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
