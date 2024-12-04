package connection

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

// ------------------------------------------------------------
// AbuseIPDB
// ------------------------------------------------------------

func TestAbuseIPDBDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	abuseIPDBConnection := AbuseIPDBConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil))

	// Case 3: Both connections have the same API key
	apiKey := "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d" // #nosec
	conn1.APIKey = &apiKey
	conn2 = &AbuseIPDBConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessKey and SecretKey
	accessKey := "access_key_value"
	secretKey := "secret_key_value"
	conn1 = &AlicloudConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessKey: &accessKey,
		SecretKey: &secretKey,
	}
	conn2 = &AlicloudConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
			Ttl:       ttl,
		},
		AccessKey:    &accessKey,
		SecretKey:    &secretKey,
		SessionToken: &sessionToken,
		Profile:      &profile,
	}

	conn2 = &AwsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
			Ttl:       ttl,
		},
		AccessKey:    &accessKey,
		SecretKey:    &secretKey,
		SessionToken: &sessionToken,

		Profile: &profile,
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
	conn2.Ttl = ttl2
	assert.False(conn1.Equals(conn2), "Connections have different Ttl values, should return false")

	// Case 7: Connections have different Profile
	conn2.Ttl = ttl // Reset Ttl to the same
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
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
		TenantID:     &tenantID,
		Environment:  &environment,
	}

	conn2 = &AzureConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same BaseURL, Username, and Password
	baseURL := "https://bitbucket.org"
	username := "user123"
	password := "password123"

	conn1 = &BitbucketConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		BaseURL:  &baseURL,
		Username: &username,
		Password: &password,
	}

	conn2 = &BitbucketConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &ClickUpConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	conn2 = &ClickUpConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey, AppKey, and APIUrl
	apiKey := "api_key_value" // #nosec G101
	appKey := "app_key_value"
	apiUrl := "api_url_value"

	conn1 = &DatadogConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
		AppKey: &appKey,
		APIUrl: &apiUrl,
	}

	conn2 = &DatadogConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &DiscordConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	conn2 = &DiscordConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
// DuckDb
// ------------------------------------------------------------

func TestDuckDbConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *DuckDbConnection
	var conn2 *DuckDbConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &DuckDbConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same file name
	connectionString := "postgres://user:password@localhost:5432/dbname"
	conn1.FileName = &connectionString
	conn2 = &DuckDbConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		FileName: &connectionString,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same connection_string and should be equal")

	// Case 4: Connections have different file names
	connectionString2 := "postgres://user:password@localhost:5432/dbname2"
	conn2.FileName = &connectionString2
	assert.False(conn1.Equals(conn2), "Connections have different connection_string, should return false")

}

func TestDuckDbConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty DuckDbConnection, should pass with no diagnostics
	conn := &DuckDbConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should should pass with no diagnostics for an empty DuckDbConnection")

	// Case 2: Validate a DuckDbConnection with file name should pass with no diagnostics
	conn = &DuckDbConnection{
		FileName: utils.ToStringPointer("postgres://user:password@localhost:5432/dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a DuckDbConnection with connection_string")
}

// ------------------------------------------------------------
// Freshdesk
// ------------------------------------------------------------

func TestFreshdeskDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	freshdeskConnection := FreshdeskConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey and Subdomain
	apiKey := "api_key_value" // #nosec: G101
	subdomain := "subdomain_value"

	conn1 = &FreshdeskConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey:    &apiKey,
		Subdomain: &subdomain,
	}

	conn2 = &FreshdeskConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &GithubConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	conn2 = &GithubConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
// GCP
// ------------------------------------------------------------

// NOTE: We do not test for Resolve as this test requires a valid GCP credentials file

// NOTE: We do not test for AccessTokens as this is created in Resolve() and is not part of the connection configuration
func TestGcpConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *GcpConnection
	var conn2 *GcpConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &GcpConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same credentials, Ttl, and AccessToken
	credentials := "credentials_value"
	ttl := 3600
	accessToken := "access_token_value"

	conn1 = &GcpConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
			Ttl:       ttl,
		},
		Credentials: &credentials,
		AccessToken: &accessToken,
	}

	conn2 = &GcpConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
			Ttl:       ttl,
		},
		Credentials: &credentials,
		AccessToken: &accessToken,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same credentials, Ttl, and AccessToken, and should be equal")

	// Case 4: Connections have different credentials
	differentCredentials := "different_credentials_value"
	conn2.Credentials = &differentCredentials
	assert.False(conn1.Equals(conn2), "Connections have different credentials, should return false")

	// Case 5: Connections have different Ttl
	conn2.Credentials = &credentials // Reset credentials to match conn1
	differentTtl := 7200
	conn2.Ttl = differentTtl
	assert.False(conn1.Equals(conn2), "Connections have different Ttl, should return false")
}

func TestGcpConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty GcpConnection, should pass with no diagnostics
	conn := &GcpConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty GcpConnection")

	// Case 2: Validate a populated GcpConnection, should pass with no diagnostics
	credentials := "credentials_value"
	ttl := 3600
	accessToken := "access_token_value"

	conn = &GcpConnection{
		ConnectionImpl: ConnectionImpl{
			Ttl: ttl,
		},
		Credentials: &credentials,
		AccessToken: &accessToken,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated GcpConnection")
}

// ------------------------------------------------------------
// GitLab
// ------------------------------------------------------------

func TestGitLabDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	gitlabConnection := GitLabConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"
	conn1 = &GitLabConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	conn2 = &GitLabConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessKey
	accessKey := "access_key_value"
	conn1 = &IPstackConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessKey: &accessKey,
	}

	conn2 = &IPstackConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec: G101
	conn1 = &IP2LocationIOConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	conn2 = &IP2LocationIOConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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

	jiraConnection := JiraConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("JIRA_API_TOKEN")
	os.Unsetenv("JIRA_TOKEN")
	os.Unsetenv("JIRA_URL")
	os.Unsetenv("JIRA_USER")

	newConnection, err := jiraConnection.Resolve(context.TODO())
	assert.Nil(err)

	newJiraConnection := newConnection.(*JiraConnection)
	assert.Equal("", *newJiraConnection.APIToken)
	assert.Equal("", *newJiraConnection.BaseURL)
	assert.Equal("", *newJiraConnection.Username)

	os.Setenv("JIRA_API_TOKEN", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("JIRA_TOKEN", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("JIRA_URL", "https://flowpipe-testorg.atlassian.net/")
	os.Setenv("JIRA_USER", "test@turbot.com")

	newConnection, err = jiraConnection.Resolve(context.TODO())
	assert.Nil(err)

	newJiraConnection = newConnection.(*JiraConnection)
	assert.Equal("ksfhashkfhakskashfghaskfagfgir327934gkegf", *newJiraConnection.APIToken)
	assert.Equal("https://flowpipe-testorg.atlassian.net/", *newJiraConnection.BaseURL)
	assert.Equal("test@turbot.com", *newJiraConnection.Username)
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
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIToken, BaseURL, and Username
	apiToken := "api_token_value"
	baseURL := "https://jira.example.com"
	username := "user123"

	conn1 = &JiraConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIToken: &apiToken,
		BaseURL:  &baseURL,
		Username: &username,
	}

	conn2 = &JiraConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
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
			ShortName: "default",
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

func TestJumpCloudConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *JumpCloudConnection
	var conn2 *JumpCloudConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &JumpCloudConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec: G101
	conn1 = &JumpCloudConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	conn2 = &JumpCloudConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")
}

func TestJumpCloudConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty JumpCloudConnection, should pass with no diagnostics
	conn := &JumpCloudConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty JumpCloudConnection")

	// Case 2: Validate a populated JumpCloudConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101
	conn = &JumpCloudConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated JumpCloudConnection")
}

//	------------------------------------------------------------
//	Mastodon
//	------------------------------------------------------------

func TestMastodonDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	// Mastodon has no standard environment variable mentioned anywhere in the docs
	accessToken := "FK2_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU" //nolint:gosec // this is not a valid connection
	server := "https://myserver.social"

	mastodonConnection := MastodonConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessToken: &accessToken,
		Server:      &server,
	}

	newConnection, err := mastodonConnection.Resolve(context.TODO())
	assert.Nil(err)

	newMastodonConnection := newConnection.(*MastodonConnection)
	assert.Equal("FK2_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU", *newMastodonConnection.AccessToken)
	assert.Equal("https://myserver.social", *newMastodonConnection.Server)
}

func TestMastodonConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *MastodonConnection
	var conn2 *MastodonConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &MastodonConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessToken and Server
	accessToken := "access_token_value"
	server := "server_value"

	conn1 = &MastodonConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessToken: &accessToken,
		Server:      &server,
	}

	conn2 = &MastodonConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessToken: &accessToken,
		Server:      &server,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same AccessToken and Server and should be equal")

	// Case 4: Connections have different AccessTokens
	differentAccessToken := "different_access_token_value"
	conn2.AccessToken = &differentAccessToken
	assert.False(conn1.Equals(conn2), "Connections have different AccessTokens, should return false")

	// Case 5: Connections have different Servers
	conn2.AccessToken = &accessToken // Reset AccessToken to match conn1
	differentServer := "different_server_value"
	conn2.Server = &differentServer
	assert.False(conn1.Equals(conn2), "Connections have different Servers, should return false")
}

func TestMastodonConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty MastodonConnection, should pass with no diagnostics
	conn := &MastodonConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty MastodonConnection")

	// Case 2: Validate a populated MastodonConnection, should pass with no diagnostics
	accessToken := "access_token_value"
	server := "server_value"

	conn = &MastodonConnection{
		AccessToken: &accessToken,
		Server:      &server,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated MastodonConnection")
}

// ------------------------------------------------------------
// Microsoft Teams
// ------------------------------------------------------------

func TestMicrosoftTeamsDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	msTeamsConnection := MicrosoftTeamsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("TEAMS_ACCESS_TOKEN")
	newConnection, err := msTeamsConnection.Resolve(context.TODO())
	assert.Nil(err)

	newMSTeamsConnection := newConnection.(*MicrosoftTeamsConnection)
	assert.Equal("", *newMSTeamsConnection.AccessToken)

	os.Setenv("TEAMS_ACCESS_TOKEN", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	newConnection, err = msTeamsConnection.Resolve(context.TODO())
	assert.Nil(err)

	newMSTeamsConnection = newConnection.(*MicrosoftTeamsConnection)
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", *newMSTeamsConnection.AccessToken)
}

func TestMicrosoftTeamsConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *MicrosoftTeamsConnection
	var conn2 *MicrosoftTeamsConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &MicrosoftTeamsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessToken
	accessToken := "access_token_value"

	conn1 = &MicrosoftTeamsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessToken: &accessToken,
	}

	conn2 = &MicrosoftTeamsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessToken: &accessToken,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same AccessToken and should be equal")

	// Case 4: Connections have different AccessTokens
	differentAccessToken := "different_access_token_value"
	conn2.AccessToken = &differentAccessToken
	assert.False(conn1.Equals(conn2), "Connections have different AccessTokens, should return false")
}

func TestMicrosoftTeamsConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty MicrosoftTeamsConnection, should pass with no diagnostics
	conn := &MicrosoftTeamsConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty MicrosoftTeamsConnection")

	// Case 2: Validate a populated MicrosoftTeamsConnection, should pass with no diagnostics
	accessToken := "access_token_value"

	conn = &MicrosoftTeamsConnection{
		AccessToken: &accessToken,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated MicrosoftTeamsConnection")
}

// ------------------------------------------------------------
// Mysql
// ------------------------------------------------------------

func TestMysqlConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *MysqlConnection
	var conn2 *MysqlConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &MysqlConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	//Case 3: Connections have same settings
	conn1 = &MysqlConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		Host:     utils.ToStringPointer("localhost"),
		Port:     utils.ToIntegerPointer(5432),
		DbName:   utils.ToStringPointer("dbname"),
	}
	conn2 = &MysqlConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		Host:     utils.ToStringPointer("localhost"),
		Port:     utils.ToIntegerPointer(5432),
		DbName:   utils.ToStringPointer("dbname"),
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same settings and should be equal")

	//Case 4: Connections have different settings

	conn2.UserName = utils.ToStringPointer("user2")
	assert.False(conn1.Equals(conn2), "Both connections have the different user_name and should not be equal")

	// Case 5: Connection 2 is a subset of Connection 1
	conn2 = &MysqlConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		DbName:   utils.ToStringPointer("dbname"),
	}
	assert.False(conn1.Equals(conn2), "Connection 2 is a subset of Connection 1, should return false")

}

func TestMysqlConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty MysqlConnection, should  should pass with no with diagnostics
	conn := &MysqlConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should  should pass with no diagnostics for an empty MysqlConnection")

	// Case 2: Validate a MysqlConnection with user name and db should pass with no diagnostics
	conn = &MysqlConnection{
		UserName: utils.ToStringPointer("user"),
		DbName:   utils.ToStringPointer("dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a MysqlConnection with user name and db")

	// Case 4: Validate a MysqlConnection with user name but no db, should pass with no diagnostics
	conn = &MysqlConnection{
		UserName: utils.ToStringPointer("user"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a MysqlConnection with user name but no db")

	// Case 5: Validate a MysqlConnection with db name but no user, should pass with no diagnostics
	conn = &MysqlConnection{
		DbName: utils.ToStringPointer("dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should  pass with no diagnostics for a MysqlConnection with db name but no user")

	pipes := &PipesConnectionMetadata{
		Connection: utils.ToStringPointer("default"),
		User:       utils.ToStringPointer("user"),
		Org:        utils.ToStringPointer("org"),
		Workspace:  utils.ToStringPointer("ws"),
	}

	// Case 6: Validate a MysqlConnection with pipes block and no connection string, should pass with no diagnostics
	conn = &MysqlConnection{
		ConnectionImpl: ConnectionImpl{
			Pipes: pipes,
		},
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a MysqlConnection with pipes block and no connection string")

	// Case 7: Validate a MysqlConnection with pipes block and user name, should fail with diagnostics
	conn = &MysqlConnection{
		ConnectionImpl: ConnectionImpl{
			Pipes: pipes,
		},
		UserName: utils.ToStringPointer("user"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 1, "Validation should fail with 1 diagnostics for a MysqlConnection with pipes block and user")
}

func TestMysqlConnectionGetConnectionString(t *testing.T) {
	type fields struct {
		db       *string
		username *string
		host     *string
		port     *int
		password *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "all values are set",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
			},
			want: "mysql://user:password@tcp(host:1234)/db",
		},
		{
			name: "db, username, host and port",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
			},
			want: "mysql://user@tcp(host:1234)/db",
		},
		{
			name: "db, username, password",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				password: utils.ToStringPointer("password"),
			},
			want: "mysql://user:password@tcp(localhost:3306)/db",
		},
		{
			name: "no host",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
			},
			want: "mysql://user:password@tcp(localhost:1234)/db",
		},
		{
			name: "db and user only",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
			},
			want: "mysql://user@tcp(localhost:3306)/db",
		},
		{
			name: "user only",
			fields: fields{
				username: utils.ToStringPointer("user"),
			},
			want: "mysql://user@tcp(localhost:3306)/mysql",
		},
		{
			name: "db only",
			fields: fields{
				db: utils.ToStringPointer("db"),
			},
			want: "mysql://root@tcp(localhost:3306)/db",
		},
		{
			name:   "empty",
			fields: fields{},
			want:   "mysql://root@tcp(localhost:3306)/mysql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MysqlConnection{
				DbName:   tt.fields.db,
				UserName: tt.fields.username,
				Host:     tt.fields.host,
				Port:     tt.fields.port,
				Password: tt.fields.password,
			}
			// call validate to initialise the default values
			_ = c.Validate()
			cs, _ := c.GetConnectionString()
			assert.Equalf(t, tt.want, cs, "GetConnectionString()")
		})
	}
}

