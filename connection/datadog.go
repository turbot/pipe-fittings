package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type DatadogConnection struct {
	modconfig.ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	AppKey *string `json:"app_key,omitempty" cty:"app_key" hcl:"app_key,optional"`
	APIUrl *string `json:"api_url,omitempty" cty:"api_url" hcl:"api_url,optional"`
}

func (c *DatadogConnection) GetConnectionType() string {
	return "datadog"
}

func (c *DatadogConnection) Resolve(ctx context.Context) (modconfig.PipelingConnection, error) {
	datadogAPIKeyEnvVar := os.Getenv("DD_CLIENT_API_KEY")
	datadogAppKeyEnvVar := os.Getenv("DD_CLIENT_APP_KEY")

	// Don't modify existing connection, resolve to a new one
	newConnection := &DatadogConnection{
		ConnectionImpl: c.ConnectionImpl,
		APIUrl:         c.APIUrl,
	}

	if c.APIKey == nil {
		newConnection.APIKey = &datadogAPIKeyEnvVar
	} else {
		newConnection.APIKey = c.APIKey
	}

	if c.AppKey == nil {
		newConnection.AppKey = &datadogAppKeyEnvVar
	} else {
		newConnection.AppKey = c.AppKey
	}

	return newConnection, nil
}

func (c *DatadogConnection) Equals(otherConnection modconfig.PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*DatadogConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	if !utils.PtrEqual(c.AppKey, other.AppKey) {
		return false
	}

	if !utils.PtrEqual(c.APIUrl, other.APIUrl) {
		return false
	}

	return true
}

func (c *DatadogConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *DatadogConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *DatadogConnection) GetTtl() int {
	return -1
}

func (c *DatadogConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["DD_CLIENT_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	if c.AppKey != nil {
		env["DD_CLIENT_APP_KEY"] = cty.StringVal(*c.AppKey)
	}
	return env
}
