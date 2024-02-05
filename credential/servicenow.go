package credential

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type ServiceNowCredential struct {
	CredentialImpl

	InstanceURL *string `json:"instance_url,omitempty" cty:"instance_url" hcl:"instance_url,optional"`
	Username    *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password    *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func (c *ServiceNowCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.InstanceURL != nil {
		env["SERVICENOW_INSTANCE_URL"] = cty.StringVal(*c.InstanceURL)
	}
	if c.Username != nil {
		env["SERVICENOW_USERNAME"] = cty.StringVal(*c.Username)
	}
	if c.Password != nil {
		env["SERVICENOW_PASSWORD"] = cty.StringVal(*c.Password)
	}
	return env
}

func (c *ServiceNowCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ServiceNowCredential) Resolve(ctx context.Context) (Credential, error) {
	servicenowInstanceURLEnvVar := os.Getenv("SERVICENOW_INSTANCE_URL")
	servicenowUsernameEnvVar := os.Getenv("SERVICENOW_USERNAME")
	servicenowPasswordEnvVar := os.Getenv("SERVICENOW_PASSWORD")

	// Don't modify existing credential, resolve to a new one
	newCreds := &ServiceNowCredential{
		CredentialImpl: c.CredentialImpl,
	}

	if c.InstanceURL == nil {
		newCreds.InstanceURL = &servicenowInstanceURLEnvVar
	} else {
		newCreds.InstanceURL = c.InstanceURL
	}

	if c.Username == nil {
		newCreds.Username = &servicenowUsernameEnvVar
	} else {
		newCreds.Username = c.Username
	}

	if c.Password == nil {
		newCreds.Password = &servicenowPasswordEnvVar
	} else {
		newCreds.Password = c.Password
	}

	return newCreds, nil
}

func (c *ServiceNowCredential) GetTtl() int {
	return -1
}

func (c *ServiceNowCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}
