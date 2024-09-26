package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const AzureConnectionType = "azure"

type AzureConnection struct {
	ConnectionImpl

	ClientID     *string `json:"client_id,omitempty" cty:"client_id" hcl:"client_id,optional"`
	ClientSecret *string `json:"client_secret,omitempty" cty:"client_secret" hcl:"client_secret,optional"`
	TenantID     *string `json:"tenant_id,omitempty" cty:"tenant_id" hcl:"tenant_id,optional"`
	Environment  *string `json:"environment,omitempty" cty:"environment" hcl:"environment,optional"`
}

func NewAzureConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &AzureConnection{
		ConnectionImpl: NewConnectionImpl(AzureConnectionType, shortName, declRange),
	}
}
func (c *AzureConnection) GetConnectionType() string {
	return AzureConnectionType
}

func (c *AzureConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AzureConnection{})
	}

	if c.ClientID == nil && c.ClientSecret == nil && c.TenantID == nil && c.Environment == nil {
		clientIDEnvVar := os.Getenv("AZURE_CLIENT_ID")
		clientSecretEnvVar := os.Getenv("AZURE_CLIENT_SECRET")
		tenantIDEnvVar := os.Getenv("AZURE_TENANT_ID")
		environmentEnvVar := os.Getenv("AZURE_ENVIRONMENT")

		// Don't modify existing connection, resolve to a new one
		newConnection := &AzureConnection{
			ConnectionImpl: c.ConnectionImpl,
			ClientID:       &clientIDEnvVar,
			ClientSecret:   &clientSecretEnvVar,
			TenantID:       &tenantIDEnvVar,
			Environment:    &environmentEnvVar,
		}

		return newConnection, nil
	}

	return c, nil
}

func (c *AzureConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*AzureConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.ClientID, other.ClientID) {
		return false
	}

	if !utils.PtrEqual(c.ClientSecret, other.ClientSecret) {
		return false
	}

	if !utils.PtrEqual(c.Environment, other.Environment) {
		return false
	}

	if !utils.PtrEqual(c.TenantID, other.TenantID) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *AzureConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.ClientID != nil || c.ClientSecret != nil || c.TenantID != nil || c.Environment != nil) {
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

func (c *AzureConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *AzureConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.ClientID != nil {
		env["AZURE_CLIENT_ID"] = cty.StringVal(*c.ClientID)
	}
	if c.ClientSecret != nil {
		env["AZURE_CLIENT_SECRET"] = cty.StringVal(*c.ClientSecret)
	}
	if c.TenantID != nil {
		env["AZURE_TENANT_ID"] = cty.StringVal(*c.TenantID)
	}
	if c.Environment != nil {
		env["AZURE_ENVIRONMENT"] = cty.StringVal(*c.Environment)
	}
	return env
}
