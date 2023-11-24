package modconfig

import (
	"github.com/turbot/pipe-fittings/options"
)

type FlowpipeConfig struct {
	Credentials    map[string]Credential
	GeneralOptions *options.General
}

func NewFlowpipeConfig() *FlowpipeConfig {
	return &FlowpipeConfig{
		Credentials:    make(map[string]Credential),
		GeneralOptions: &options.General{},
	}
}
