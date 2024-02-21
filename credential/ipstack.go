package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type IPstackCredential struct {
	CredentialImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
}

func (c *IPstackCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessKey != nil {
		env["IPSTACK_ACCESS_KEY"] = cty.StringVal(*c.AccessKey)
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

func (c *IPstackCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.AccessKey == nil {
		ipstackAccessKeyEnvVar := os.Getenv("IPSTACK_ACCESS_KEY")

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

// TODO: Check the connection config required attributes
// Steampipe uses Token, but the Flowpipe credential uses AccessKey

// type IPStackConnectionConfig struct {
// 	AccessKey *string `cty:"access_key" hcl:"access_key"`
// }

// func (c *IPStackConnectionConfig) GetCredential(name string) Credential {

// 	ipstackCred := &IPstackCredential{
// 		CredentialImpl: CredentialImpl{
// 			HclResourceImpl: modconfig.HclResourceImpl{
// 				FullName:        name,
// 				ShortName:       name,
// 				UnqualifiedName: name,
// 			},
// 		},

// 		AccessKey: c.AccessKey,
// 	}

// 	return ipstackCred
// }
