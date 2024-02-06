package flowpipeconfig

import (
	"log/slog"

	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/modconfig"
)

type FlowpipeConfig struct {
	Credentials  map[string]credential.Credential
	Integrations map[string]modconfig.Integration
}

func NewFlowpipeConfig() *FlowpipeConfig {
	defaultCreds, err := credential.DefaultCredentials()
	if err != nil {
		slog.Error("Unable to create default credentials", "error", err)
		return nil
	}

	fpConfig := FlowpipeConfig{
		Credentials:  defaultCreds,
		Integrations: make(map[string]modconfig.Integration),
	}

	return &fpConfig
}
