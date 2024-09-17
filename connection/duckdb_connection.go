package connection

import "github.com/turbot/pipe-fittings/modconfig"

type DuckDbConnection struct {
	modconfig.ConnectionImpl
	Database *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
}
