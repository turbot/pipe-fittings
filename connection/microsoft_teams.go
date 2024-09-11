package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type MicrosoftTeamsConnection struct {
	ConnectionImpl

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func (c *MicrosoftTeamsConnection) GetConnectionType() string {
	return "microsoft_teams"
}

func (c *MicrosoftTeamsConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
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

	return true
}

func (c *MicrosoftTeamsConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *MicrosoftTeamsConnection) GetTtl() int {
	return -1
}

func (c *MicrosoftTeamsConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *MicrosoftTeamsConnection) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessToken != nil {
		env["TEAMS_ACCESS_TOKEN"] = cty.StringVal(*c.AccessToken)
	}
	return env
}
