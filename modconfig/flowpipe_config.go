package modconfig

type FlowpipeConfig struct {
	Credentials map[string]Credential
}

func NewFlowpipeConfig() *FlowpipeConfig {
	fpConfig := FlowpipeConfig{
		Credentials: DefaultCredentials(),
	}

	return &fpConfig
}
