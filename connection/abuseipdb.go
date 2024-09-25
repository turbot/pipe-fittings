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

const AbuseIPDBConnectionType = "abuseipdb"

type AbuseIPDBConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func NewAbuseIPDBConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &AbuseIPDBConnection{
		ConnectionImpl: NewConnectionImpl(AbuseIPDBConnectionType, shortName, declRange),
	}
}
func (c *AbuseIPDBConnection) GetConnectionType() string {
	return AbuseIPDBConnectionType
}

func (c *AbuseIPDBConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AbuseIPDBConnection{})
	}

	if c.APIKey == nil {
		abuseIPDBAPIKeyEnvVar := os.Getenv("ABUSEIPDB_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &AbuseIPDBConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &abuseIPDBAPIKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *AbuseIPDBConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*AbuseIPDBConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *AbuseIPDBConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.APIKey != nil) {
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

func (c *AbuseIPDBConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *AbuseIPDBConnection) GetEnv() map[string]cty.Value {
	// There is no environment variable listed in the AbuseIPDB official API docs
	// https://www.abuseipdb.com/api.html
	return nil
}
