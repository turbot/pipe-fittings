package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type PipelingConnection interface {
	modconfig.HclResource
	modconfig.ResourceWithMetadata

	SetHclResourceImpl(hclResourceImpl modconfig.HclResourceImpl)
	GetConnectionType() string
	SetConnectionType(string)
	GetUnqualifiedName() string

	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (PipelingConnection, error)
	GetTtl() int // in seconds

	Validate() hcl.Diagnostics
	getEnv() map[string]cty.Value

	Equals(PipelingConnection) bool
}
