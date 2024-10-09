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

func ConnectionTypeMeetsRequiredType(requiredType, actualType string) bool {
	// handle type connection and connection.<subtype>
	requiredTypeParts := strings.Split(requiredType, ".")
	typeParts := strings.Split(actualType, ".")

	if len(requiredTypeParts) == 1 && requiredTypeParts[0] != typeParts[0] {
		return false
	} else if len(requiredTypeParts) == 2 && requiredType != actualType {
		return false
	}
	return true
}
