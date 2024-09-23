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

const ServiceNowConnectionType = "servicenow"

type ServiceNowConnection struct {
	ConnectionImpl

	InstanceURL *string `json:"instance_url,omitempty" cty:"instance_url" hcl:"instance_url,optional"`
	Username    *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password    *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func NewServiceNowConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &ServiceNowConnection{
		ConnectionImpl: NewConnectionImpl(ServiceNowConnectionType, shortName, declRange),
	}
}
func (c *ServiceNowConnection) GetConnectionType() string {
	return ServiceNowConnectionType
}

func (c *ServiceNowConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	servicenowInstanceURLEnvVar := os.Getenv("SERVICENOW_INSTANCE_URL")
	servicenowUsernameEnvVar := os.Getenv("SERVICENOW_USERNAME")
	servicenowPasswordEnvVar := os.Getenv("SERVICENOW_PASSWORD")

	// Don't modify existing connection, resolve to a new one
	newConnection := &ServiceNowConnection{
		ConnectionImpl: c.ConnectionImpl,
	}

	if c.InstanceURL == nil {
		newConnection.InstanceURL = &servicenowInstanceURLEnvVar
	} else {
		newConnection.InstanceURL = c.InstanceURL
	}

	if c.Username == nil {
		newConnection.Username = &servicenowUsernameEnvVar
	} else {
		newConnection.Username = c.Username
	}

	if c.Password == nil {
		newConnection.Password = &servicenowPasswordEnvVar
	} else {
		newConnection.Password = c.Password
	}

	return newConnection, nil
}

func (c *ServiceNowConnection) Equals(otherConnection PipelingConnection) bool {
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

	other, ok := otherConnection.(*ServiceNowConnection)
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

func (c *ServiceNowConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *ServiceNowConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *ServiceNowConnection) GetEnv() map[string]cty.Value {
	return nil
}
