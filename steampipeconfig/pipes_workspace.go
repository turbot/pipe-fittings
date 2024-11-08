package steampipeconfig

import (
	"github.com/turbot/pipe-fittings/connection"
	"strings"
)

// IsPipesWorkspaceIdentifier returns whether name is a cloud workspace identifier
// of the form: {identity_handle}/{workspace_handle},
func IsPipesWorkspaceConnectionString(csp connection.ConnectionStringProvider) bool {
	if cs, ok := csp.(connection.ConnectionString); ok {
		return len(strings.Split(cs.ConnectionString, "/")) == 2
	}
	return false
}

// IsPipesWorkspaceIdentifier returns whether name is a cloud workspace identifier
// of the form: {identity_handle}/{workspace_handle},
func IsPipesWorkspaceIdentifier(name string) bool {
	return len(strings.Split(name, "/")) == 2
}
