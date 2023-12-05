package modconfig

import (
	"github.com/turbot/pipe-fittings/options"
)

type FlowpipeConfig struct {
	Credentials    map[string]Credential
	GeneralOptions *options.General
}

func NewFlowpipeConfig() *FlowpipeConfig {
	fpConfig := FlowpipeConfig{
		Credentials:    DefaultCredentials(),
		GeneralOptions: &options.General{},
	}

	return &fpConfig
}
