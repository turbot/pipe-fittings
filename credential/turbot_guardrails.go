package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type GuardrailsCredential struct {
	CredentialImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	Workspace *string `json:"workspace,omitempty" cty:"workspace" hcl:"workspace,optional"`
}

func (c *GuardrailsCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessKey != nil {
		env["TURBOT_ACCESS_KEY"] = cty.StringVal(*c.AccessKey)
	}
	if c.SecretKey != nil {
		env["TURBOT_SECRET_KEY"] = cty.StringVal(*c.SecretKey)
	}
	if c.Workspace != nil {
		env["TURBOT_WORKSPACE"] = cty.StringVal(*c.Workspace)
	}
	return env
}

func (c *GuardrailsCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GuardrailsCredential) Equals(other *GuardrailsCredential) bool {
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

	if !utils.StringPtrEqual(c.SecretKey, other.SecretKey) {
		return false
	}

	if !utils.StringPtrEqual(c.Workspace, other.Workspace) {
		return false
	}

	return true
}

func (c *GuardrailsCredential) Resolve(ctx context.Context) (Credential, error) {
	guardrailsAccessKeyEnvVar := os.Getenv("TURBOT_ACCESS_KEY")
	guardrailsSecretKeyEnvVar := os.Getenv("TURBOT_SECRET_KEY")
	guardrailsWorkspaceEnvVar := os.Getenv("TURBOT_WORKSPACE")

	// Don't modify existing credential, resolve to a new one
	newCreds := &GuardrailsCredential{
		CredentialImpl: c.CredentialImpl,
		Workspace:      c.Workspace,
	}

	if c.AccessKey == nil {
		newCreds.AccessKey = &guardrailsAccessKeyEnvVar
	} else {
		newCreds.AccessKey = c.AccessKey
	}

	if c.SecretKey == nil {
		newCreds.SecretKey = &guardrailsSecretKeyEnvVar
	} else {
		newCreds.SecretKey = c.SecretKey
	}

	if c.Workspace == nil {
		newCreds.Workspace = &guardrailsWorkspaceEnvVar
	} else {
		newCreds.Workspace = c.Workspace
	}

	return newCreds, nil
}

func (c *GuardrailsCredential) GetTtl() int {
	return -1
}

func (c *GuardrailsCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type GuardrailsConnectionConfig struct {
	AccessKey *string `cty:"access_key" hcl:"access_key,optional"`
	Profile   *string `cty:"profile" hcl:"profile,optional"`
	SecretKey *string `cty:"secret_key" hcl:"secret_key,optional"`
	Workspace *string `cty:"workspace" hcl:"workspace,optional"`
}

func (c *GuardrailsConnectionConfig) GetCredential(name string) Credential {

	guardrailsCred := &GuardrailsCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
		Workspace: c.Workspace,
	}

	return guardrailsCred
}
