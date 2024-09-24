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

const OpsgenieConnectionType = "opsgenie"

type OpsgenieConnection struct {
	ConnectionImpl

	AlertAPIKey    *string `json:"alert_api_key,omitempty" cty:"alert_api_key" hcl:"alert_api_key,optional"`
	IncidentAPIKey *string `json:"incident_api_key,omitempty" cty:"incident_api_key" hcl:"incident_api_key,optional"`
}

func NewOpsgenieConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &OpsgenieConnection{
		ConnectionImpl: NewConnectionImpl(OpsgenieConnectionType, shortName, declRange),
	}
}
func (c *OpsgenieConnection) GetConnectionType() string {
	return OpsgenieConnectionType
}

func (c *OpsgenieConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

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

	impl := c.GetConnectionImpl()
	if impl.Equals(otherConnection.GetConnectionImpl()) == false {
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
	if c.Pipes != nil && (c.AlertAPIKey != nil || c.IncidentAPIKey != nil) {
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

func (c *OpsgenieConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OpsgenieConnection) GetEnv() map[string]cty.Value {
	return nil
}
