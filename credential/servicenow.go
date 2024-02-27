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

func (c *ServiceNowCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*ServiceNowCredential)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.InstanceURL, other.InstanceURL) {
		return false
	}

	if !utils.PtrEqual(c.Username, other.Username) {
		return false
	}

	if !utils.PtrEqual(c.Password, other.Password) {
		return false
	}

	return true
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

type ServiceNowConnectionConfig struct {
	ClientID     *string `cty:"client_id" hcl:"client_id,optional"`
	ClientSecret *string `cty:"client_secret" hcl:"client_secret,optional"`
	InstanceURL  *string `cty:"instance_url" hcl:"instance_url"`
	Objects      *string `cty:"objects" hcl:"objects,optional"`
	Password     *string `cty:"password" hcl:"password"`
	Username     *string `cty:"username" hcl:"username"`
}

func (c *ServiceNowConnectionConfig) GetCredential(name string, shortName string) Credential {

	serviceNowCred := &ServiceNowCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       shortName,
				UnqualifiedName: name,
			},
			Type: "servicenow",
		},

		InstanceURL: c.InstanceURL,
		Username:    c.Username,
		Password:    c.Password,
	}

	return serviceNowCred
}
