package connection

import "github.com/turbot/pipe-fittings/modconfig"

type SqlLiteConnection struct {
	modconfig.ConnectionImpl
	Database *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
}
