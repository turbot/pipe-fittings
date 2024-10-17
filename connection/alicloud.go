package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const AlicloudConnectionType = "alicloud"

type AlicloudConnection struct {
	ConnectionImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
}

func NewAlicloudConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &AlicloudConnection{
		ConnectionImpl: NewConnectionImpl(AlicloudConnectionType, shortName, declRange),
	}
}
func (c *AlicloudConnection) GetConnectionType() string {
	return AlicloudConnectionType
}

func (c *AlicloudConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AlicloudConnection{ConnectionImpl: c.ConnectionImpl})
	}

	// The order of precedence for the environment variable
	// 1. ALIBABACLOUD_ACCESS_KEY_ID
	// 2. ALICLOUD_ACCESS_KEY_ID
	// 3. ALICLOUD_ACCESS_KEY
	alicloudAccessKeyEnvVar := os.Getenv("ALICLOUD_ACCESS_KEY")
	if os.Getenv("ALICLOUD_ACCESS_KEY_ID") != "" {
		alicloudAccessKeyEnvVar = os.Getenv("ALICLOUD_ACCESS_KEY_ID")
	}
	if os.Getenv("ALIBABACLOUD_ACCESS_KEY_ID") != "" {
		alicloudAccessKeyEnvVar = os.Getenv("ALIBABACLOUD_ACCESS_KEY_ID")
	}

	// The order of precedence for the environment variable
	// 1. ALIBABACLOUD_ACCESS_KEY_SECRET
	// 2. ALICLOUD_ACCESS_KEY_SECRET
	// 3. ALICLOUD_SECRET_KEY
	alicloudSecretKeyEnvVar := os.Getenv("ALICLOUD_SECRET_KEY")
	if os.Getenv("ALICLOUD_ACCESS_KEY_SECRET") != "" {
		alicloudSecretKeyEnvVar = os.Getenv("ALICLOUD_ACCESS_KEY_SECRET")
	}
	if os.Getenv("ALIBABACLOUD_ACCESS_KEY_SECRET") != "" {
		alicloudSecretKeyEnvVar = os.Getenv("ALIBABACLOUD_ACCESS_KEY_SECRET")
	}

	// Don't modify existing connection, resolve to a new one
	newConnection := &AlicloudConnection{
		ConnectionImpl: c.ConnectionImpl,
	}

	if c.AccessKey == nil {
		newConnection.AccessKey = &alicloudAccessKeyEnvVar
	} else {
		newConnection.AccessKey = c.AccessKey
	}

	if c.SecretKey == nil {
		newConnection.SecretKey = &alicloudSecretKeyEnvVar
	} else {
		newConnection.SecretKey = c.SecretKey
	}

	return newConnection, nil
}

func (c *AlicloudConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*AlicloudConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AccessKey, other.AccessKey) {
		return false
	}

	if !utils.PtrEqual(c.SecretKey, other.SecretKey) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *AlicloudConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.AccessKey != nil || c.SecretKey != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}
	if c.AccessKey != nil && c.SecretKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "access_key defined without secret_key",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}

	if c.SecretKey != nil && c.AccessKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "secret_key defined without access_key",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}

	return hcl.Diagnostics{}
}

func (c *AlicloudConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *AlicloudConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}

	// Alicloud uses 3 different environment variables
	// Hence instead of configuring one, set the value to all the variables
	if c.AccessKey != nil {
		accessKey := cty.StringVal(*c.AccessKey)

		env["ALIBABACLOUD_ACCESS_KEY_ID"] = accessKey
		env["ALICLOUD_ACCESS_KEY_ID"] = accessKey
		env["ALICLOUD_ACCESS_KEY"] = accessKey
	}
	if c.SecretKey != nil {
		secretKey := cty.StringVal(*c.SecretKey)

		env["ALIBABACLOUD_ACCESS_KEY_SECRET"] = secretKey
		env["ALICLOUD_ACCESS_KEY_SECRET"] = secretKey
		env["ALICLOUD_SECRET_KEY"] = secretKey
	}
	return env
}
