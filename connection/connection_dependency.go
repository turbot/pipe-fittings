package connection

import "strings"

// ConnectionDependency is a ConnectionStringProvider implementation that contains a dependency to another connection
type ConnectionDependency struct {
	ConnectionName string
}

func (c ConnectionDependency) GetConnectionString(opt ...ConnectionStringOpt) (string, error) {
	return c.ConnectionName, nil
}

func NewConnectionDependency(depPath string) ConnectionStringProvider {
	return &ConnectionDependency{
		ConnectionName: strings.TrimPrefix(depPath, "connection."),
	}
}
