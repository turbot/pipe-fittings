package cmdconfig

type EnvVarType int

const (
	EnvVarTypeString EnvVarType = iota
	EnvVarTypeInt
	EnvVarTypeBool
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=EnvVarType
