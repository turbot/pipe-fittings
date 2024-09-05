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

func TestAwsConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *AwsConnection
	var conn2 *AwsConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &AwsConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessKey, SecretKey, SessionToken, Ttl, and Profile
	accessKey := "access_key_value"
	secretKey := "secret_key_value"
	sessionToken := "session_token_value"
	ttl := 3600
	profile := "profile_value"

	conn1 = &AwsConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		AccessKey:    &accessKey,
		SecretKey:    &secretKey,
		SessionToken: &sessionToken,
		Ttl:          &ttl,
		Profile:      &profile,
	}

	conn2 = &AwsConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		AccessKey:    &accessKey,
		SecretKey:    &secretKey,
		SessionToken: &sessionToken,
		Ttl:          &ttl,
		Profile:      &profile,
	}

	assert.True(conn1.Equals(conn2), "Both connections should have the same values and be equal")

	// Case 4: Connections have different AccessKey
	accessKey2 := "different_access_key"
	conn2.AccessKey = &accessKey2
	assert.False(conn1.Equals(conn2), "Connections have different AccessKeys, should return false")

	// Case 5: Connections have different SecretKey
	conn2.AccessKey = &accessKey // Reset AccessKey to the same
	secretKey2 := "different_secret_key"
	conn2.SecretKey = &secretKey2
	assert.False(conn1.Equals(conn2), "Connections have different SecretKeys, should return false")

	// Case 6: Connections have different Ttl
	conn2.SessionToken = &sessionToken // Reset SessionToken to the same
	ttl2 := 7200
	conn2.Ttl = &ttl2
	assert.False(conn1.Equals(conn2), "Connections have different Ttl values, should return false")

	// Case 7: Connections have different Profile
	conn2.Ttl = &ttl // Reset Ttl to the same
	profile2 := "different_profile"
	conn2.Profile = &profile2
	assert.False(conn1.Equals(conn2), "Connections have different Profile values, should return false")

	// NOTE: Finding out why we are not comparing SessionToken
	// // Case 8: Connections have different SessionToken
	// conn2.SecretKey = &secretKey // Reset SecretKey to the same
	// sessionToken2 := "different_session_token"
	// conn2.SessionToken = &sessionToken2
	// assert.False(conn1.Equals(conn2), "Connections have different SessionTokens, should return false")
}
