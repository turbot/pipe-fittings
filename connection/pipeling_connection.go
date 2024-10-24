package connection

import (
	"context"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type PipelingConnection interface {
	GetConnectionType() string
	GetShortName() string
	Name() string

	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (PipelingConnection, error)
	GetTtl() int // in seconds

	Validate() hcl.Diagnostics
	GetEnv() map[string]cty.Value

	Equals(PipelingConnection) bool
	GetConnectionImpl() *ConnectionImpl

	SetTtl(int)
}

// ConnectionStringProvider is implemented by all connections which can provide a connection string
type ConnectionStringProvider interface {
	GetConnectionString() string
}

// SearchPathProvider is implemented by all connections which can provide a connection string
type SearchPathProvider interface {
	ConnectionStringProvider
	GetSearchPath() []string
	GetSearchPathPrefix() []string
}

func ConnectionTypeMeetsRequiredType(requiredType, actualResourceType, actualType string) bool {
	// handle type connection and connection.<subtype>
	requiredTypeParts := strings.Split(requiredType, ".")

	if len(requiredTypeParts) == 1 && requiredTypeParts[0] != actualResourceType {
		return false
	} else if len(requiredTypeParts) == 2 && (requiredTypeParts[0] != actualResourceType || requiredTypeParts[1] != actualType) {
		return false
	}
	return true
}
