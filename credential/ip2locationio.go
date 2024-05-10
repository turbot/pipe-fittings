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

type IP2LocationIOCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *IP2LocationIOCredential) getEnv() map[string]cty.Value {
	// There is no environment variable listed in the IP2LocationIO official API docs
	// https://www.ip2location.io/ip2location-documentation
	return nil
}

func (c *IP2LocationIOCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *IP2LocationIOCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*IP2LocationIOCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *IP2LocationIOCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		ip2locationAPIKeyEnvVar := os.Getenv("IP2LOCATIONIO_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &IP2LocationIOCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &ip2locationAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *IP2LocationIOCredential) GetTtl() int {
	return -1
}

func (c *IP2LocationIOCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type IP2LocationIOConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
}

func (c *IP2LocationIOConnectionConfig) GetCredential(name string, shortName string) Credential {

	ip2LocationIOCred := &IP2LocationIOCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "ip2locationio",
		},

		APIKey: c.APIKey,
	}

	return ip2LocationIOCred
}
