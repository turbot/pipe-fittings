package connection

type DuckDbConnection struct {
	ConnectionImpl
	Database *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
}
