package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type OpenAIConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
}

func (c *OpenAIConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	if c.APIKey == nil {
		apiKeyEnvVar := os.Getenv("OPENAI_API_KEY")

		// Don't modify existing connection, resolve to a new one
		newConnection := &OpenAIConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &apiKeyEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *OpenAIConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*OpenAIConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	return true
}

func (c *OpenAIConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *OpenAIConnection) GetTtl() int {
	return -1
}

func (c *OpenAIConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *OpenAIConnection) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.APIKey != nil {
		env["OPENAI_API_KEY"] = cty.StringVal(*c.APIKey)
	}
	return env
}
