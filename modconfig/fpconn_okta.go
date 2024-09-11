package modconfig

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type OktaConnection struct {
	ConnectionImpl

	Domain *string `json:"domain,omitempty" cty:"domain" hcl:"domain,optional"`
	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (c *OktaConnection) GetConnectionType() string {
	return "okta"
}

func (c *OktaConnection) Resolve(ctx context.Context) (PipelingConnection, error) {

	if c.Token == nil && c.Domain == nil {
		apiTokenEnvVar := os.Getenv("OKTA_CLIENT_TOKEN")
		domainEnvVar := os.Getenv("OKTA_ORGURL")

		// Don't modify existing connection, resolve to a new one
		newCreds := &OktaConnection{
			ConnectionImpl: c.ConnectionImpl,
			Token:          &apiTokenEnvVar,
			Domain:         &domainEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *OktaConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*OktaConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Domain, other.Domain) {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *OktaConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *OktaConnection) GetTtl() int {
	return -1
}

func (c *OktaConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OktaConnection) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["OKTA_CLIENT_TOKEN"] = cty.StringVal(*c.Token)
	}
	if c.Domain != nil {
		env["OKTA_ORGURL"] = cty.StringVal(*c.Domain)
	}
	return env
}
