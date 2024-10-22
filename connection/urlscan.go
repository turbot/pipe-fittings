package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const UrlscanConnectionType = "urlscan"

type UrlscanConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func NewUrlscanConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &UrlscanConnection{
		ConnectionImpl: NewConnectionImpl(UrlscanConnectionType, shortName, declRange),
	}
}
func (c *UrlscanConnection) GetConnectionType() string {
	return UrlscanConnectionType
}

func (c *UrlscanConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &UrlscanConnection{ConnectionImpl: c.ConnectionImpl})
	}

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

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *UrlscanConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.APIKey != nil) {
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

func (c *UrlscanConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *UrlscanConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["URLSCAN_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}
