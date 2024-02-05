package modconfig

type FlowpipeConfig struct {
	Credentials  map[string]Credential
	Integrations map[string]Integration
}

func NewFlowpipeConfig() *FlowpipeConfig {
	fpConfig := FlowpipeConfig{
		Credentials:  DefaultCredentials(),
		Integrations: make(map[string]Integration),
	}

	return &fpConfig
}
