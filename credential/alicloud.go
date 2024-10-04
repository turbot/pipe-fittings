package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type AlicloudCredential struct {
	CredentialImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
}

func (c *AlicloudCredential) Validate() hcl.Diagnostics {

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

func (c *AlicloudCredential) Resolve(ctx context.Context) (Credential, error) {

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

	// Don't modify existing credential, resolve to a new one
	newCreds := &AlicloudCredential{
		CredentialImpl: c.CredentialImpl,
	}

	if c.AccessKey == nil {
		newCreds.AccessKey = &alicloudAccessKeyEnvVar
	} else {
		newCreds.AccessKey = c.AccessKey
	}

	if c.SecretKey == nil {
		newCreds.SecretKey = &alicloudSecretKeyEnvVar
	} else {
		newCreds.SecretKey = c.SecretKey
	}

	return newCreds, nil
}

// in seconds
func (c *AlicloudCredential) GetTtl() int {
	return -1
}

func (c *AlicloudCredential) getEnv() map[string]cty.Value {
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

func (c *AlicloudCredential) CtyValue() (cty.Value, error) {
	return ctyValueForCredential(c)
}

func (c *AlicloudCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*AlicloudCredential)
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

type AlicloudConnectionConfig struct {
	AccessKey        *string  `cty:"access_key" hcl:"access_key,optional"`
	IgnoreErrorCodes []string `cty:"ignore_error_codes" hcl:"ignore_error_codes,optional"`
	Regions          []string `cty:"regions" hcl:"regions,optional"`
	SecretKey        *string  `cty:"secret_key" hcl:"secret_key,optional"`
}

func (c *AlicloudConnectionConfig) GetCredential(name string, shortName string) Credential {

	alicloudCred := &AlicloudCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "alicloud",
		},

		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
	}

	return alicloudCred
}
