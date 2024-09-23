package connection

type PipesConnectionMetadata struct {
	CloudHost  *string `json:"cloud_host,omitempty" cty:"cloud_host" hcl:"cloud_host,optional"`
	Workspace  *string `json:"workspace,omitempty" cty:"workspace" hcl:"workspace,optional"`
	Connection *string `json:"connection,omitempty" cty:"connection" hcl:"connection,optional"`
}
