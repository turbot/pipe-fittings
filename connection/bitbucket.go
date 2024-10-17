package connection

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const BitbucketConnectionType = "bitbucket"

type BitbucketConnection struct {
	ConnectionImpl

	BaseURL  *string `json:"base_url,omitempty" cty:"base_url" hcl:"base_url,optional"`
	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func NewBitbucketConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &BitbucketConnection{
		ConnectionImpl: NewConnectionImpl(BitbucketConnectionType, shortName, declRange),
	}
}
func (c *BitbucketConnection) GetConnectionType() string {
	return BitbucketConnectionType
}

func (c *BitbucketConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &BitbucketConnection{ConnectionImpl: c.ConnectionImpl})
	}

	if c.Password == nil && c.BaseURL == nil && c.Username == nil {
		bitbucketURLEnvVar := os.Getenv("BITBUCKET_API_BASE_URL")
		bitbucketUsernameEnvVar := os.Getenv("BITBUCKET_USERNAME")
		bitbucketPasswordEnvVar := os.Getenv("BITBUCKET_PASSWORD")

		// Don't modify existing connection, resolve to a new one
		newConnection := &BitbucketConnection{
			ConnectionImpl: c.ConnectionImpl,
			Password:       &bitbucketPasswordEnvVar,
			BaseURL:        &bitbucketURLEnvVar,
			Username:       &bitbucketUsernameEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *BitbucketConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*BitbucketConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.BaseURL, other.BaseURL) {
		return false
	}

	if !utils.PtrEqual(c.Username, other.Username) {
		return false
	}

	if !utils.PtrEqual(c.Password, other.Password) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *BitbucketConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.BaseURL != nil || c.Username != nil || c.Password != nil) {
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

func (c *BitbucketConnection) CtyValue() (cty.Value, error) {

	return ctyValueForConnection(c)

}

func (c *BitbucketConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.BaseURL != nil {
		env["BITBUCKET_API_BASE_URL"] = cty.StringVal(*c.BaseURL)
	}
	if c.Username != nil {
		env["BITBUCKET_USERNAME"] = cty.StringVal(*c.Username)
	}
	if c.Password != nil {
		env["BITBUCKET_PASSWORD"] = cty.StringVal(*c.Password)
	}
	return env
}
