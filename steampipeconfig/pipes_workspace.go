package steampipeconfig

import (
	"log/slog"
	"strings"

	"github.com/turbot/pipe-fittings/connection"
)

// IsPipesWorkspaceConnectionString returns whether name is a cloud workspace identifier
// of the form: {identity_handle}/{workspace_handle},
func IsPipesWorkspaceConnectionString(csp connection.ConnectionStringProvider) bool {
	// if the connection string is dynamic, assume it is a NOT workspace connection
	if _, dynamic := csp.(connection.DynamicConnectionStringProvider); dynamic {
		return false
	}

	connectionString, err := csp.GetConnectionString()
	if err != nil {
		// unexpected - we do not expect errors from non dynamic connection strings
		slog.Warn("unexpected error getting connection string from non-dynamic provider", "type", csp, "error", err)
		return false
	}

	return len(strings.Split(connectionString, "/")) == 2
}

// IsPipesWorkspaceIdentifier returns whether name is a cloud workspace identifier
// of the form: {identity_handle}/{workspace_handle},
func IsPipesWorkspaceIdentifier(name string) bool {
	return len(strings.Split(name, "/")) == 2
}
