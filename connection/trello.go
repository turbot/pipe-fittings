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

const TrelloConnectionType = "trello"

type TrelloConnection struct {
	ConnectionImpl

	APIKey *string `json:"api_key,omitempty" cty:"api_key" hcl:"api_key,optional"`
	Token  *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func NewTrelloConnection(block *hcl.Block) PipelingConnection {
	return &TrelloConnection{
		ConnectionImpl: NewConnectionImpl(block),
	}
}

func (c *TrelloConnection) Resolve(ctx context.Context) (PipelingConnection, error) {

	if c.APIKey == nil && c.Token == nil {
		apiKeyEnvVar := os.Getenv("TRELLO_API_KEY")
		tokenEnvVar := os.Getenv("TRELLO_TOKEN")
		// Don't modify existing connection, resolve to a new one
		newCreds := &TrelloConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIKey:         &apiKeyEnvVar,
			Token:          &tokenEnvVar,
		}

		return newCreds, nil
	}

	return c, nil
}

func (c *TrelloConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*TrelloConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIKey, other.APIKey) {
		return false
	}

	if !utils.PtrEqual(c.Token, other.Token) {
		return false
	}

	return true
}

func (c *TrelloConnection) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *TrelloConnection) GetTtl() int {
	return -1
}

func (c *TrelloConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *TrelloConnection) GetEnv() map[string]cty.Value {
	return nil
}
