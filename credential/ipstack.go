package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type IPstackCredential struct {
	CredentialImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
}

func (c *IPstackCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}

	// The IPstack access key can be configured either of these environment variables
	// IPSTACK_ACCESS_KEY and IPSTACK_TOKEN
	if c.AccessKey != nil {
		env["IPSTACK_ACCESS_KEY"] = cty.StringVal(*c.AccessKey)
		env["IPSTACK_TOKEN"] = cty.StringVal(*c.AccessKey)
	}
	return env
}

func (c *IPstackCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IPstackCredential) Equals(other *IPstackCredential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && other == nil {
		return true
	}

	if (c == nil && other != nil) || (c != nil && other == nil) {
		return false
	}

	if !utils.StringPtrEqual(c.AccessKey, other.AccessKey) {
		return false
	}

	return true
}

func (c *IPstackCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.AccessKey == nil {
		// The order of precedence for the IPstack access key environment variable
		// 1. IPSTACK_ACCESS_KEY
		// 2. IPSTACK_TOKEN

		ipstackAccessKeyEnvVar := os.Getenv("IPSTACK_TOKEN")
		if os.Getenv("IPSTACK_ACCESS_KEY") != "" {
			ipstackAccessKeyEnvVar = os.Getenv("IPSTACK_ACCESS_KEY")
		}

		// Don't modify existing credential, resolve to a new one
		newCreds := &IPstackCredential{
			CredentialImpl: c.CredentialImpl,
			AccessKey:      &ipstackAccessKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *IPstackCredential) GetTtl() int {
	return -1
}

func (c *IPstackCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type IPStackConnectionConfig struct {
	AccessKey *string `cty:"access_key" hcl:"access_key,optional"`
	Https     *bool   `cty:"https" hcl:"https,optional"`
	Security  *bool   `cty:"security" hcl:"security,optional"`
	Token     *string `cty:"token" hcl:"token,optional"`
}

func (c *IPStackConnectionConfig) GetCredential(name string) Credential {

	// Steampipe uses the attribute 'token' to configure the credential; whereas
	// Flowpipe uses the attribute 'access_key' which is documented in the
	// IPStack's official documentation: https://ipstack.com/documentation
	// Since, Steampipe is still using the attribute 'token' we need a special handling in Flowpipe
	// to support both 'access_key' and 'token', and the order of precedence will be 'access_key' and 'token'
	var ipstackAccessKey string
	if c.Token != nil {
		ipstackAccessKey = *c.Token
	}
	if c.AccessKey != nil {
		ipstackAccessKey = *c.AccessKey
	}

	ipstackCred := &IPstackCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		AccessKey: &ipstackAccessKey,
	}

	return ipstackCred
}
