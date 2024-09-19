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

const VaultConnectionType = "vault"

type VaultConnection struct {
	ConnectionImpl

	Address *string `json:"address,omitempty" cty:"address" hcl:"address,optional"`
	Token   *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewVaultConnection(block *hcl.Block) PipelingConnection {
	return &VaultConnection{
		ConnectionImpl: NewConnectionImpl(block),
	}
}

func (c *VaultConnection) Resolve(ctx context.Context) (PipelingConnection, error) {

	if c.Token == nil && c.Address == nil {
		tokenEnvVar := os.Getenv("VAULT_TOKEN")
		addressEnvVar := os.Getenv("VAULT_ADDR")

		// Don't modify existing connection, resolve to a new one
		newConnection := &VaultConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &tokenEnvVar,
			Address:        &addressEnvVar,
		}

		return newConnection, nil
	}

	return c, nil
}

func (c *VaultConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*VaultConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Address, other.Address) {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *VaultConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *VaultConnection) GetTtl() int {
	return -1
}

func (c *VaultConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *VaultConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["VAULT_TOKEN"] = cty.StringVal(*c.Token)
	}
	if c.Address != nil {
		env["VAULT_ADDR"] = cty.StringVal(*c.Address)
	}
	return env
}
