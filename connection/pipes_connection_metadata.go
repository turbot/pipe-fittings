package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/perr"

	"github.com/turbot/pipe-fittings/constants"
)

const (
	prodCoudHost   = "pipes.turbot.com"
	stgCloudHost   = "pipes.turbot-stg.com"
	devCloudHost   = "pipes.turbot-dev.com"
	localCloudHost = "pipes.turbot-local.com"
	mockCloudHost  = "localhost:7104"
)

var allowedCloudHosts = []string{
	prodCoudHost,
	stgCloudHost,
	devCloudHost,
	localCloudHost,
	mockCloudHost,
}

type PipesConnectionMetadata struct {
	CloudHost  *string `json:"cloud_host,omitempty" cty:"cloud_host" hcl:"cloud_host,optional"`
	User       *string `json:"user,omitempty" cty:"user" hcl:"user,optional"`
	Org        *string `json:"org,omitempty" cty:"org" hcl:"org,optional"`
	Workspace  *string `json:"workspace,omitempty" cty:"workspace" hcl:"workspace,optional"`
	Connection *string `json:"connection,omitempty" cty:"connection" hcl:"connection,optional"`
}

type pipesApiCredResponse struct {
	Config    PipelingConnection `json:"config"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	UseBefore time.Time          `json:"use_before"`
}

func (m PipesConnectionMetadata) Resolve(ctx context.Context, target PipelingConnection) (PipelingConnection, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	err := m.callPipesCredApi(target)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (m PipesConnectionMetadata) callPipesCredApi(target PipelingConnection) error {
	// get token from env
	// NOTE: use app specific pipes token env, e.g. FLOWPIPE_PIPES_TOKEN
	token, ok := os.LookupEnv(app_specific.EnvPipesToken)
	if !ok || token == "" {
		return fmt.Errorf("missing environment variable %s", app_specific.EnvPipesToken)
	}

	// API endpoint
	url := m.endpoint()

	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return perr.InternalWithMessage("failed to create request")
	}

	// Set the Authorization header with the Bearer token
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return perr.InternalWithMessage("failed to execute request")
	}
	defer resp.Body.Close()

	// Check if the status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return perr.InternalWithMessage(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	// Parse the JSON response
	return m.handlePipesCredApiResponse(resp.Body, target)
}

func (m PipesConnectionMetadata) handlePipesCredApiResponse(resp io.ReadCloser, target PipelingConnection) error {
	// unmarshal into pipesApiCredResponse, populated with the target connection as Config
	var apiResponse pipesApiCredResponse
	apiResponse.Config = target

	err := json.NewDecoder(resp).Decode(&apiResponse)
	if err != nil {
		return perr.InternalWithMessage("failed to decode response body")
	}

	// always set a ttl, even if pipes does not provide one
	ttl := constants.DefaultConnectionTtl
	// if api response contains a use before, set the ttl from it
	if !apiResponse.UseBefore.IsZero() {
		ttl = int(time.Until(apiResponse.UseBefore).Seconds())
	}

	target.SetTtl(ttl)

	return nil
}

func (m PipesConnectionMetadata) endpoint() string {
	cloudHost := prodCoudHost

	if m.CloudHost != nil {
		cloudHost = *m.CloudHost
	}

	if m.CloudHost != nil && *m.CloudHost == mockCloudHost {
		if m.Org != nil {
			return fmt.Sprintf("http://%s/api/v0/org/%s/workspace/%s/connection/%s/private", cloudHost, *m.Org, *m.Workspace, *m.Connection)
		}
		return fmt.Sprintf("http://%s/api/v0/user/%s/workspace/%s/connection/%s/private", cloudHost, *m.User, *m.Workspace, *m.Connection)
	}

	// org or user?
	if m.Org != nil {
		return fmt.Sprintf("https://%s/api/v0/org/%s/workspace/%s/connection/%s/private", cloudHost, *m.Org, *m.Workspace, *m.Connection)
	}
	return fmt.Sprintf("https://%s/api/v0/user/%s/workspace/%s/connection/%s/private", cloudHost, *m.User, *m.Workspace, *m.Connection)
}

func (m PipesConnectionMetadata) validate() error {
	// connection, workspace and either user or org are required
	if m.Connection == nil {
		return perr.BadRequestWithMessage("connection is required")

	}
	if m.Workspace == nil {
		return perr.BadRequestWithMessage("workspace is required")
	}
	if m.User == nil && m.Org == nil {
		return perr.BadRequestWithMessage("either user or org is required")
	}
	// if org is provided, user is not allowed
	if m.Org != nil && m.User != nil {
		return perr.BadRequestWithMessage("only one of user or org is allowed")
	}

	// cloudhost, if provided, must END in pipes.turbot.com
	if m.CloudHost != nil {
		valid := false
		for _, host := range allowedCloudHosts {
			if strings.HasSuffix(*m.CloudHost, host) {
				valid = true
				break
			}
		}
		if !valid {
			return perr.BadRequestWithMessage("cloud_host must end in one of the allowed hosts")
		}
	}

	return nil
}
