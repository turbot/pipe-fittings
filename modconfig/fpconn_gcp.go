package modconfig

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/oauth2/google"
)

type GcpConnection struct {
	ConnectionImpl

	Credentials *string `json:"credentials,omitempty" cty:"credentials" hcl:"credentials,optional"`
	Ttl         *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`

	AccessToken *string `json:"access_token,omitempty" cty:"access_token" hcl:"access_token,optional"`
}

func (c *GcpConnection) GetConnectionType() string {
	return "gcp"
}

func (c *GcpConnection) Resolve(ctx context.Context) (PipelingConnection, error) {

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
	return hcl.Diagnostics{}
}

func (c *GcpConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GcpConnection) GetTtl() int {
	if c.Ttl == nil {
		return 5 * 60 // in seconds
	}
	return *c.Ttl
}

func (c *GcpConnection) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}
