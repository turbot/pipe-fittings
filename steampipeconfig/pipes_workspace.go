package steampipeconfig

import "strings"

// IsPipesWorkspaceIdentifier returns whether name is a cloud workspace identifier
// of the form: {identity_handle}/{workspace_handle},
func IsPipesWorkspaceIdentifier(name string) bool {
	return len(strings.Split(name, "/")) == 2
}
