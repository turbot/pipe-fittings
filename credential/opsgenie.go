package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type OpsgenieCredential struct {
	CredentialImpl

	AlertAPIKey    *string `json:"alert_api_key,omitempty" cty:"alert_api_key" hcl:"alert_api_key,optional"`
	IncidentAPIKey *string `json:"incident_api_key,omitempty" cty:"incident_api_key" hcl:"incident_api_key,optional"`
}

func (c *OpsgenieCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AlertAPIKey != nil {
		env["OPSGENIE_ALERT_API_KEY"] = cty.StringVal(*c.AlertAPIKey)
	}
	if c.IncidentAPIKey != nil {
		env["OPSGENIE_INCIDENT_API_KEY"] = cty.StringVal(*c.IncidentAPIKey)
	}
	return env
}

func (c *OpsgenieCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
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

func (c *OpsgenieConnectionConfig) GetCredential(name string) Credential {

	opsgenieCred := &OpsgenieCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		AlertAPIKey:    c.AlertAPIKey,
		IncidentAPIKey: c.IncidentAPIKey,
	}

	return opsgenieCred
}
