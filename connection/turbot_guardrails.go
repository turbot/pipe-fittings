package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const GuardrailsConnectionType = "turbot_guardrails"

type GuardrailsConnection struct {
	ConnectionImpl

	AccessKey *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	Workspace *string `json:"workspace,omitempty" cty:"workspace" hcl:"workspace,optional"`
}

func NewGuardrailsConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &GuardrailsConnection{
		ConnectionImpl: NewConnectionImpl(GuardrailsConnectionType, shortName, declRange),
	}
}
func (c *GuardrailsConnection) GetConnectionType() string {
	return GuardrailsConnectionType
}

func (c *GuardrailsConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

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

func (c *GuardrailsConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	impl := c.GetConnectionImpl()
	if impl.Equals(otherConnection.GetConnectionImpl()) == false {
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
	if c.Pipes != nil && (c.AccessKey != nil || c.SecretKey != nil || c.Workspace != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}
	return hcl.Diagnostics{}
}

func (c *GuardrailsConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
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
