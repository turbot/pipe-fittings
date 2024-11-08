package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"strings"
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

type ConnectionStringOpt func(ConnectionStringProvider)

type TimeRangeProvider interface {
	SetTimeRange(utils.TimeRange)
}

// ConnectionStringProvider is implemented by all connections which can provide a connection string
type ConnectionStringProvider interface {
	GetConnectionString(...ConnectionStringOpt) (string, error)
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
