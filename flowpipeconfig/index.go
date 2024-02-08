package flowpipeconfig

import (
	"log/slog"

	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/modconfig"
)

type FlowpipeConfig struct {
	Credentials  map[string]credential.Credential
	Integrations map[string]modconfig.Integration
	Notifiers    map[string]modconfig.Notifier
}

func NewFlowpipeConfig() *FlowpipeConfig {
	defaultCreds, err := credential.DefaultCredentials()
	if err != nil {
		slog.Error("Unable to create default credentials", "error", err)
		return nil
	}

	defaultIntegrations, err := modconfig.DefaultIntegrations()
	if err != nil {
		slog.Error("Unable to create default integrations", "error", err)
		return nil
	}

	defaultNotifiers, err := modconfig.DefaultNotifiers(defaultIntegrations["webform.default"])
	if err != nil {
		slog.Error("Unable to create default notifiers", "error", err)
		return nil
	}

	fpConfig := FlowpipeConfig{
		Credentials:  defaultCreds,
		Integrations: defaultIntegrations,
		Notifiers:    defaultNotifiers,
	}

	return &fpConfig
}