// ------------------------------------------------------------
// Okta
// ------------------------------------------------------------

func TestOktaDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	oktaConnection := OktaConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("OKTA_CLIENT_TOKEN")
	os.Unsetenv("OKTA_ORGURL")

	newConnection, err := oktaConnection.Resolve(context.TODO())
	assert.Nil(err)

	newOktaConnection := newConnection.(*OktaConnection)
	assert.Equal("", *newOktaConnection.Token)
	assert.Equal("", *newOktaConnection.Domain)

	os.Setenv("OKTA_CLIENT_TOKEN", "00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb")
	os.Setenv("OKTA_ORGURL", "https://dev-50078045.okta.com")

	newConnection, err = oktaConnection.Resolve(context.TODO())
	assert.Nil(err)

	newOktaConnection = newConnection.(*OktaConnection)
	assert.Equal("00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb", *newOktaConnection.Token)
	assert.Equal("https://dev-50078045.okta.com", *newOktaConnection.Domain)
}

func TestOktaConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *OktaConnection
	var conn2 *OktaConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &OktaConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Domain and Token
	domain := "example.okta.com"
	token := "token_value"

	conn1 = &OktaConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Domain: &domain,
		Token:  &token,
	}

	conn2 = &OktaConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Domain: &domain,
		Token:  &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same Domain and Token and should be equal")

	// Case 4: Connections have different Domains
	differentDomain := "different.okta.com"
	conn2.Domain = &differentDomain
	assert.False(conn1.Equals(conn2), "Connections have different Domains, should return false")

	// Case 5: Connections have different Tokens
	conn2.Domain = &domain // Reset Domain to match conn1
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different Tokens, should return false")
}

func TestOktaConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty OktaConnection, should pass with no diagnostics
	conn := &OktaConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty OktaConnection")

	// Case 2: Validate a populated OktaConnection, should pass with no diagnostics
	domain := "example.okta.com"
	token := "token_value"

	conn = &OktaConnection{
		Domain: &domain,
		Token:  &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated OktaConnection")
}

// ------------------------------------------------------------
// OpenAI
// ------------------------------------------------------------

func TestOpenAIDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	openAIConnection := OpenAIConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("OPENAI_API_KEY")

	newConnnection, err := openAIConnection.Resolve(context.TODO())
	assert.Nil(err)

	newOpenAIConnection := newConnnection.(*OpenAIConnection)
	assert.Equal("", *newOpenAIConnection.APIKey)

	os.Setenv("OPENAI_API_KEY", "sk-jwgthNa...")

	newConnnection, err = openAIConnection.Resolve(context.TODO())
	assert.Nil(err)

	newOpenAIConnection = newConnnection.(*OpenAIConnection)
	assert.Equal("sk-jwgthNa...", *newOpenAIConnection.APIKey)
}

func TestOpenAIConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *OpenAIConnection
	var conn2 *OpenAIConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &OpenAIConnection{}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec: G101
	conn1 = &OpenAIConnection{
		APIKey: &apiKey,
	}
	conn2 = &OpenAIConnection{
		APIKey: &apiKey,
	}
	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")
}

func TestOpenAIConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty OpenAIConnection, should pass with no diagnostics
	conn := &OpenAIConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty OpenAIConnection")

	// Case 2: Validate a populated OpenAIConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101
	conn = &OpenAIConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated OpenAIConnection")
}

// ------------------------------------------------------------
// OpsGenie
// ------------------------------------------------------------

func TestOpsgenieDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	opsgenieConnection := OpsgenieConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("OPSGENIE_ALERT_API_KEY")
	os.Unsetenv("OPSGENIE_INCIDENT_API_KEY")

	newConnection, err := opsgenieConnection.Resolve(context.TODO())
	assert.Nil(err)

	newOpsgenieConnection := newConnection.(*OpsgenieConnection)
	assert.Equal("", *newOpsgenieConnection.AlertAPIKey)
	assert.Equal("", *newOpsgenieConnection.IncidentAPIKey)

	os.Setenv("OPSGENIE_ALERT_API_KEY", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("OPSGENIE_INCIDENT_API_KEY", "jkgdgjdgjldjgdjlgjdlgjlgjldjgldjlgjdl")

	newConnection, err = opsgenieConnection.Resolve(context.TODO())
	assert.Nil(err)

	newOpsgenieConnection = newConnection.(*OpsgenieConnection)
	assert.Equal("ksfhashkfhakskashfghaskfagfgir327934gkegf", *newOpsgenieConnection.AlertAPIKey)
	assert.Equal("jkgdgjdgjldjgdjlgjdlgjlgjldjgldjlgjdl", *newOpsgenieConnection.IncidentAPIKey)
}

func TestOpsgenieConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *OpsgenieConnection
	var conn2 *OpsgenieConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &OpsgenieConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AlertAPIKey and IncidentAPIKey
	alertAPIKey := "alert_api_key_value" // #nosec: G101
	incidentAPIKey := "incident_api_key_value"

	conn1 = &OpsgenieConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AlertAPIKey:    &alertAPIKey,
		IncidentAPIKey: &incidentAPIKey,
	}

	conn2 = &OpsgenieConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AlertAPIKey:    &alertAPIKey,
		IncidentAPIKey: &incidentAPIKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same AlertAPIKey and IncidentAPIKey and should be equal")

	// Case 4: Connections have different AlertAPIKeys
	differentAlertAPIKey := "different_alert_api_key_value"
	conn2.AlertAPIKey = &differentAlertAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different AlertAPIKeys, should return false")

	// Case 5: Connections have different IncidentAPIKeys
	conn2.AlertAPIKey = &alertAPIKey // Reset AlertAPIKey to match conn1
	differentIncidentAPIKey := "different_incident_api_key_value"
	conn2.IncidentAPIKey = &differentIncidentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different IncidentAPIKeys, should return false")
}

func TestOpsgenieConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty OpsgenieConnection, should pass with no diagnostics
	conn := &OpsgenieConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty OpsgenieConnection")

	// Case 2: Validate a populated OpsgenieConnection, should pass with no diagnostics
	alertAPIKey := "alert_api_key_value" // #nosec: G101
	incidentAPIKey := "incident_api_key_value"

	conn = &OpsgenieConnection{
		AlertAPIKey:    &alertAPIKey,
		IncidentAPIKey: &incidentAPIKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated OpsgenieConnection")
}

// ------------------------------------------------------------
// PagerDuty
// ------------------------------------------------------------

func TestPagerDutyDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	pagerDutyConnection := PagerDutyConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("PAGERDUTY_TOKEN")
	newConnection, err := pagerDutyConnection.Resolve(context.TODO())
	assert.Nil(err)

	newPagerDutyConnection := newConnection.(*PagerDutyConnection)
	assert.Equal("", *newPagerDutyConnection.Token)

	os.Setenv("PAGERDUTY_TOKEN", "u+AtBdqvNtestTokeNcg")

	newConnection, err = pagerDutyConnection.Resolve(context.TODO())
	assert.Nil(err)

	newPagerDutyConnection = newConnection.(*PagerDutyConnection)
	assert.Equal("u+AtBdqvNtestTokeNcg", *newPagerDutyConnection.Token)
}

func TestPagerDutyConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *PagerDutyConnection
	var conn2 *PagerDutyConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &PagerDutyConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"

	conn1 = &PagerDutyConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	conn2 = &PagerDutyConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same Token and should be equal")

	// Case 4: Connections have different Tokens
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different Tokens, should return false")
}

func TestPagerDutyConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty PagerDutyConnection, should pass with no diagnostics
	conn := &PagerDutyConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty PagerDutyConnection")

	// Case 2: Validate a populated PagerDutyConnection, should pass with no diagnostics
	token := "token_value"

	conn = &PagerDutyConnection{
		Token: &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated PagerDutyConnection")
}

// ------------------------------------------------------------
// Postgres
// ------------------------------------------------------------

func TestPostgresConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *PostgresConnection
	var conn2 *PostgresConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &PostgresConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	//Case 3: Connections have same settings
	conn1 = &PostgresConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		Host:     utils.ToStringPointer("localhost"),
		Port:     utils.ToIntegerPointer(5432),
		DbName:   utils.ToStringPointer("dbname"),
	}
	conn2 = &PostgresConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		Host:     utils.ToStringPointer("localhost"),
		Port:     utils.ToIntegerPointer(5432),
		DbName:   utils.ToStringPointer("dbname"),
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same settings and should be equal")

	//Case 4: Connections have different settings

	conn2.UserName = utils.ToStringPointer("user2")
	assert.False(conn1.Equals(conn2), "Both connections have the different user_name and should not be equal")

	// Case 5: Connection 2 is a subset of Connection 1
	conn2 = &PostgresConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		DbName:   utils.ToStringPointer("dbname"),
	}
	assert.False(conn1.Equals(conn2), "Connection 2 is a subset of Connection 1, should return false")

}

func TestPostgresConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty PostgresConnection, should should pass with no diagnostics
	conn := &PostgresConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should should pass with no diagnostics for an empty PostgresConnection")

	// Case 2: Validate a PostgresConnection with user name and db should pass with no diagnostics
	conn = &PostgresConnection{
		UserName: utils.ToStringPointer("user"),
		DbName:   utils.ToStringPointer("dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a PostgresConnection with user name and db")

	// Case 3: Validate a PostgresConnection with user name but no db should pass with no diagnostics
	conn = &PostgresConnection{
		UserName: utils.ToStringPointer("user"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a PostgresConnection with user name but no db")

	// Case 4: Validate a PostgresConnection with db name but no user, should pass with no diagnostics
	conn = &PostgresConnection{
		DbName: utils.ToStringPointer("dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a PostgresConnection with db name but no user")

	pipes := &PipesConnectionMetadata{
		Connection: utils.ToStringPointer("default"),
		User:       utils.ToStringPointer("user"),
		Org:        utils.ToStringPointer("org"),
		Workspace:  utils.ToStringPointer("ws"),
	}

	// Case 5: Validate a PostgresConnection with pipes block and no connection string, should pass with no diagnostics
	conn = &PostgresConnection{
		ConnectionImpl: ConnectionImpl{
			Pipes: pipes,
		},
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a PostgresConnection with pipes block and no connection string")

	// Case 6: Validate a PostgresConnection with pipes block and user name, should fail with diagnostics
	conn = &PostgresConnection{
		ConnectionImpl: ConnectionImpl{
			Pipes: pipes,
		},
		UserName: utils.ToStringPointer("user"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 1, "Validation should fail with 1 diagnostics for a PostgresConnection with pipes block and user")
}

func TestPostgresConnectionGetConnectionString(t *testing.T) {
	type fields struct {
		db       *string
		username *string
		host     *string
		port     *int
		password *string
		sslMode  *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "all values are set",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: "postgresql://user:password@host:1234/db?sslmode=allow",
		},
		{
			name: "db user host and port",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
			},
			want: "postgresql://user@host:1234/db",
		},
		{
			name: "db, user, password",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				password: utils.ToStringPointer("password"),
			},
			want: "postgresql://user:password@localhost:5432/db",
		},
		{
			name: "no host",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: "postgresql://user:password@localhost:1234/db?sslmode=allow",
		},
		{
			name: "db and user only",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
			},
			want: "postgresql://user@localhost:5432/db",
		},
		{
			name: "user only",
			fields: fields{
				username: utils.ToStringPointer("user"),
			},
			want: "postgresql://user@localhost:5432/postgres",
		},
		{
			name: "db only",
			fields: fields{
				db: utils.ToStringPointer("db"),
			},
			want: "postgresql://postgres@localhost:5432/db",
		},
		{
			name:   "empty",
			fields: fields{},
			want:   "postgresql://postgres@localhost:5432/postgres",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PostgresConnection{
				DbName:   tt.fields.db,
				UserName: tt.fields.username,
				Host:     tt.fields.host,
				Port:     tt.fields.port,
				Password: tt.fields.password,
				SslMode:  tt.fields.sslMode,
			}
			// call validate to initialise the default values
			_ = c.Validate()
			cs, _ := c.GetConnectionString()
			assert.Equalf(t, tt.want, cs, "GetConnectionString()")
		})
	}
}

func TestPostgresConnectionGetEnv(t *testing.T) {
	type fields struct {
		db       *string
		username *string
		host     *string
		port     *int
		password *string
		sslMode  *string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]cty.Value
	}{
		{
			name: "all values are set",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
				"PGSSLMODE":  cty.StringVal("allow"),
			},
		},
		{
			name: "db, user, host and port",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
			},
		},
		{
			name: "db, user, password",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				password: utils.ToStringPointer("password"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultPostgresPort)),
				"PGHOST":     cty.StringVal(defaultPostgresHost),
			},
		},
		{
			name: "no host",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
				"PGHOST":     cty.StringVal(defaultPostgresHost),
				"PGSSLMODE":  cty.StringVal("allow"),
			},
		},
		{
			name: "db and user only",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultPostgresPort)),
				"PGHOST":     cty.StringVal(defaultPostgresHost),
			},
		},
		{
			name: "user only",
			fields: fields{
				username: utils.ToStringPointer("user"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal(defaultPostgresDbName),
				"PGUSER":     cty.StringVal("user"),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultPostgresPort)),
				"PGHOST":     cty.StringVal(defaultPostgresHost),
			},
		},
		{
			name: "db only",
			fields: fields{
				db: utils.ToStringPointer("db"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal(defaultPostgresDbName),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultPostgresPort)),
				"PGHOST":     cty.StringVal(defaultPostgresHost),
			},
		},
		{
			name:   "empty",
			fields: fields{},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal(defaultPostgresDbName),
				"PGUSER":     cty.StringVal(defaultPostgresDbName),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultPostgresPort)),
				"PGHOST":     cty.StringVal(defaultPostgresHost),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PostgresConnection{
				DbName:   tt.fields.db,
				UserName: tt.fields.username,
				Host:     tt.fields.host,
				Port:     tt.fields.port,
				Password: tt.fields.password,
				SslMode:  tt.fields.sslMode,
			}
			// call validate to initialise the default values
			_ = c.Validate()

			assert.Equalf(t, tt.want, c.GetEnv(), "GetEnv()")
		})
	}
}

// ------------------------------------------------------------
// SendGrid
// ------------------------------------------------------------

func TestSendGridDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	sendGridConnection := SendGridConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("SENDGRID_API_KEY")
	newConnection, err := sendGridConnection.Resolve(context.TODO())
	assert.Nil(err)

	newSendGridConnection := newConnection.(*SendGridConnection)
	assert.Equal("", *newSendGridConnection.APIKey)

	os.Setenv("SENDGRID_API_KEY", "SGsomething")

	newConnection, err = sendGridConnection.Resolve(context.TODO())
	assert.Nil(err)

	newSendGridConnection = newConnection.(*SendGridConnection)
	assert.Equal("SGsomething", *newSendGridConnection.APIKey)
}

func TestSendGridConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *SendGridConnection
	var conn2 *SendGridConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &SendGridConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec

	conn1 = &SendGridConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	conn2 = &SendGridConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")
}

func TestSendGridConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty SendGridConnection, should pass with no diagnostics
	conn := &SendGridConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty SendGridConnection")

	// Case 2: Validate a populated SendGridConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec

	conn = &SendGridConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated SendGridConnection")
}

// ------------------------------------------------------------
// ServiceNow
// ------------------------------------------------------------

func TestServiceNowDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	guardrailsConnection := ServiceNowConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("SERVICENOW_INSTANCE_URL")
	os.Unsetenv("SERVICENOW_USERNAME")
	os.Unsetenv("SERVICENOW_PASSWORD")

	newConnection, err := guardrailsConnection.Resolve(context.TODO())
	assert.Nil(err)

	newServiceNowConnection := newConnection.(*ServiceNowConnection)
	assert.Equal("", *newServiceNowConnection.InstanceURL)
	assert.Equal("", *newServiceNowConnection.Username)
	assert.Equal("", *newServiceNowConnection.Password)

	os.Setenv("SERVICENOW_INSTANCE_URL", "https://a1b2c3d4.service-now.com")
	os.Setenv("SERVICENOW_USERNAME", "john.hill")
	os.Setenv("SERVICENOW_PASSWORD", "j0t3-$j@H3")

	newConnection, err = guardrailsConnection.Resolve(context.TODO())
	assert.Nil(err)

	newServiceNowConnection = newConnection.(*ServiceNowConnection)
	assert.Equal("https://a1b2c3d4.service-now.com", *newServiceNowConnection.InstanceURL)
	assert.Equal("john.hill", *newServiceNowConnection.Username)
	assert.Equal("j0t3-$j@H3", *newServiceNowConnection.Password)
}

func TestServiceNowConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *ServiceNowConnection
	var conn2 *ServiceNowConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &ServiceNowConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same InstanceURL, Username, and Password
	instanceURL := "https://servicenow.example.com"
	username := "user123"
	password := "password123"

	conn1 = &ServiceNowConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		InstanceURL: &instanceURL,
		Username:    &username,
		Password:    &password,
	}

	conn2 = &ServiceNowConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		InstanceURL: &instanceURL,
		Username:    &username,
		Password:    &password,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same InstanceURL, Username, and Password and should be equal")

	// Case 4: Connections have different InstanceURLs
	differentInstanceURL := "https://different.servicenow.com"
	conn2.InstanceURL = &differentInstanceURL
	assert.False(conn1.Equals(conn2), "Connections have different InstanceURLs, should return false")

	// Case 5: Connections have different Usernames
	conn2.InstanceURL = &instanceURL // Reset InstanceURL to match conn1
	differentUsername := "different_user"
	conn2.Username = &differentUsername
	assert.False(conn1.Equals(conn2), "Connections have different Usernames, should return false")

	// Case 6: Connections have different Passwords
	conn2.Username = &username // Reset Username to match conn1
	differentPassword := "different_password"
	conn2.Password = &differentPassword
	assert.False(conn1.Equals(conn2), "Connections have different Passwords, should return false")
}

func TestServiceNowConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty ServiceNowConnection, should pass with no diagnostics
	conn := &ServiceNowConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty ServiceNowConnection")

	// Case 2: Validate a populated ServiceNowConnection, should pass with no diagnostics
	instanceURL := "https://servicenow.example.com"
	username := "user123"
	password := "password123"

	conn = &ServiceNowConnection{
		InstanceURL: &instanceURL,
		Username:    &username,
		Password:    &password,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated ServiceNowConnection")
}

// ------------------------------------------------------------
// Slack
// ------------------------------------------------------------

func TestSlackConnectionResolve(t *testing.T) {
	assert := assert.New(t)

	// Simulate a Slack connection with no token initially
	slackConnection := SlackConnection{}

	// Test case 1: No SLACK_TOKEN environment variable, should return connection with empty token
	os.Unsetenv("SLACK_TOKEN")
	newConnection, err := slackConnection.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newConnection)

	newSlackConnection, ok := newConnection.(*SlackConnection)
	assert.True(ok)
	assert.NotNil(newSlackConnection.Token)
	assert.Equal("", *newSlackConnection.Token)

	// Test case 2: SLACK_TOKEN environment variable is set
	os.Setenv("SLACK_TOKEN", "xoxb-12345-mock-slack-token")

	newConnection, err = slackConnection.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newConnection)

	newSlackConnection, ok = newConnection.(*SlackConnection)
	assert.True(ok)
	assert.NotNil(newSlackConnection.Token)                                // The Token field should not be nil
	assert.Equal("xoxb-12345-mock-slack-token", *newSlackConnection.Token) // The Token should match the environment variable

	// Cleanup
	os.Unsetenv("SLACK_TOKEN")
}

func TestSlackConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *SlackConnection
	var conn2 *SlackConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &SlackConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"

	conn1 = &SlackConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	conn2 = &SlackConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same Token and should be equal")

	// Case 4: Connections have different Tokens
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different Tokens, should return false")
}

func TestSlackConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty SlackConnection, should pass with no diagnostics
	conn := &SlackConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty SlackConnection")

	// Case 2: Validate a populated SlackConnection, should pass with no diagnostics
	token := "token_value"

	conn = &SlackConnection{
		Token: &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated SlackConnection")
}

// ------------------------------------------------------------
// Sqlite
// ------------------------------------------------------------

func TestSqliteConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *SqliteConnection
	var conn2 *SqliteConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &SqliteConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same  file name
	connectionString := "postgres://user:password@localhost:5432/dbname"
	conn1.FileName = &connectionString
	conn2 = &SqliteConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		FileName: &connectionString,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same connection_string and should be equal")

	// Case 4: Connections have different file name
	connectionString2 := "postgres://user:password@localhost:5432/dbname2"
	conn2.FileName = &connectionString2
	assert.False(conn1.Equals(conn2), "Connections have different connection_string, should return false")

}

func TestSqliteConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty SqliteConnection, should should pass with no diagnostics
	conn := &SqliteConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should should pass with no diagnostics for an empty SqliteConnection")

	// Case 2: Validate a SqliteConnection with  file name should pass with no diagnostics
	conn = &SqliteConnection{
		FileName: utils.ToStringPointer("postgres://user:password@localhost:5432/dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a SqliteConnection with connection_string")
}

// ------------------------------------------------------------
// SteampipePg
// ------------------------------------------------------------

func TestSteampipePgConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *SteampipePgConnection
	var conn2 *SteampipePgConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &SteampipePgConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	//Case 3: Connections have same settings
	conn1 = &SteampipePgConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		Host:     utils.ToStringPointer("localhost"),
		Port:     utils.ToIntegerPointer(5432),
		DbName:   utils.ToStringPointer("dbname"),
	}
	conn2 = &SteampipePgConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		Host:     utils.ToStringPointer("localhost"),
		Port:     utils.ToIntegerPointer(5432),
		DbName:   utils.ToStringPointer("dbname"),
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same settings and should be equal")

	//Case 4: Connections have different settings

	conn2.UserName = utils.ToStringPointer("user2")
	assert.False(conn1.Equals(conn2), "Both connections have the different user_name and should not be equal")

	// Case 5: Connection 2 is a subset of Connection 1
	conn2 = &SteampipePgConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		UserName: utils.ToStringPointer("user"),
		Password: utils.ToStringPointer("password"),
		DbName:   utils.ToStringPointer("dbname"),
	}
	assert.False(conn1.Equals(conn2), "Connection 2 is a subset of Connection 1, should return false")
}

func TestSteampipePgConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty SteampipePgConnection, should pass with no diagnostics
	conn := &SteampipePgConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty SteampipePgConnection")

	// Case 2: Validate a SteampipePgConnection with user name and db should pass with no diagnostics
	conn = &SteampipePgConnection{
		UserName: utils.ToStringPointer("user"),
		DbName:   utils.ToStringPointer("dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a SteampipePgConnection with user name and db")

	// Case 3: Validate a SteampipePgConnection with user name but no db, should should pass with no diagnostics
	conn = &SteampipePgConnection{
		UserName: utils.ToStringPointer("user"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should should pass with no diagnostics for a SteampipePgConnection with user name but no db")

	// Case 4: Validate a SteampipePgConnection with db name but no user, should should pass with no diagnostics
	conn = &SteampipePgConnection{
		DbName: utils.ToStringPointer("dbname"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should should pass with no diagnostics for a SteampipePgConnection with db name but no user")

	pipes := &PipesConnectionMetadata{
		Connection: utils.ToStringPointer("default"),
		User:       utils.ToStringPointer("user"),
		Org:        utils.ToStringPointer("org"),
		Workspace:  utils.ToStringPointer("ws"),
	}

	// Case 5: Validate a SteampipePgConnection with pipes block and no connection string, should pass with no diagnostics
	conn = &SteampipePgConnection{
		ConnectionImpl: ConnectionImpl{
			Pipes: pipes,
		},
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a SteampipePgConnection with pipes block and no connection string")

	// Case 6: Validate a SteampipePgConnection with pipes block and user name, should fail with diagnostics
	conn = &SteampipePgConnection{
		ConnectionImpl: ConnectionImpl{
			Pipes: pipes,
		},
		UserName: utils.ToStringPointer("user"),
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 1, "Validation should fail with 1 diagnostics for a SteampipePgConnection with pipes block and user")
}

func TestSteampipeConnectionGetConnectionString(t *testing.T) {
	type fields struct {
		db       *string
		username *string
		host     *string
		port     *int
		password *string
		sslMode  *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "all values are set",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: "postgresql://user:password@host:1234/db?sslmode=allow",
		},
		{
			name: "db user host and port",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
			},
			want: "postgresql://user@host:1234/db",
		},
		{
			name: "password",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				password: utils.ToStringPointer("password"),
			},
			want: "postgresql://user:password@localhost:9193/db",
		},
		{
			name: "no host",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: "postgresql://user:password@localhost:1234/db?sslmode=allow",
		},
		{
			name: "db and user only",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
			},
			want: "postgresql://user@localhost:9193/db",
		},
		{
			name: "user only",
			fields: fields{
				username: utils.ToStringPointer("user"),
			},
			want: "postgresql://user@localhost:9193/steampipe",
		},
		{
			name: "no user",
			fields: fields{
				db: utils.ToStringPointer("db"),
			},
			want: "postgresql://steampipe@localhost:9193/db",
		},
		{
			name:   "empty",
			fields: fields{},
			want:   "postgresql://steampipe@localhost:9193/steampipe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SteampipePgConnection{
				DbName:   tt.fields.db,
				UserName: tt.fields.username,
				Host:     tt.fields.host,
				Port:     tt.fields.port,
				Password: tt.fields.password,
				SslMode:  tt.fields.sslMode,
			}
			cs, _ := c.GetConnectionString()
			assert.Equalf(t, tt.want, cs, "GetConnectionString()")
		})
	}
}

func TestSteampipeConnectionGetEnv(t *testing.T) {
	type fields struct {
		db       *string
		username *string
		host     *string
		port     *int
		password *string
		sslMode  *string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]cty.Value
	}{
		{
			name: "all values are set",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
				"PGSSLMODE":  cty.StringVal("allow"),
			},
		},
		{
			name: "db, user, host and port",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
			},
		},
		{
			name: "db, user, password",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				password: utils.ToStringPointer("password"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultSteampipePort)),
				"PGHOST":     cty.StringVal(defaultSteampipeHost),
			},
		},
		{
			name: "no host",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
				"PGHOST":     cty.StringVal(defaultSteampipeHost),
				"PGSSLMODE":  cty.StringVal("allow"),
			},
		},
		{
			name: "db and user only",
			fields: fields{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultSteampipePort)),
				"PGHOST":     cty.StringVal(defaultSteampipeHost),
			},
		},
		{
			name: "user only",
			fields: fields{
				username: utils.ToStringPointer("user"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal(defaultSteampipeDbName),
				"PGUSER":     cty.StringVal("user"),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultSteampipePort)),
				"PGHOST":     cty.StringVal(defaultSteampipeHost),
			},
		},
		{
			name: "db only",
			fields: fields{
				db: utils.ToStringPointer("db"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal(defaultSteampipeDbName),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultSteampipePort)),
				"PGHOST":     cty.StringVal(defaultSteampipeHost),
			},
		},
		{
			name:   "empty",
			fields: fields{},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal(defaultSteampipeDbName),
				"PGUSER":     cty.StringVal(defaultSteampipeDbName),
				"PGPORT":     cty.StringVal(strconv.Itoa(defaultSteampipePort)),
				"PGHOST":     cty.StringVal(defaultSteampipeHost),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SteampipePgConnection{
				DbName:   tt.fields.db,
				UserName: tt.fields.username,
				Host:     tt.fields.host,
				Port:     tt.fields.port,
				Password: tt.fields.password,
				SslMode:  tt.fields.sslMode,
			}
			assert.Equalf(t, tt.want, c.GetEnv(), "GetEnv()")
		})
	}
}

// ------------------------------------------------------------
// Trello
// ------------------------------------------------------------

func TestTrelloDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	trelloConnection := TrelloConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("TRELLO_API_KEY")
	os.Unsetenv("TRELLO_TOKEN")

	newConnection, err := trelloConnection.Resolve(context.TODO())
	assert.Nil(err)

	newTrelloConnection := newConnection.(*TrelloConnection)
	assert.Equal("", *newTrelloConnection.APIKey)
	assert.Equal("", *newTrelloConnection.Token)

	os.Setenv("TRELLO_API_KEY", "dmgdhdfhfhfhi")
	os.Setenv("TRELLO_TOKEN", "17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4")

	newConnection, err = trelloConnection.Resolve(context.TODO())
	assert.Nil(err)

	newTrelloConnection = newConnection.(*TrelloConnection)
	assert.Equal("dmgdhdfhfhfhi", *newTrelloConnection.APIKey)
	assert.Equal("17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4", *newTrelloConnection.Token)
}

func TestTrelloConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *TrelloConnection
	var conn2 *TrelloConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &TrelloConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey and Token
	apiKey := "api_key_value" // #nosec: G101
	token := "token_value"

	conn1 = &TrelloConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
		Token:  &token,
	}

	conn2 = &TrelloConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
		Token:  &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and Token and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")

	// Case 5: Connections have different Tokens
	conn2.APIKey = &apiKey // Reset APIKey to match conn1
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different Tokens, should return false")
}

func TestTrelloConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty TrelloConnection, should pass with no diagnostics
	conn := &TrelloConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty TrelloConnection")

	// Case 2: Validate a populated TrelloConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101
	token := "token_value"

	conn = &TrelloConnection{
		APIKey: &apiKey,
		Token:  &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated TrelloConnection")
}

// ------------------------------------------------------------
// Turbot Guardrails
// ------------------------------------------------------------

func TestGuardrailsDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	guardrailsConnection := GuardrailsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("TURBOT_ACCESS_KEY")
	os.Unsetenv("TURBOT_SECRET_KEY")
	os.Unsetenv("TURBOT_WORKSPACE")

	newConnection, err := guardrailsConnection.Resolve(context.TODO())
	assert.Nil(err)

	newGuardrailsConnection := newConnection.(*GuardrailsConnection)
	assert.Equal("", *newGuardrailsConnection.AccessKey)
	assert.Equal("", *newGuardrailsConnection.SecretKey)

	os.Setenv("TURBOT_ACCESS_KEY", "c8e2c2ed-1ca8-429b-b369-123")
	os.Setenv("TURBOT_SECRET_KEY", "a3d8385d-47f7-40c5-a90c-123")
	os.Setenv("TURBOT_WORKSPACE", "https://my_workspace.saas.turbot.com")

	newConnection, err = guardrailsConnection.Resolve(context.TODO())
	assert.Nil(err)

	newGuardrailsConnection = newConnection.(*GuardrailsConnection)
	assert.Equal("c8e2c2ed-1ca8-429b-b369-123", *newGuardrailsConnection.AccessKey)
	assert.Equal("a3d8385d-47f7-40c5-a90c-123", *newGuardrailsConnection.SecretKey)
}

func TestGuardrailsConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *GuardrailsConnection
	var conn2 *GuardrailsConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &GuardrailsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same AccessKey, SecretKey, and Workspace
	accessKey := "access_key_value"
	secretKey := "secret_key_value"
	workspace := "workspace_value"

	conn1 = &GuardrailsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessKey: &accessKey,
		SecretKey: &secretKey,
		Workspace: &workspace,
	}

	conn2 = &GuardrailsConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		AccessKey: &accessKey,
		SecretKey: &secretKey,
		Workspace: &workspace,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same AccessKey, SecretKey, and Workspace and should be equal")

	// Case 4: Connections have different AccessKeys
	differentAccessKey := "different_access_key_value"
	conn2.AccessKey = &differentAccessKey
	assert.False(conn1.Equals(conn2), "Connections have different AccessKeys, should return false")

	// Case 5: Connections have different SecretKeys
	conn2.AccessKey = &accessKey // Reset AccessKey to match conn1
	differentSecretKey := "different_secret_key_value"
	conn2.SecretKey = &differentSecretKey
	assert.False(conn1.Equals(conn2), "Connections have different SecretKeys, should return false")

	// Case 6: Connections have different Workspaces
	conn2.SecretKey = &secretKey // Reset SecretKey to match conn1
	differentWorkspace := "different_workspace_value"
	conn2.Workspace = &differentWorkspace
	assert.False(conn1.Equals(conn2), "Connections have different Workspaces, should return false")
}

func TestGuardrailsConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty GuardrailsConnection, should pass with no diagnostics
	conn := &GuardrailsConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty GuardrailsConnection")

	// Case 2: Validate a populated GuardrailsConnection, should pass with no diagnostics
	accessKey := "access_key_value"
	secretKey := "secret_key_value"
	workspace := "workspace_value"

	conn = &GuardrailsConnection{
		AccessKey: &accessKey,
		SecretKey: &secretKey,
		Workspace: &workspace,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated GuardrailsConnection")
}

// ------------------------------------------------------------
// Turbot Pipes
// ------------------------------------------------------------

func TestPipesDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	pipesConnection := PipesConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("PIPES_TOKEN")
	newConnection, err := pipesConnection.Resolve(context.TODO())
	assert.Nil(err)

	newPipesConnection := newConnection.(*PipesConnection)
	assert.Equal("", *newPipesConnection.Token)

	os.Setenv("PIPES_TOKEN", "tpt_cld630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb")

	newConnection, err = pipesConnection.Resolve(context.TODO())
	assert.Nil(err)

	newPipesConnection = newConnection.(*PipesConnection)
	assert.Equal("tpt_cld630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb", *newPipesConnection.Token)
}

func TestPipesConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *PipesConnection
	var conn2 *PipesConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &PipesConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Token
	token := "token_value"

	conn1 = &PipesConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	conn2 = &PipesConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Token: &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same Token and should be equal")

	// Case 4: Connections have different Tokens
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different Tokens, should return false")
}

func TestPipesConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty PipesConnection, should pass with no diagnostics
	conn := &PipesConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty PipesConnection")

	// Case 2: Validate a populated PipesConnection, should pass with no diagnostics
	token := "token_value"

	conn = &PipesConnection{
		Token: &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated PipesConnection")
}

// ------------------------------------------------------------
// UptimeRobot
// ------------------------------------------------------------

func TestUptimeRobotDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	uptimeRobotConnection := UptimeRobotConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("UPTIMEROBOT_API_KEY")
	newConnection, err := uptimeRobotConnection.Resolve(context.TODO())
	assert.Nil(err)

	newUptimeRobotConnection := newConnection.(*UptimeRobotConnection)
	assert.Equal("", *newUptimeRobotConnection.APIKey)

	os.Setenv("UPTIMEROBOT_API_KEY", "u1123455-ecaf32fwer633fdf4f33dd3c445")

	newConnection, err = uptimeRobotConnection.Resolve(context.TODO())
	assert.Nil(err)

	newUptimeRobotConnection = newConnection.(*UptimeRobotConnection)
	assert.Equal("u1123455-ecaf32fwer633fdf4f33dd3c445", *newUptimeRobotConnection.APIKey)
}

func TestUptimeRobotConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *UptimeRobotConnection
	var conn2 *UptimeRobotConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &UptimeRobotConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec: G101

	conn1 = &UptimeRobotConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	conn2 = &UptimeRobotConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")
}

func TestUptimeRobotConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty UptimeRobotConnection, should pass with no diagnostics
	conn := &UptimeRobotConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty UptimeRobotConnection")

	// Case 2: Validate a populated UptimeRobotConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101

	conn = &UptimeRobotConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated UptimeRobotConnection")
}

// ------------------------------------------------------------
// URLScan
// ------------------------------------------------------------

func TestUrlscanDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	urlscanConnection := UrlscanConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("URLSCAN_API_KEY")
	newConnection, err := urlscanConnection.Resolve(context.TODO())
	assert.Nil(err)

	newUrlscanConnection := newConnection.(*UrlscanConnection)
	assert.Equal("", *newUrlscanConnection.APIKey)

	os.Setenv("URLSCAN_API_KEY", "4d7e9123-e127-56c1-8d6a-59cad2f12abc")

	newConnection, err = urlscanConnection.Resolve(context.TODO())
	assert.Nil(err)

	newUrlscanConnection = newConnection.(*UrlscanConnection)
	assert.Equal("4d7e9123-e127-56c1-8d6a-59cad2f12abc", *newUrlscanConnection.APIKey)
}

func TestUrlscanConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *UrlscanConnection
	var conn2 *UrlscanConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &UrlscanConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec: G101

	conn1 = &UrlscanConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	conn2 = &UrlscanConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")
}

func TestUrlscanConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty UrlscanConnection, should pass with no diagnostics
	conn := &UrlscanConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty UrlscanConnection")

	// Case 2: Validate a populated UrlscanConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101

	conn = &UrlscanConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated UrlscanConnection")
}

// ------------------------------------------------------------
// Vault
// ------------------------------------------------------------

func TestVaultDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	vaultConnection := VaultConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("VAULT_TOKEN")
	os.Unsetenv("VAULT_ADDR")

	newConnection, err := vaultConnection.Resolve(context.TODO())
	assert.Nil(err)

	newVaultConnection := newConnection.(*VaultConnection)
	assert.Equal("", *newVaultConnection.Token)
	assert.Equal("", *newVaultConnection.Address)

	os.Setenv("VAULT_TOKEN", "hsv-fhhwskfkwh")
	os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")

	newConnection, err = vaultConnection.Resolve(context.TODO())
	assert.Nil(err)

	newVaultConnection = newConnection.(*VaultConnection)
	assert.Equal("hsv-fhhwskfkwh", *newVaultConnection.Token)
	assert.Equal("http://127.0.0.1:8200", *newVaultConnection.Address)
}

func TestVaultConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *VaultConnection
	var conn2 *VaultConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &VaultConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Address and Token
	address := "https://vault.example.com"
	token := "vault_token_value" // #nosec: G101

	conn1 = &VaultConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Address: &address,
		Token:   &token,
	}

	conn2 = &VaultConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Address: &address,
		Token:   &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same Address and Token and should be equal")

	// Case 4: Connections have different Addresses
	differentAddress := "https://different-vault.example.com"
	conn2.Address = &differentAddress
	assert.False(conn1.Equals(conn2), "Connections have different Addresses, should return false")

	// Case 5: Connections have different Tokens
	conn2.Address = &address // Reset Address to match conn1
	differentToken := "different_vault_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different Tokens, should return false")
}

func TestVaultConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty VaultConnection, should pass with no diagnostics
	conn := &VaultConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty VaultConnection")

	// Case 2: Validate a populated VaultConnection, should pass with no diagnostics
	address := "https://vault.example.com"
	token := "vault_token_value" // #nosec: G101

	conn = &VaultConnection{
		Address: &address,
		Token:   &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated VaultConnection")
}

// ------------------------------------------------------------
// VirusTotal
// ------------------------------------------------------------

func TestVirusTotalDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	virusTotalConnection := VirusTotalConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("VTCLI_APIKEY")
	newCollection, err := virusTotalConnection.Resolve(context.TODO())
	assert.Nil(err)

	newVirusTotalCollection := newCollection.(*VirusTotalConnection)
	assert.Equal("", *newVirusTotalCollection.APIKey)

	os.Setenv("VTCLI_APIKEY", "w5kukcma7yfj8m8p5rkjx5chg3nno9z7h7wr4o8uq1n0pmr5dfejox4oz4xr7g3c")

	newCollection, err = virusTotalConnection.Resolve(context.TODO())
	assert.Nil(err)

	newVirusTotalCollection = newCollection.(*VirusTotalConnection)
	assert.Equal("w5kukcma7yfj8m8p5rkjx5chg3nno9z7h7wr4o8uq1n0pmr5dfejox4oz4xr7g3c", *newVirusTotalCollection.APIKey)
}

func TestVirusTotalConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *VirusTotalConnection
	var conn2 *VirusTotalConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &VirusTotalConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same APIKey
	apiKey := "api_key_value" // #nosec: G101

	conn1 = &VirusTotalConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	conn2 = &VirusTotalConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		APIKey: &apiKey,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same APIKey and should be equal")

	// Case 4: Connections have different APIKeys
	differentAPIKey := "different_api_key_value"
	conn2.APIKey = &differentAPIKey
	assert.False(conn1.Equals(conn2), "Connections have different APIKeys, should return false")
}

func TestVirusTotalConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty VirusTotalConnection, should pass with no diagnostics
	conn := &VirusTotalConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty VirusTotalConnection")

	// Case 2: Validate a populated VirusTotalConnection, should pass with no diagnostics
	apiKey := "api_key_value" // #nosec: G101

	conn = &VirusTotalConnection{
		APIKey: &apiKey,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated VirusTotalConnection")
}

// ------------------------------------------------------------
// Zendesk
// ------------------------------------------------------------

func TestZendeskDefaultConnection(t *testing.T) {
	assert := assert.New(t)

	zendeskConnection := ZendeskConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("ZENDESK_SUBDOMAIN")
	os.Unsetenv("ZENDESK_EMAIL")
	os.Unsetenv("ZENDESK_API_TOKEN")

	newConnection, err := zendeskConnection.Resolve(context.TODO())
	assert.Nil(err)

	newZendeskConnection := newConnection.(*ZendeskConnection)
	assert.Equal("", *newZendeskConnection.Subdomain)
	assert.Equal("", *newZendeskConnection.Email)
	assert.Equal("", *newZendeskConnection.Token)

	os.Setenv("ZENDESK_SUBDOMAIN", "dmi")
	os.Setenv("ZENDESK_EMAIL", "pam@dmi.com")
	os.Setenv("ZENDESK_API_TOKEN", "17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4")

	newConnection, err = zendeskConnection.Resolve(context.TODO())
	assert.Nil(err)

	newZendeskConnection = newConnection.(*ZendeskConnection)
	assert.Equal("dmi", *newZendeskConnection.Subdomain)
	assert.Equal("pam@dmi.com", *newZendeskConnection.Email)
	assert.Equal("17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4", *newZendeskConnection.Token)
}

func TestZendeskConnectionEquals(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Both connections are nil
	var conn1 *ZendeskConnection
	var conn2 *ZendeskConnection
	assert.True(conn1.Equals(conn2), "Both connections should be nil and equal")

	// Case 2: One connection is nil
	conn1 = &ZendeskConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
	}
	assert.False(conn1.Equals(nil), "One connection is nil, should return false")

	// Case 3: Both connections have the same Email, Subdomain, and Token
	email := "user@example.com"
	subdomain := "mycompany"
	token := "token_value"

	conn1 = &ZendeskConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Email:     &email,
		Subdomain: &subdomain,
		Token:     &token,
	}

	conn2 = &ZendeskConnection{
		ConnectionImpl: ConnectionImpl{
			ShortName: "default",
		},
		Email:     &email,
		Subdomain: &subdomain,
		Token:     &token,
	}

	assert.True(conn1.Equals(conn2), "Both connections have the same Email, Subdomain, and Token and should be equal")

	// Case 4: Connections have different Emails
	differentEmail := "different@example.com"
	conn2.Email = &differentEmail
	assert.False(conn1.Equals(conn2), "Connections have different Emails, should return false")

	// Case 5: Connections have different Subdomains
	conn2.Email = &email // Reset Email to match conn1
	differentSubdomain := "differentcompany"
	conn2.Subdomain = &differentSubdomain
	assert.False(conn1.Equals(conn2), "Connections have different Subdomains, should return false")

	// Case 6: Connections have different Tokens
	conn2.Subdomain = &subdomain // Reset Subdomain to match conn1
	differentToken := "different_token_value"
	conn2.Token = &differentToken
	assert.False(conn1.Equals(conn2), "Connections have different Tokens, should return false")
}

func TestZendeskConnectionValidate(t *testing.T) {
	assert := assert.New(t)

	// Case 1: Validate an empty ZendeskConnection, should pass with no diagnostics
	conn := &ZendeskConnection{}
	diagnostics := conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for an empty ZendeskConnection")

	// Case 2: Validate a populated ZendeskConnection, should pass with no diagnostics
	email := "user@example.com"
	subdomain := "mycompany"
	token := "token_value"

	conn = &ZendeskConnection{
		Email:     &email,
		Subdomain: &subdomain,
		Token:     &token,
	}
	diagnostics = conn.Validate()
	assert.Len(diagnostics, 0, "Validation should pass with no diagnostics for a populated ZendeskConnection")
}
