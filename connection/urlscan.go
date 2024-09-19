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

const UrlscanConnectionType = "urlscan"

type UrlscanConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func NewUrlscanConnection(block *hcl.Block) PipelingConnection {
	return &UrlscanConnection{
		ConnectionImpl: NewConnectionImpl(block),
	}
}

func (c *UrlscanConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.APIKey == nil {
		urlscanAPIKeyEnvVar := os.Getenv("URLSCAN_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &UrlscanConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &urlscanAPIKeyEnvVar,
		}
		return newConnection, nil
	}

	return c, nil
}

func (c *UrlscanConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*UrlscanConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *UrlscanConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *UrlscanConnection) GetTtl() int {
	return -1
}

func (c *UrlscanConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *UrlscanConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["URLSCAN_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}
