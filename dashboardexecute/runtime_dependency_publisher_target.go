package dashboardexecute

import (
	"github.com/turbot/pipe-fittings/dashboardtypes"
)

type RuntimeDependencyPublishTarget struct {
	transform func(*dashboardtypes.ResolvedRuntimeDependencyValue) *dashboardtypes.ResolvedRuntimeDependencyValue
	channel   chan *dashboardtypes.ResolvedRuntimeDependencyValue
}
