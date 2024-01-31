package modconfig

type FlowpipeConfig struct {
	Credentials map[string]Credential
	// CredentialImports map[string]CredentialImport
}

func NewFlowpipeConfig() *FlowpipeConfig {
	fpConfig := FlowpipeConfig{
		Credentials: DefaultCredentials(),
	}

	return &fpConfig
}
