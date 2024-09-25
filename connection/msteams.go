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

const MicrosoftTeamsConnectionType = "microsoft_teams"

type MicrosoftTeamsConnection struct {
	ConnectionImpl

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func NewMicrosoftTeamsConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &MicrosoftTeamsConnection{
		ConnectionImpl: NewConnectionImpl(MicrosoftTeamsConnectionType, shortName, declRange),
	}
}
func (c *MicrosoftTeamsConnection) GetConnectionType() string {
	return MicrosoftTeamsConnectionType
}

func (c *MicrosoftTeamsConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &MicrosoftTeamsConnection{})
	}

	if c.AccessToken == nil {
		msTeamsAccessTokenEnvVar := os.Getenv("TEAMS_ACCESS_TOKEN")

		// Don't modify existing connection, resolve to a new one
		newConnection := &MicrosoftTeamsConnection{
			ConnectionImpl: c.ConnectionImpl,
			AccessToken:    &msTeamsAccessTokenEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *MicrosoftTeamsConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*MicrosoftTeamsConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AccessToken, other.AccessToken) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *MicrosoftTeamsConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.AccessToken != nil) {
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

func (c *MicrosoftTeamsConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *MicrosoftTeamsConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessToken != nil {
		env["TEAMS_ACCESS_TOKEN"] = cty.StringVal(*c.AccessToken)
	}
	return env
}
