package connection

const SqlLiteConnectionType = "sqllite"

type SqlLiteConnection struct {
	ConnectionImpl
	Database *string `json:"database,omitempty" cty:"database" hcl:"database,optional"`
}
