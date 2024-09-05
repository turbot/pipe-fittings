package connection

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
)

// ------------------------------------------------------------
// AbuseIPDB
// ------------------------------------------------------------

func TestAbuseIPDBDefaultConnection(t *testing.T) {
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
	apiKey := "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d" // #nosec
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
	apiKey2 := "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de1d" // #nosec
	conn2.APIKey = &apiKey2
	assert.False(conn1.Equals(conn2))
}

func TestAbuseIPDBConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty AbuseIPDBConnection, should pass with no diagnostics
	conn := &AbuseIPDBConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty AbuseIPDBConnection")

	// Case 2: Validate a populated AbuseIPDBConnection, should pass with no diagnostics
	apiKey := "some_api_key"
	conn = &AbuseIPDBConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated AbuseIPDBConnection")
}

// ------------------------------------------------------------
// Alicloud
// ------------------------------------------------------------

func TestAlicloudConnection(t *testing.T) {

	assert := assert.New(t)

	alicloudConnection := AlicloudConnection{}

	os.Setenv("ALIBABACLOUD_ACCESS_KEY_ID", "foo")
	os.Setenv("ALIBABACLOUD_ACCESS_KEY_SECRET", "bar")

	newConnection, err := alicloudConnection.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newConnection)

	newAlicloudConnection := newConnection.(*AlicloudConnection)

	assert.Equal("foo", *newAlicloudConnection.AccessKey)
	assert.Equal("bar", *newAlicloudConnection.SecretKey)
}

func TestAlicloudConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *AlicloudConnection
	var conn2 *AlicloudConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &AlicloudConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessKey and SecretKey
	accessKey := "access_key_value"
	secretKey := "secret_key_value"
	conn1 = &AlicloudConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		AccessKey: &accessKey,
		SecretKey: &secretKey,
	}
	conn2 = &AlicloudConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		AccessKey: &accessKey,
		SecretKey: &secretKey,
	}
	assert.True(conn1.Equals(conn2), "Both connections have the same values and should be equal")

	// Case 4: Connections have different AccessKeys
	differentAccessKey := "different_access_key_value"
	conn2.AccessKey = &differentAccessKey
	assert.False(conn1.Equals(conn2), "Connections have different AccessKeys, should return false")

	// Case 5: Connections have different SecretKeys
	conn2.AccessKey = &accessKey // Reset AccessKey to match conn1
	differentSecretKey := "different_secret_key_value"
	conn2.SecretKey = &differentSecretKey
	assert.False(conn1.Equals(conn2), "Connections have different SecretKeys, should return false")
}

func TestAlicloudConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both AccessKey and SecretKey are nil, should pass validation
	conn := &AlicloudConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Both AccessKey and SecretKey are nil, validation should pass")

	// Case 2: AccessKey is defined, SecretKey is nil, should fail validation
	accessKey := "access_key_value"
	conn = &AlicloudConnection{
		AccessKey: &accessKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 1, "AccessKey defined without SecretKey, should return an error")
	assert.Equal(hcl.DiagError, diagnostics[0].Severity, "Severity should be DiagError")
	assert.Equal("access_key defined without secret_key", diagnostics[0].Summary)

	// Case 3: SecretKey is defined, AccessKey is nil, should fail validation
	secretKey := "secret_key_value"
	conn = &AlicloudConnection{
		SecretKey: &secretKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 1, "SecretKey defined without AccessKey, should return an error")
	assert.Equal(hcl.DiagError, diagnostics[0].Severity, "Severity should be DiagError")
	assert.Equal("secret_key defined without access_key", diagnostics[0].Summary)

	// Case 4: Both AccessKey and SecretKey are defined, should pass validation
	conn = &AlicloudConnection{
		AccessKey: &accessKey,
		SecretKey: &secretKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Both AccessKey and SecretKey are defined, validation should pass")
}

// ------------------------------------------------------------
// AWS
// ------------------------------------------------------------

func TestAwsConnection(t *testing.T) {

	assert := assert.New(t)

	awsConnecion := AwsConnection{}

	os.Setenv("AWS_ACCESS_KEY_ID", "foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "bar")

	newConnection, err := awsConnecion.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newConnection)

	newAwsConnection := newConnection.(*AwsConnection)

	assert.Equal("foo", *newAwsConnection.AccessKey)
	assert.Equal("bar", *newAwsConnection.SecretKey)
	assert.Nil(newAwsConnection.SessionToken)
}

// NOTE: We do not test for SessionToken as this is created in Resolve() and is not part of the connection configuration
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
}

func TestAwsConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both AccessKey and SecretKey are nil, should pass validation
	conn := &AwsConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Both AccessKey and SecretKey are nil, validation should pass")

	// Case 2: AccessKey is defined, SecretKey is nil, should fail validation
	accessKey := "access_key_value"
	conn = &AwsConnection{
		AccessKey: &accessKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 1, "AccessKey defined without SecretKey, should return an error")
	assert.Equal(hcl.DiagError, diagnostics[0].Severity, "Severity should be DiagError")
	assert.Equal("access_key defined without secret_key", diagnostics[0].Summary)

	// Case 3: SecretKey is defined, AccessKey is nil, should fail validation
	secretKey := "secret_key_value"
	conn = &AwsConnection{
		SecretKey: &secretKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 1, "SecretKey defined without AccessKey, should return an error")
	assert.Equal(hcl.DiagError, diagnostics[0].Severity, "Severity should be DiagError")
	assert.Equal("secret_key defined without access_key", diagnostics[0].Summary)

	// Case 4: Both AccessKey and SecretKey are defined, should pass validation
	conn = &AwsConnection{
		AccessKey: &accessKey,
		SecretKey: &secretKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Both AccessKey and SecretKey are defined, validation should pass")
}

// ------------------------------------------------------------
// Azure
// ------------------------------------------------------------

func TestAzureDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	azureConnection := AzureConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("AZURE_CLIENT_ID")
	os.Unsetenv("AZURE_CLIENT_SECRET")
	os.Unsetenv("AZURE_TENANT_ID")
	os.Unsetenv("AZURE_ENVIRONMENT")

	newConnection, err := azureConnection.Resolve(context.TODO())
	assert.Nil(err)

	newAzureConnections := newConnection.(*AzureConnection)
	assert.Equal("", *newAzureConnections.ClientID)
	assert.Equal("", *newAzureConnections.ClientSecret)
	assert.Equal("", *newAzureConnections.TenantID)
	assert.Equal("", *newAzureConnections.Environment)

	os.Setenv("AZURE_CLIENT_ID", "clienttoken")
	os.Setenv("AZURE_CLIENT_SECRET", "clientsecret")
	os.Setenv("AZURE_TENANT_ID", "tenantid")
	os.Setenv("AZURE_ENVIRONMENT", "environmentvar")

	newConnection, err = azureConnection.Resolve(context.TODO())
	assert.Nil(err)

	newAzureConnections = newConnection.(*AzureConnection)
	assert.Equal("clienttoken", *newAzureConnections.ClientID)
	assert.Equal("clientsecret", *newAzureConnections.ClientSecret)
	assert.Equal("tenantid", *newAzureConnections.TenantID)
	assert.Equal("environmentvar", *newAzureConnections.Environment)
}

func TestAzureConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *AzureConnection
	var conn2 *AzureConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &AzureConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same ClientID, ClientSecret, TenantID, and Environment
	clientID := "client_id_value"
	clientSecret := "client_secret_value"
	tenantID := "tenant_id_value"
	environment := "environment_value"

	conn1 = &AzureConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
		TenantID:     &tenantID,
		Environment:  &environment,
	}

	conn2 = &AzureConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
		TenantID:     &tenantID,
		Environment:  &environment,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same values and should be equal")

	// Case 4: Connections have different ClientIDs
	differentClientID := "different_client_id_value"
	conn2.ClientID = &differentClientID
	assert.False(conn1.Equals(conn2), "Connections have different ClientIDs, should return false")

	// Case 5: Connections have different ClientSecrets
	conn2.ClientID = &clientID // Reset ClientID to match conn1
	differentClientSecret := "different_client_secret_value"
	conn2.ClientSecret = &differentClientSecret
	assert.False(conn1.Equals(conn2), "Connections have different ClientSecrets, should return false")

	// Case 6: Connections have different TenantIDs
	conn2.ClientSecret = &clientSecret // Reset ClientSecret to match conn1
	differentTenantID := "different_tenant_id_value"
	conn2.TenantID = &differentTenantID
	assert.False(conn1.Equals(conn2), "Connections have different TenantIDs, should return false")

	// Case 7: Connections have different Environments
	conn2.TenantID = &tenantID // Reset TenantID to match conn1
	differentEnvironment := "different_environment_value"
	conn2.Environment = &differentEnvironment
	assert.False(conn1.Equals(conn2), "Connections have different Environments, should return false")
}

func TestAzureConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty AzureConnection, should pass with no diagnostics
	conn := &AzureConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty AzureConnection")

	// Case 2: Validate a populated AzureConnection, should pass with no diagnostics
	clientID := "client_id_value"
	clientSecret := "client_secret_value"
	tenantID := "tenant_id_value"
	environment := "environment_value"

	conn = &AzureConnection{
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
		TenantID:     &tenantID,
		Environment:  &environment,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated AzureConnection")
}
