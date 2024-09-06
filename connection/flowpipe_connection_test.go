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

	newAbuseIPDBConnection := newConnection.(*AbuseIPDBConnection)
	assert.Equal("", *newAbuseIPDBConnection.APIKey)

	os.Setenv("ABUSEIPDB_API_KEY", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	newConnection, err = abuseIPDBConnection.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBConnection = newConnection.(*AbuseIPDBConnection)
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", *newAbuseIPDBConnection.APIKey)
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

// ------------------------------------------------------------
// BitBucket
// ------------------------------------------------------------

func TestBitbucketDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	bitbucketConnection := BitbucketConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("BITBUCKET_API_BASE_URL")
	os.Unsetenv("BITBUCKET_USERNAME")
	os.Unsetenv("BITBUCKET_PASSWORD")

	newConnection, err := bitbucketConnection.Resolve(context.TODO())
	assert.Nil(err)

	newBitbucketConnections := newConnection.(*BitbucketConnection)
	assert.Equal("", *newBitbucketConnections.BaseURL)
	assert.Equal("", *newBitbucketConnections.Username)
	assert.Equal("", *newBitbucketConnections.Password)

	os.Setenv("BITBUCKET_API_BASE_URL", "https://api.bitbucket.org/2.0")
	os.Setenv("BITBUCKET_USERNAME", "test@turbot.com")
	os.Setenv("BITBUCKET_PASSWORD", "ksfhashkfhakskashfghaskfagfgir327934gkegf")

	newConnection, err = bitbucketConnection.Resolve(context.TODO())
	assert.Nil(err)

	newBitbucketConnections = newConnection.(*BitbucketConnection)
	assert.Equal("https://api.bitbucket.org/2.0", *newBitbucketConnections.BaseURL)
	assert.Equal("test@turbot.com", *newBitbucketConnections.Username)
	assert.Equal("ksfhashkfhakskashfghaskfagfgir327934gkegf", *newBitbucketConnections.Password)
}

func TestBitbucketConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *BitbucketConnection
	var conn2 *BitbucketConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &BitbucketConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same BaseURL, Username, and Password
	baseURL := "https://bitbucket.org"
	username := "user123"
	password := "password123"

	conn1 = &BitbucketConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		BaseURL:  &baseURL,
		Username: &username,
		Password: &password,
	}

	conn2 = &BitbucketConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		BaseURL:  &baseURL,
		Username: &username,
		Password: &password,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same values and should be equal")

	// Case 4: Connections have different BaseURLs
	differentBaseURL := "https://different.bitbucket.org"
	conn2.BaseURL = &differentBaseURL
	assert.False(conn1.Equals(conn2), "Connections have different BaseURLs, should return false")

	// Case 5: Connections have different Usernames
	conn2.BaseURL = &baseURL // Reset BaseURL to match conn1
	differentUsername := "different_user"
	conn2.Username = &differentUsername
	assert.False(conn1.Equals(conn2), "Connections have different Usernames, should return false")

	// Case 6: Connections have different Passwords
	conn2.Username = &username // Reset Username to match conn1
	differentPassword := "different_password"
	conn2.Password = &differentPassword
	assert.False(conn1.Equals(conn2), "Connections have different Passwords, should return false")
}

func TestBitbucketConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty BitbucketConnection, should pass with no diagnostics
	conn := &BitbucketConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty BitbucketConnection")

	// Case 2: Validate a populated BitbucketConnection, should pass with no diagnostics
	baseURL := "https://bitbucket.org"
	username := "user123"
	password := "password123"

	conn = &BitbucketConnection{
		BaseURL:  &baseURL,
		Username: &username,
		Password: &password,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated BitbucketConnection")
}

// ------------------------------------------------------------
// ClickUp
// ------------------------------------------------------------

func TestClickUpDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	clickUpConnection := ClickUpConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("CLICKUP_TOKEN")
	newConnection, err := clickUpConnection.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpConnections := newConnection.(*ClickUpConnection)
	assert.Equal("", *newClickUpConnections.Token)

	os.Setenv("CLICKUP_TOKEN", "pk_616_L5H36X3CXXXXXXXWEAZZF0NM5")

	newConnection, err = clickUpConnection.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpConnections = newConnection.(*ClickUpConnection)
	assert.Equal("pk_616_L5H36X3CXXXXXXXWEAZZF0NM5", *newClickUpConnections.Token)
}

func TestClickUpConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *ClickUpConnection
	var conn2 *ClickUpConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &ClickUpConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &ClickUpConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	conn2 = &ClickUpConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same token and should be equal")

	// Case 4: Connections have different Tokens
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different tokens, should return false")
}

func TestClickUpConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty ClickUpConnection, should pass with no diagnostics
	conn := &ClickUpConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty ClickUpConnection")

	// Case 2: Validate a populated ClickUpConnection, should pass with no diagnostics
	token := "token_value"
	conn = &ClickUpConnection{
		Token: &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated ClickUpConnection")
}

// ------------------------------------------------------------
// Datadog
// ------------------------------------------------------------

func TestDatadogDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	datadogConnection := DatadogConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("DD_CLIENT_API_KEY")
	os.Unsetenv("DD_CLIENT_APP_KEY")

	newConnection, err := datadogConnection.Resolve(context.TODO())
	assert.Nil(err)

	newDatadogConnection := newConnection.(*DatadogConnection)
	assert.Equal("", *newDatadogConnection.APIKey)
	assert.Equal("", *newDatadogConnection.AppKey)

	os.Setenv("DD_CLIENT_API_KEY", "b1cf23432fwef23fg24grg31gr")
	os.Setenv("DD_CLIENT_APP_KEY", "1a2345bc23fwefrg13g233f")

	newConnection, err = datadogConnection.Resolve(context.TODO())
	assert.Nil(err)

	newDatadogConnection = newConnection.(*DatadogConnection)
	assert.Equal("b1cf23432fwef23fg24grg31gr", *newDatadogConnection.APIKey)
	assert.Equal("1a2345bc23fwefrg13g233f", *newDatadogConnection.AppKey)
}

func TestDatadogConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *DatadogConnection
	var conn2 *DatadogConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &DatadogConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey, AppKey, and APIUrl
	apiKey := "api_key_value" // #nosec G101
	appKey := "app_key_value"
	apiUrl := "api_url_value"

	conn1 = &DatadogConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIKey: &apiKey,
		AppKey: &appKey,
		APIUrl: &apiUrl,
	}

	conn2 = &DatadogConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIKey: &apiKey,
		AppKey: &appKey,
		APIUrl: &apiUrl,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same values and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")

	// Case 5: Connections have different AppKeys
	conn2.APIKey = &apiKey // Reset APIKey to match conn1
	differentAppKey := "different_app_key_value"
	conn2.AppKey = &differentAppKey
	assert.False(conn1.Equals(conn2), "Connections have different AppKeys, should return false")

	// Case 6: Connections have different APIUrls
	conn2.AppKey = &appKey // Reset AppKey to match conn1
	differentAPIUrl := "different_api_url_value"
	conn2.APIUrl = &differentAPIUrl
	assert.False(conn1.Equals(conn2), "Connections have different APIUrls, should return false")
}

func TestDatadogConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty DatadogConnection, should pass with no diagnostics
	conn := &DatadogConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty DatadogConnection")

	// Case 2: Validate a populated DatadogConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec G101
	appKey := "app_key_value"
	apiUrl := "api_url_value"

	conn = &DatadogConnection{
		APIKey: &apiKey,
		AppKey: &appKey,
		APIUrl: &apiUrl,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated DatadogConnection")
}

// ------------------------------------------------------------
// Discord
// ------------------------------------------------------------

func TestDiscordDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	discordConnection := DiscordConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("DISCORD_TOKEN")
	newConnection, err := discordConnection.Resolve(context.TODO())
	assert.Nil(err)

	newDiscordConnections := newConnection.(*DiscordConnection)
	assert.Equal("", *newDiscordConnections.Token)

	os.Setenv("DISCORD_TOKEN", "00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb")

	newConnection, err = discordConnection.Resolve(context.TODO())
	assert.Nil(err)

	newDiscordConnections = newConnection.(*DiscordConnection)
	assert.Equal("00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb", *newDiscordConnections.Token)
}

func TestDiscordConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *DiscordConnection
	var conn2 *DiscordConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &DiscordConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &DiscordConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	conn2 = &DiscordConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same token and should be equal")

	// Case 4: Connections have different Tokens
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different tokens, should return false")
}

func TestDiscordConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty DiscordConnection, should pass with no diagnostics
	conn := &DiscordConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty DiscordConnection")

	// Case 2: Validate a populated DiscordConnection, should pass with no diagnostics
	token := "token_value"
	conn = &DiscordConnection{
		Token: &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated DiscordConnection")
}

// ------------------------------------------------------------
// Freshdesk
// ------------------------------------------------------------

func TestFreshdeskDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	freshdeskConnection := FreshdeskConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("FRESHDESK_API_KEY")
	os.Unsetenv("FRESHDESK_SUBDOMAIN")

	newConnection, err := freshdeskConnection.Resolve(context.TODO())
	assert.Nil(err)

	newFreshdeskConnections := newConnection.(*FreshdeskConnection)
	assert.Equal("", *newFreshdeskConnections.APIKey)
	assert.Equal("", *newFreshdeskConnections.Subdomain)

	os.Setenv("FRESHDESK_API_KEY", "b1cf23432fwef23fg24grg31gr")
	os.Setenv("FRESHDESK_SUBDOMAIN", "sub-domain")

	newConnection, err = freshdeskConnection.Resolve(context.TODO())
	assert.Nil(err)

	newFreshdeskConnections = newConnection.(*FreshdeskConnection)
	assert.Equal("b1cf23432fwef23fg24grg31gr", *newFreshdeskConnections.APIKey)
	assert.Equal("sub-domain", *newFreshdeskConnections.Subdomain)
}

func TestFreshdeskConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *FreshdeskConnection
	var conn2 *FreshdeskConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &FreshdeskConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey and Subdomain
	apiKey := "api_key_value" // #nosec: G101
	subdomain := "subdomain_value"

	conn1 = &FreshdeskConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIKey:    &apiKey,
		Subdomain: &subdomain,
	}

	conn2 = &FreshdeskConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIKey:    &apiKey,
		Subdomain: &subdomain,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and Subdomain, and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")

	// Case 5: Connections have different Subdomains
	conn2.APIKey = &apiKey // Reset APIKey to match conn1
	differentSubdomain := "different_subdomain_value"
	conn2.Subdomain = &differentSubdomain
	assert.False(conn1.Equals(conn2), "Connections have different Subdomains, should return false")
}

func TestFreshdeskConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty FreshdeskConnection, should pass with no diagnostics
	conn := &FreshdeskConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty FreshdeskConnection")

	// Case 2: Validate a populated FreshdeskConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101
	subdomain := "subdomain_value"

	conn = &FreshdeskConnection{
		APIKey:    &apiKey,
		Subdomain: &subdomain,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated FreshdeskConnection")
}

// ------------------------------------------------------------
// GitHub
// ------------------------------------------------------------

func TestGitHubDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	githubConnection := GithubConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("GITHUB_TOKEN")
	newConnection, err := githubConnection.Resolve(context.TODO())
	assert.Nil(err)

	newGithubAccessTokenConnection := newConnection.(*GithubConnection)
	assert.Equal("", *newGithubAccessTokenConnection.Token)

	os.Setenv("GITHUB_TOKEN", "ghpat-ljgllghhegweroyuouo67u5476070owetylh")

	newConnection, err = githubConnection.Resolve(context.TODO())
	assert.Nil(err)

	newGithubAccessTokenConnection = newConnection.(*GithubConnection)
	assert.Equal("ghpat-ljgllghhegweroyuouo67u5476070owetylh", *newGithubAccessTokenConnection.Token)
}

func TestGithubConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *GithubConnection
	var conn2 *GithubConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &GithubConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &GithubConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	conn2 = &GithubConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same token and should be equal")

	// Case 4: Connections have different Tokens
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different tokens, should return false")
}

func TestGithubConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty GithubConnection, should pass with no diagnostics
	conn := &GithubConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty GithubConnection")

	// Case 2: Validate a populated GithubConnection, should pass with no diagnostics
	token := "token_value"
	conn = &GithubConnection{
		Token: &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated GithubConnection")
}

// ------------------------------------------------------------
// GitLab
// ------------------------------------------------------------

func TestGitLabDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	gitlabConnection := GitLabConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("GITLAB_TOKEN")
	newConnection, err := gitlabConnection.Resolve(context.TODO())
	assert.Nil(err)

	newGitLabAccessTokenConnection := newConnection.(*GitLabConnection)
	assert.Equal("", *newGitLabAccessTokenConnection.Token)

	os.Setenv("GITLAB_TOKEN", "glpat-ljgllghhegweroyuouo67u5476070owetylh")

	newConnection, err = gitlabConnection.Resolve(context.TODO())
	assert.Nil(err)

	newGitLabAccessTokenConnection = newConnection.(*GitLabConnection)
	assert.Equal("glpat-ljgllghhegweroyuouo67u5476070owetylh", *newGitLabAccessTokenConnection.Token)
}

func TestGitLabConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *GitLabConnection
	var conn2 *GitLabConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &GitLabConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &GitLabConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	conn2 = &GitLabConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		Token: &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same token and should be equal")

	// Case 4: Connections have different Tokens
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different tokens, should return false")
}

func TestGitLabConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty GitLabConnection, should pass with no diagnostics
	conn := &GitLabConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty GitLabConnection")

	// Case 2: Validate a populated GitLabConnection, should pass with no diagnostics
	token := "token_value"
	conn = &GitLabConnection{
		Token: &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated GitLabConnection")
}

// ------------------------------------------------------------
// Ipstack
// ------------------------------------------------------------

func TestIPstackDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	ipstackConnection := IPstackConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("IPSTACK_ACCESS_KEY")
	os.Unsetenv("IPSTACK_TOKEN")
	newConnection, err := ipstackConnection.Resolve(context.TODO())
	assert.Nil(err)

	newIPstackConnection := newConnection.(*IPstackConnection)
	assert.Equal("", *newIPstackConnection.AccessKey)

	os.Setenv("IPSTACK_ACCESS_KEY", "1234801bfsffsdf123455e6cfaf2")
	os.Setenv("IPSTACK_TOKEN", "1234801bfsffsdf123455e6cfaf2")

	newConnection, err = ipstackConnection.Resolve(context.TODO())
	assert.Nil(err)

	newIPstackConnection = newConnection.(*IPstackConnection)
	assert.Equal("1234801bfsffsdf123455e6cfaf2", *newIPstackConnection.AccessKey)
}

func TestIPstackConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *IPstackConnection
	var conn2 *IPstackConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &IPstackConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessKey
	accessKey := "access_key_value"
	conn1 = &IPstackConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		AccessKey: &accessKey,
	}

	conn2 = &IPstackConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		AccessKey: &accessKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same AccessKey and should be equal")

	// Case 4: Connections have different AccessKeys
	differentAccessKey := "different_access_key_value"
	conn2.AccessKey = &differentAccessKey
	assert.False(conn1.Equals(conn2), "Connections have different AccessKeys, should return false")
}

func TestIPstackConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty IPstackConnection, should pass with no diagnostics
	conn := &IPstackConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty IPstackConnection")

	// Case 2: Validate a populated IPstackConnection, should pass with no diagnostics
	accessKey := "access_key_value"
	conn = &IPstackConnection{
		AccessKey: &accessKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated IPstackConnection")
}

// ------------------------------------------------------------
// IP2LocationIO
// ------------------------------------------------------------

func TestIP2LocationIODefaultConnection(t *testing.T) {
	assert := assert.New(t)

	ip2LocationConnection := IP2LocationIOConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("IP2LOCATIONIO_API_KEY")
	newConnection, err := ip2LocationConnection.Resolve(context.TODO())
	assert.Nil(err)

	newIP2LocationConnections := newConnection.(*IP2LocationIOConnection)
	assert.Equal("", *newIP2LocationConnections.APIKey)

	os.Setenv("IP2LOCATIONIO_API_KEY", "12345678901A23BC4D5E6FG78HI9J101")

	newConnection, err = ip2LocationConnection.Resolve(context.TODO())
	assert.Nil(err)

	newIP2LocationConnections = newConnection.(*IP2LocationIOConnection)
	assert.Equal("12345678901A23BC4D5E6FG78HI9J101", *newIP2LocationConnections.APIKey)
}

func TestIP2LocationIOConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *IP2LocationIOConnection
	var conn2 *IP2LocationIOConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &IP2LocationIOConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec: G101
	conn1 = &IP2LocationIOConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIKey: &apiKey,
	}

	conn2 = &IP2LocationIOConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIKey: &apiKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")
}

