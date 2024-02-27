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

type JumpCloudCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *JumpCloudCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["JUMPCLOUD_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *JumpCloudCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *JumpCloudCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*JumpCloudCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *JumpCloudCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		apiKeyEnvVar := os.Getenv("JUMPCLOUD_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &JumpCloudCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &apiKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *JumpCloudCredential) GetTtl() int {
	return -1
}

func (c *JumpCloudCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type JumpCloudConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
	OrgID  *string `cty:"org_id" hcl:"org_id,optional"`
}

func (c *JumpCloudConnectionConfig) GetCredential(name string, shortName string) Credential {

	jumpCloudCred := &JumpCloudCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "jumpcloud",
		},

		APIKey: c.APIKey,
	}

	return jumpCloudCred
}
