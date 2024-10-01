package connection

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const MastodonConnectionType = "mastodon"

type MastodonConnection struct {
	ConnectionImpl

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
	Server      *string `json:"server,omitempty" cty:"server" hcl:"server,optional"`
}

func NewMastodonConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &MastodonConnection{
		ConnectionImpl: NewConnectionImpl(MastodonConnectionType, shortName, declRange),
	}
}
func (c *MastodonConnection) GetConnectionType() string {
	return MastodonConnectionType
}

func (c *MastodonConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &MastodonConnection{})
	}
	return c, nil
}

func (c *MastodonConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*MastodonConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AccessToken, other.AccessToken) {
		return false
	}

	if !utils.PtrEqual(c.Server, other.Server) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *MastodonConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.AccessToken != nil || c.Server != nil) {
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

func (c *MastodonConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *MastodonConnection) GetEnv() map[string]cty.Value {
	// Mastodon has no standard environment variable mentioned anywhere in the docs
	return nil
}
