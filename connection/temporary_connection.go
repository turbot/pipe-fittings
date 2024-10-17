package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// TemporaryConnection is a connection that is not yet resolved
// is is used when a Flowpipe mod database field references a connection.
// Connections are not resolved until runtime, so the database field of the mod is populated with the connection name
// (TemporaryConnection.GetConnectionString returns the connection name)
type TemporaryConnection struct {
	ConnectionImpl
}

func (t *TemporaryConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(t)
}

func (t *TemporaryConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	return t, nil
}

func (t *TemporaryConnection) Validate() hcl.Diagnostics {
	return nil
}

func (t *TemporaryConnection) GetEnv() map[string]cty.Value {
	return map[string]cty.Value{}
}

func (t *TemporaryConnection) Equals(connection PipelingConnection) bool {
	return false
}

// GetConnectionString returns the connection name (including "connection." prefix)
// this will be resolved to a connection at run time
func (t *TemporaryConnection) GetConnectionString() string {
	name := "connection." + t.FullName
	return name
}
