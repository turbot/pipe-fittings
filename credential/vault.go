package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type VaultCredential struct {
	CredentialImpl

	Token   *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	Address *string `json:"address,omitempty" cty:"address" hcl:"address,optional"`
}

func (c *VaultCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["VAULT_TOKEN"] = cty.StringVal(*c.Token)
	}
	if c.Address != nil {
		env["VAULT_ADDR"] = cty.StringVal(*c.Address)
	}
	return env
}

func (c *VaultCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *VaultCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.Token == nil && c.Address == nil {
		tokenEnvVar := os.Getenv("VAULT_TOKEN")
		addressEnvVar := os.Getenv("VAULT_ADDR")

		// Don't modify existing credential, resolve to a new one
		newCreds := &VaultCredential{
			CredentialImpl: c.CredentialImpl,
			Token:          &tokenEnvVar,
			Address:        &addressEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *VaultCredential) GetTtl() int {
	return -1
}

func (c *VaultCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type VaultConnectionConfig struct {
	Address     *string `cty:"address" hcl:"address"`
	AuthType    *string `cty:"auth_type" hcl:"auth_type,optional"`
	AwsProvider *string `cty:"aws_provider" hcl:"aws_provider,optional"`
	AwsRole     *string `cty:"aws_role" hcl:"aws_role,optional"`
	Token       *string `cty:"token" hcl:"token,optional"`
}

func (c *VaultConnectionConfig) GetCredential(name string) Credential {

	vaultCred := &VaultCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		Address: c.Address,
		Token:   c.Token,
	}

	return vaultCred
}
