package connection

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
)

func TestAbuseIPDBDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	abuseIPDBConnection := AbuseIPDBConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("ABUSEIPDB_API_KEY")
	newCreds, err := abuseIPDBConnection.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBConnections := newCreds.(*AbuseIPDBConnection)
	assert.Equal("", *newAbuseIPDBConnections.APIKey)

	os.Setenv("ABUSEIPDB_API_KEY", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	newCreds, err = abuseIPDBConnection.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBConnections = newCreds.(*AbuseIPDBConnection)
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", *newAbuseIPDBConnections.APIKey)
}

func TestAwsConnection(t *testing.T) {

	assert := assert.New(t)

	awsCred := AwsConnection{}

	os.Setenv("AWS_ACCESS_KEY_ID", "foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "bar")

	newCreds, err := awsCred.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newCreds)

	newAwsCreds := newCreds.(*AwsConnection)

	assert.Equal("foo", *newAwsCreds.AccessKey)
	assert.Equal("bar", *newAwsCreds.SecretKey)
	assert.Nil(newAwsCreds.SessionToken)
}
