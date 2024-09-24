package connection

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/oauth2/google"
)

const (
	GcpConnectionType = "gcp"
	defaultGcpTtl     = 5 * 60
)

type GcpConnection struct {
	ConnectionImpl

	Credentials *string `json:"credentials,omitempty" cty:"credentials" hcl:"credentials,optional"`
	Ttl         *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func NewGcpConnection(shortName string, declRange hcl.Range) PipelingConnection {
	res := &GcpConnection{
		ConnectionImpl: NewConnectionImpl(GcpConnectionType, shortName, declRange),
	}
	// set the default ttl
	res.SetTtl(defaultGcpTtl)
	return res
}
func (c *GcpConnection) GetConnectionType() string {
	return GcpConnectionType
}

func (c *GcpConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

	// First check if the credential file is supplied
	var credentialFile string
	if c.Credentials != nil && *c.Credentials != "" {
		credentialFile = *c.Credentials
	} else {
		credentialFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credentialFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, perr.InternalWithMessage("failed to get user home directory " + err.Error())
			}

			// If not, check if the default credential file exists
			credentialFile = filepath.Join(homeDir, ".config/gcloud/application_default_credentials.json")
		}
	}

	if credentialFile == "" {
		return c, nil
	}

	// Try to resolve this credential file
	creds, err := os.ReadFile(credentialFile)
	if err != nil {
		return nil, perr.InternalWithMessage("failed to read credential file " + err.Error())
	}

	var credData map[string]interface{}
	if err := json.Unmarshal(creds, &credData); err != nil {
		return nil, perr.InternalWithMessage("failed to parse credential file " + err.Error())
	}

	// Service Account / Authorized User flow
	if credData["type"] == "service_account" || credData["type"] == "authorized_user" {
		// Get a token source using the service account key file

		credentialParam := google.CredentialsParams{
			Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
		}

		credentials, err := google.CredentialsFromJSONWithParams(context.TODO(), creds, credentialParam)
		if err != nil {
			return nil, perr.InternalWithMessage("failed to get credentials from JSON " + err.Error())
		}

		tokenSource := credentials.TokenSource

		// Get the token
		token, err := tokenSource.Token()
		if err != nil {
			return nil, perr.InternalWithMessage("failed to get token from token source " + err.Error())
		}

		newConnection := &GcpConnection{
			AccessToken: &token.AccessToken,
			Credentials: &credentialFile,
		}
		return newConnection, nil
	}

	// oauth2 flow (untested)
	config, err := google.ConfigFromJSON(creds, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, perr.InternalWithMessage("failed to get config from JSON " + err.Error())
	}

	token, err := config.Exchange(context.Background(), "authorization-code")
	if err != nil {
		return nil, perr.InternalWithMessage("failed to get token from config " + err.Error())
	}

	newConnection := &GcpConnection{
		AccessToken: &token.AccessToken,
		Credentials: &credentialFile,
	}
	return newConnection, nil
}

func (c *GcpConnection) Equals(otherConnection PipelingConnection) bool {
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

	other, ok := otherConnection.(*GcpConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.Credentials, other.Credentials) {
		return false
	}

	if !utils.SafeIntEqual(c.Ttl, other.Ttl) {
		return false
	}

	return true
}

func (c *GcpConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.Credentials != nil || c.Ttl != nil || c.AccessToken != nil) {
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

func (c *GcpConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GcpConnection) GetTtl() int {
	// NOTE: if pipes metadata was set we should return the ttl which it returns,
	// rather than any manually configured ttl
	// however - if the HCL contains both a ttl AND pipe metadata
	// this is a validation error so we don't need to check for that here

	// if a ttl was set in the connection config, return it
	if c.Ttl != nil {
		return *c.Ttl
	}

	// otherwise return the base ttl, which will contain either the default, or the value returned by pipes
	return c.ConnectionImpl.GetTtl()
}

func (c *GcpConnection) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}
