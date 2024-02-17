package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
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

	alicloudAccessKeyEnvVar := os.Getenv("ALIBABACLOUD_ACCESS_KEY_ID")
	alicloudSecretKeyEnvVar := os.Getenv("ALIBABACLOUD_ACCESS_KEY_SECRET")

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
	if c.AccessKey != nil {
		env["ALIBABACLOUD_ACCESS_KEY_ID"] = cty.StringVal(*c.AccessKey)
	}
	if c.SecretKey != nil {
		env["ALIBABACLOUD_ACCESS_KEY_SECRET"] = cty.StringVal(*c.SecretKey)
	}
	return env
}

func (c *AlicloudCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

type AlicloudConnectionConfig struct {
	AccessKey        *string  `cty:"access_key" hcl:"access_key,optional"`
	IgnoreErrorCodes []string `cty:"ignore_error_codes" hcl:"ignore_error_codes,optional"`
	Regions          []string `cty:"regions" hcl:"regions,optional"`
	SecretKey        *string  `cty:"secret_key" hcl:"secret_key,optional"`
}

func (c *AlicloudConnectionConfig) GetCredential(name string) Credential {

	alicloudCred := &AlicloudCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
	}

	return alicloudCred
}
