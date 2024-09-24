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

const OktaConnectionType = "okta"

type OktaConnection struct {
	ConnectionImpl

	Domain *string `json:"domain,omitempty" cty:"domain" hcl:"domain,optional"`
	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewOktaConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &OktaConnection{
		ConnectionImpl: NewConnectionImpl(OktaConnectionType, shortName, declRange),
	}
}
func (c *OktaConnection) GetConnectionType() string {
	return OktaConnectionType
}

func (c *OktaConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

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

	impl := c.GetConnectionImpl()
	if impl.Equals(otherConnection.GetConnectionImpl()) == false {
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
	if c.Pipes != nil && (c.Token != nil || c.Domain != nil) {
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

func (c *OktaConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OktaConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["OKTA_CLIENT_TOKEN"] = cty.StringVal(*c.Token)
	}
	if c.Domain != nil {
		env["OKTA_ORGURL"] = cty.StringVal(*c.Domain)
	}
	return env
}
