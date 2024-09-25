package connection

import (
	"context"
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