func TestIP2LocationIOConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty IP2LocationIOConnection, should pass with no diagnostics
	conn := &IP2LocationIOConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty IP2LocationIOConnection")

	// Case 2: Validate a populated IP2LocationIOConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101
	conn = &IP2LocationIOConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated IP2LocationIOConnection")
}

// ------------------------------------------------------------
// Jira
// ------------------------------------------------------------

func TestJiraDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	jiraCred := JiraConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("JIRA_API_TOKEN")
	os.Unsetenv("JIRA_TOKEN")
	os.Unsetenv("JIRA_URL")
	os.Unsetenv("JIRA_USER")

	newConnection, err := jiraCred.Resolve(context.TODO())
	assert.Nil(err)

	newJiraCreds := newConnection.(*JiraConnection)
	assert.Equal("", *newJiraCreds.APIToken)
	assert.Equal("", *newJiraCreds.BaseURL)
	assert.Equal("", *newJiraCreds.Username)

	os.Setenv("JIRA_API_TOKEN", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("JIRA_TOKEN", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("JIRA_URL", "https://flowpipe-testorg.atlassian.net/")
	os.Setenv("JIRA_USER", "test@turbot.com")

	newConnection, err = jiraCred.Resolve(context.TODO())
	assert.Nil(err)

	newJiraCreds = newConnection.(*JiraConnection)
	assert.Equal("ksfhashkfhakskashfghaskfagfgir327934gkegf", *newJiraCreds.APIToken)
	assert.Equal("https://flowpipe-testorg.atlassian.net/", *newJiraCreds.BaseURL)
	assert.Equal("test@turbot.com", *newJiraCreds.Username)
}

func TestJiraConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *JiraConnection
	var conn2 *JiraConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &JiraConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIToken, BaseURL, and Username
	apiToken := "api_token_value"
	baseURL := "https://jira.example.com"
	username := "user123"

	conn1 = &JiraConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIToken: &apiToken,
		BaseURL:  &baseURL,
		Username: &username,
	}

	conn2 = &JiraConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		APIToken: &apiToken,
		BaseURL:  &baseURL,
		Username: &username,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIToken, BaseURL, and Username and should be equal")

	// Case 4: Connections have different APITokens
	differentAPIToken := "different_api_token_value"
	conn2.APIToken = &differentAPIToken
	assert.False(conn1.Equals(conn2), "Connections have different APITokens, should return false")

	// Case 5: Connections have different BaseURLs
	conn2.APIToken = &apiToken // Reset APIToken to match conn1
	differentBaseURL := "https://jira.different.com"
	conn2.BaseURL = &differentBaseURL
	assert.False(conn1.Equals(conn2), "Connections have different BaseURLs, should return false")

	// Case 6: Connections have different Usernames
	conn2.BaseURL = &baseURL // Reset BaseURL to match conn1
	differentUsername := "different_user"
	conn2.Username = &differentUsername
	assert.False(conn1.Equals(conn2), "Connections have different Usernames, should return false")
}

func TestJiraConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty JiraConnection, should pass with no diagnostics
	conn := &JiraConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty JiraConnection")

	// Case 2: Validate a populated JiraConnection, should pass with no diagnostics
	apiToken := "api_token_value"
	baseURL := "https://jira.example.com"
	username := "user123"

	conn = &JiraConnection{
		APIToken: &apiToken,
		BaseURL:  &baseURL,
		Username: &username,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated JiraConnection")
}

// ------------------------------------------------------------
// JumpCloud
// ------------------------------------------------------------

func TestJumpCloudDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	jumpCloudConnection := JumpCloudConnection{
		ConnectionImpl: ConnectionImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("JUMPCLOUD_API_KEY")

	newConnection, err := jumpCloudConnection.Resolve(context.TODO())
	assert.Nil(err)

	newJumpCloudConnection := newConnection.(*JumpCloudConnection)
	assert.Equal("", *newJumpCloudConnection.APIKey)

	os.Setenv("JUMPCLOUD_API_KEY", "sk-jwgthNa...")

	newConnection, err = jumpCloudConnection.Resolve(context.TODO())
	assert.Nil(err)

	newJumpCloudConnection = newConnection.(*JumpCloudConnection)
	assert.Equal("sk-jwgthNa...", *newJumpCloudConnection.APIKey)
}
