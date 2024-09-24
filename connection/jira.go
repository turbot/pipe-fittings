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

const JiraConnectionType = "jira"

type JiraConnection struct {
	ConnectionImpl

	APIToken *string `json:"api_token,omitempty" cty:"api_token" hcl:"api_token,optional"`
	BaseURL  *string `json:"base_url,omitempty" cty:"base_url" hcl:"base_url,optional"`
	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
}

func NewJiraConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &JiraConnection{
		ConnectionImpl: NewConnectionImpl(JiraConnectionType, shortName, declRange),
	}
}
func (c *JiraConnection) GetConnectionType() string {
	return JiraConnectionType
}

func (c *JiraConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

	if c.APIToken == nil && c.BaseURL == nil && c.Username == nil {
		// The order of precedence for the Jira API token environment variable
		// 1. JIRA_API_TOKEN
		// 2. JIRA_TOKEN
		jiraAPITokenEnvVar := os.Getenv("JIRA_TOKEN")
		if os.Getenv("JIRA_API_TOKEN") != "" {
			jiraAPITokenEnvVar = os.Getenv("JIRA_API_TOKEN")
		}

		jiraURLEnvVar := os.Getenv("JIRA_URL")
		jiraUserEnvVar := os.Getenv("JIRA_USER")

		// Don't modify existing connection, resolve to a new one
		newConnection := &JiraConnection{
			ConnectionImpl: c.ConnectionImpl,
			APIToken:       &jiraAPITokenEnvVar,
			BaseURL:        &jiraURLEnvVar,
			Username:       &jiraUserEnvVar,
		}

		return newConnection, nil
	}
	return c, nil
}

func (c *JiraConnection) Equals(otherConnection PipelingConnection) bool {
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

	other, ok := otherConnection.(*JiraConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.APIToken, other.APIToken) {
		return false
	}

	if !utils.PtrEqual(c.BaseURL, other.BaseURL) {
		return false
	}

	if !utils.PtrEqual(c.Username, other.Username) {
		return false
	}

	return true
}

func (c *JiraConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.APIToken != nil || c.BaseURL != nil || c.Username != nil) {
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

func (c *JiraConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *JiraConnection) GetEnv() map[string]cty.Value {
	return nil
}
