package connection

const SteampipePgConnectionType = "steampipe_pg"

type SteampipePgConnection struct {
	ConnectionImpl
	UserName   *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Host       *string `json:"host,omitempty" cty:"host" hcl:"host,optional"`
	Port       *int    `json:"port,omitempty" cty:"port" hcl:"port,optional"`
	Password   *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
	SearchPath *string `json:"search_path,omitempty" cty:"search_path" hcl:"search_path,optional"`
}
