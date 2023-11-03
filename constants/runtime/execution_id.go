package runtime

import (
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
	"time"

	"github.com/turbot/go-kit/helpers"
)

var (
	ExecutionID = helpers.GetMD5Hash(fmt.Sprintf("%d", time.Now().Nanosecond()))[:4]
)

var (
	// App name used by connections which issue user-initiated queries
	ClientConnectionAppName = fmt.Sprintf("%s_%s", app_specific.ClientConnectionAppNamePrefix, ExecutionID)

	// App name used for queries which support user-initiated queries (load schema, load connection state etc.)
	ClientSystemConnectionAppName = fmt.Sprintf("%s_%s", app_specific.ClientSystemConnectionAppNamePrefix, ExecutionID)

	// App name used for service related queries (plugin manager, refresh connection)
	ServiceConnectionAppName = fmt.Sprintf("%s_%s", app_specific.ServiceConnectionAppNamePrefix, ExecutionID)
)
