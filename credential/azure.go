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

type AzureCredential struct {
	CredentialImpl

	ClientID     *string `json:"client_id,omitempty" cty:"client_id" hcl:"client_id,optional"`
	ClientSecret *string `json:"client_secret,omitempty" cty:"client_secret" hcl:"client_secret,optional"`
	TenantID     *string `json:"tenant_id,omitempty" cty:"tenant_id" hcl:"tenant_id,optional"`
	Environment  *string `json:"environment,omitempty" cty:"environment" hcl:"environment,optional"`
}

func (c *AzureCredential) getEnv() map[string]cty.Value {
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

func (c *AzureCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *AzureCredential) Equals(otherCredential Credential) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherCredential) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherCredential)) || (c != nil && helpers.IsNil(otherCredential)) {
		return false
	}

	other, ok := otherCredential.(*AzureCredential)
	if !ok {
		return false
	}

	if !utils.StringPtrEqual(c.ClientID, other.ClientID) {
		return false
	}

	if !utils.StringPtrEqual(c.ClientSecret, other.ClientSecret) {
		return false
	}

	if !utils.StringPtrEqual(c.Environment, other.Environment) {
		return false
	}

	if !utils.StringPtrEqual(c.TenantID, other.TenantID) {
		return false
	}

	return true
}

func (c *AzureCredential) Resolve(ctx context.Context) (Credential, error) {

	if c.ClientID == nil && c.ClientSecret == nil && c.TenantID == nil && c.Environment == nil {
		clientIDEnvVar := os.Getenv("AZURE_CLIENT_ID")
		clientSecretEnvVar := os.Getenv("AZURE_CLIENT_SECRET")
		tenantIDEnvVar := os.Getenv("AZURE_TENANT_ID")
		environmentEnvVar := os.Getenv("AZURE_ENVIRONMENT")

		// Don't modify existing credential, resolve to a new one
		newCreds := &AzureCredential{
			CredentialImpl: c.CredentialImpl,
			ClientID:       &clientIDEnvVar,
			ClientSecret:   &clientSecretEnvVar,
			TenantID:       &tenantIDEnvVar,
			Environment:    &environmentEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *AzureCredential) GetTtl() int {
	return -1
}

func (c *AzureCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type AzureConnectionConfig struct {
	CertificatePassword *string  `cty:"certificate_password" hcl:"certificate_password,optional"`
	CertificatePath     *string  `cty:"certificate_path" hcl:"certificate_path,optional"`
	ClientID            *string  `cty:"client_id" hcl:"client_id,optional"`
	ClientSecret        *string  `cty:"client_secret" hcl:"client_secret,optional"`
	Environment         *string  `cty:"environment" hcl:"environment,optional"`
	IgnoreErrorCodes    []string `cty:"ignore_error_codes" hcl:"ignore_error_codes,optional"`
	Password            *string  `cty:"password" hcl:"password,optional"`
	SubscriptionID      *string  `cty:"subscription_id" hcl:"subscription_id,optional"`
	TenantID            *string  `cty:"tenant_id" hcl:"tenant_id,optional"`
	Username            *string  `cty:"username" hcl:"username,optional"`
}

func (c *AzureConnectionConfig) GetCredential(name string) Credential {

	azureCred := &AzureCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				FullName:        name,
				ShortName:       name,
				UnqualifiedName: name,
			},
		},

		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Environment:  c.Environment,
		TenantID:     c.TenantID,
	}

	return azureCred
}
