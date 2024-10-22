package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type OpsgenieCredential struct {
	CredentialImpl

	AlertAPIKey    *string `json:"alert_api_key,omitempty" cty:"alert_api_key" hcl:"alert_api_key,optional"`
	IncidentAPIKey *string `json:"incident_api_key,omitempty" cty:"incident_api_key" hcl:"incident_api_key,optional"`
}

func (c *OpsgenieCredential) getEnv() map[string]cty.Value {
	return nil
}

func (c *OpsgenieCredential) CtyValue() (cty.Value, error) {
	return ctyValueForCredential(c)
}

func (c *OpsgenieCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*OpsgenieCredential)
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

func (c *OpsgenieCredential) Resolve(ctx context.Context) (Credential, error) {
	if c.AlertAPIKey == nil && c.IncidentAPIKey == nil {
		alertAPIKeyEnvVar := os.Getenv("OPSGENIE_ALERT_API_KEY")
		incidentAPIKeyEnvVar := os.Getenv("OPSGENIE_INCIDENT_API_KEY")

		// Don't modify existing credential, resolve to a new one
		newCreds := &OpsgenieCredential{
			CredentialImpl: c.CredentialImpl,
			AlertAPIKey:    &alertAPIKeyEnvVar,
			IncidentAPIKey: &incidentAPIKeyEnvVar,
		}

		return newCreds, nil
	}
	return c, nil
}

func (c *OpsgenieCredential) GetTtl() int {
	return -1
}

func (c *OpsgenieCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type OpsgenieConnectionConfig struct {
	AlertAPIKey    *string `cty:"alert_api_key" hcl:"alert_api_key"`
	IncidentAPIKey *string `cty:"incident_api_key" hcl:"incident_api_key"`
}

func (c *OpsgenieConnectionConfig) GetCredential(name string, shortName string) Credential {

	opsgenieCred := &OpsgenieCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "opsgenie",
		},

		AlertAPIKey:    c.AlertAPIKey,
		IncidentAPIKey: c.IncidentAPIKey,
	}

	return opsgenieCred
}
