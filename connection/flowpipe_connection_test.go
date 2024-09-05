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
	newConnection, err := abuseIPDBConnection.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBConnections := newConnection.(*AbuseIPDBConnection)
	assert.Equal("", *newAbuseIPDBConnections.APIKey)

	os.Setenv("ABUSEIPDB_API_KEY", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	newConnection, err = abuseIPDBConnection.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBConnections = newConnection.(*AbuseIPDBConnection)
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", *newAbuseIPDBConnections.APIKey)
}

func TestAbuseIPDBConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *AbuseIPDBConnection
	var conn2 *AbuseIPDBConnection
	assert.True(conn1.Equals(conn2))

	// Case 2: One connection is nil
	conn1 = &AbuseIPDBConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil))

	// Case 3: Both connections have the same API key
	apiKey := "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d"
	conn1.APIKey = &apiKey
	conn2 = &AbuseIPDBConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIKey: &apiKey,
	}
	assert.True(conn1.Equals(conn2))

	// Case 4: Connections have different API keys
	apiKey2 := "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de1d"
	conn2.APIKey = &apiKey2
	assert.False(conn1.Equals(conn2))
}

func TestAwsConnection(t *testing.T) {

	assert := assert.New(t)

	awsCred := AwsConnection{}

	os.Setenv("AWS_ACCESS_KEY_ID", "foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "bar")

	newConnection, err := awsCred.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newConnection)

	newAwsCreds := newConnection.(*AwsConnection)

	assert.Equal("foo", *newAwsCreds.AccessKey)
	assert.Equal("bar", *newAwsCreds.SecretKey)
	assert.Nil(newAwsCreds.SessionToken)
}
