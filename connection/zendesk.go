package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const ZendeskConnectionType = "zendesk"

type ZendeskConnection struct {
	ConnectionImpl

	Email     *string `json:"email,omitempty" cty:"email" hcl:"email,optional"`
	Subdomain *string `json:"subdomain,omitempty" cty:"subdomain" hcl:"subdomain,optional"`
	Token     *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewZendeskConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &ZendeskConnection{
		ConnectionImpl: NewConnectionImpl(ZendeskConnectionType, shortName, declRange),
	}
}
func (c *ZendeskConnection) GetConnectionType() string {
	return ZendeskConnectionType
}

func (c *ZendeskConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

	if c.Subdomain == nil && c.Email == nil && c.Token == nil {
		subdomainEnvVar := os.Getenv("ZENDESK_SUBDOMAIN")
		emailEnvVar := os.Getenv("ZENDESK_EMAIL")
		tokenEnvVar := os.Getenv("ZENDESK_API_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &ZendeskConnection{
			ConnectionImpl: c.ConnectionImpl,
			Subdomain:      &subdomainEnvVar,
			Email:          &emailEnvVar,
			Token:          &tokenEnvVar,
		}

		return newConnection, nil
	}

	return c, nil
}

func (c *ZendeskConnection) Equals(otherConnection PipelingConnection) bool {
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

	other, ok := otherConnection.(*ZendeskConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Email, other.Email) {
		return false
	}

	if !utils.PtrEqual(c.Subdomain, other.Subdomain) {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *ZendeskConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *ZendeskConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ZendeskConnection) GetEnv() map[string]cty.Value {
	return nil
}
