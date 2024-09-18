package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type AlicloudConnection struct {
	ConnectionImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
}

func (c *AlicloudConnection) GetConnectionType() string {
	return "alicloud"
}

func (c *AlicloudConnection) Resolve(ctx context.Context) (PipelingConnection, error) {

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

	return true
}

func (c *AlicloudConnection) Validate() hcl.Diagnostics {

	if c.AccessKey != nil && c.SecretKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "access_key defined without secret_key",
				Subject:  &c.DeclRange,
			},
		}
	}

	if c.SecretKey != nil && c.AccessKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "secret_key defined without access_key",
				Subject:  &c.DeclRange,
			},
		}
	}

	return hcl.Diagnostics{}
}

// in seconds
func (c *AlicloudConnection) GetTtl() int {
	return -1
}

func (c *AlicloudConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
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
