package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type GuardrailsConnection struct {
	modconfig.ConnectionImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	Workspace *string `json:"workspace,omitempty" cty:"workspace" hcl:"workspace,optional"`
}

func (c *GuardrailsConnection) GetConnectionType() string {
	return "guardrails"
}

func (c *GuardrailsConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
	guardrailsAccessKeyEnvVar := os.Getenv("TURBOT_ACCESS_KEY")
	guardrailsSecretKeyEnvVar := os.Getenv("TURBOT_SECRET_KEY")
	guardrailsWorkspaceEnvVar := os.Getenv("TURBOT_WORKSPACE")

	// Don't modify existing connection, resolve to a new one
	newConnection := &GuardrailsConnection{
		ConnectionImpl: c.ConnectionImpl,
		Workspace:      c.Workspace,
	}

	if c.AccessKey == nil {
		newConnection.AccessKey = &guardrailsAccessKeyEnvVar
	} else {
		newConnection.AccessKey = c.AccessKey
	}

	if c.SecretKey == nil {
		newConnection.SecretKey = &guardrailsSecretKeyEnvVar
	} else {
		newConnection.SecretKey = c.SecretKey
	}

	if c.Workspace == nil {
		newConnection.Workspace = &guardrailsWorkspaceEnvVar
	} else {
		newConnection.Workspace = c.Workspace
	}

	return newConnection, nil
}

func (c *GuardrailsConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*GuardrailsConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AccessKey, other.AccessKey) {
		return false
	}

	if !utils.PtrEqual(c.SecretKey, other.SecretKey) {
		return false
	}

	if !utils.PtrEqual(c.Workspace, other.Workspace) {
		return false
	}

	return true
}

func (c *GuardrailsConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *GuardrailsConnection) GetTtl() int {
	return -1
}

func (c *GuardrailsConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GuardrailsConnection) GetEnv() map[string]cty.Value {
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
