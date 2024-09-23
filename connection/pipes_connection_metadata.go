package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/pipe-fittings/constants"
	"time"
)

type PipesConnectionMetadata struct {
	CloudHost  *string `json:"cloud_host,omitempty" cty:"cloud_host" hcl:"cloud_host,optional"`
	User       *string `json:"user,omitempty" cty:"user" hcl:"user,optional"`
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
	jsonResponse, err := m.callPipesCredApi(ctx)
	if err != nil {
		return nil, err
	}
	return m.handlePipesCredApiResponse(jsonResponse, target)
}

func (m PipesConnectionMetadata) handlePipesCredApiResponse(jsonResponse json.RawMessage, target PipelingConnection) (PipelingConnection, error) {
	// unmarshal
	var apiResponse pipesApiCredResponse
	apiResponse.Config = target
	err := json.Unmarshal(jsonResponse, &apiResponse)
	if err != nil {
		return nil, err
	}

	// always set a ttl, even if pipes does not provide one
	ttl := constants.DefaultConnectionTtl
	// if api response contains a use before, set the ttl from it
	if !apiResponse.UseBefore.IsZero() {
		ttl = int(apiResponse.UseBefore.Sub(time.Now()).Seconds())
	}

	target.SetTtl(ttl)

	return target, nil
}

func (m PipesConnectionMetadata) callPipesCredApi(ctx context.Context) (json.RawMessage, error) {
	return nil, fmt.Errorf("not implemented")
}
