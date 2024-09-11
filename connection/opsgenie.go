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

type OpsgenieConnection struct {
	ConnectionImpl

	AlertAPIKey    *string `json:"alert_api_key,omitempty" cty:"alert_api_key" hcl:"alert_api_key,optional"`
	IncidentAPIKey *string `json:"incident_api_key,omitempty" cty:"incident_api_key" hcl:"incident_api_key,optional"`
}

func (c *OpsgenieConnection) GetConnectionType() string {
	return "opsgenie"
}

func (c *OpsgenieConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.AlertAPIKey == nil && c.IncidentAPIKey == nil {
		alertAPIKeyEnvVar := os.Getenv("OPSGENIE_ALERT_API_KEY")
		incidentAPIKeyEnvVar := os.Getenv("OPSGENIE_INCIDENT_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &OpsgenieConnection{
			ConnectionImpl: c.ConnectionImpl,
			AlertAPIKey:    &alertAPIKeyEnvVar,
			IncidentAPIKey: &incidentAPIKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *OpsgenieConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*OpsgenieConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AlertAPIKey, other.AlertAPIKey) {
		return false
	}

	if !utils.PtrEqual(c.IncidentAPIKey, other.IncidentAPIKey) {
		return false
	}

	return true
}

func (c *OpsgenieConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *OpsgenieConnection) GetTtl() int {
	return -1
}

func (c *OpsgenieConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OpsgenieConnection) getEnv() map[string]cty.Value {
	return nil
}
