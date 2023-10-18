package dashboardexecute

import (
	"github.com/turbot/pipe-fittings/dashboardtypes"
)

type RuntimeDependencyPublishOption = func(target *RuntimeDependencyPublishTarget)

func WithTransform(transform func(*dashboardtypes.ResolvedRuntimeDependencyValue) *dashboardtypes.ResolvedRuntimeDependencyValue) RuntimeDependencyPublishOption {
	return func(c *RuntimeDependencyPublishTarget) {
		c.transform = transform
	}
}
