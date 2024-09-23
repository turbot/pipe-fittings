package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type UptimeRobotCredential struct {
	CredentialImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *UptimeRobotCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["UPTIMEROBOT_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}

func (c *UptimeRobotCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *UptimeRobotCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*UptimeRobotCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *UptimeRobotCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.APIKey == nil {
		uptimeRobotAPIKeyEnvVar := os.Getenv("UPTIMEROBOT_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &UptimeRobotCredential{
			CredentialImpl: c.CredentialImpl,
			APIKey:         &uptimeRobotAPIKeyEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *UptimeRobotCredential) GetTtl() int {
	return -1
}

func (c *UptimeRobotCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type UptimeRobotConnectionConfig struct {
	APIKey *string `cty:"api_key" hcl:"api_key"`
}

func (c *UptimeRobotConnectionConfig) GetCredential(name string, shortName string) Credential {

	uptimeRobotCred := &UptimeRobotCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "uptimerobot",
		},

		APIKey: c.APIKey,
	}

	return uptimeRobotCred
}
