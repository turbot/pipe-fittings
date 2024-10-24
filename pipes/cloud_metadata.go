package pipes

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/sperr"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/pipes-sdk-go"
)

func GetPipesMetadata(ctx context.Context, workspaceDatabaseString, token string) (*steampipeconfig.PipesMetadata, error) {
	client := newPipesClient(token)

	parts := strings.Split(workspaceDatabaseString, "/")
	if len(parts) != 2 {
		return nil, sperr.New("invalid argument '%s' - must be in format <identity>/<workspace>", workspaceDatabaseString)
	}
	identityHandle := parts[0]
	workspaceHandle := parts[1]

	// get the identity
	identity, _, err := client.Identities.Get(ctx, identityHandle).Execute()
	if err != nil {
		return nil, sperr.New("Invalid 'workspace-database' argument '%s'.\nPlease check the identity and workspace names and try again.", workspaceDatabaseString)
	}

	// get the workspace
	var cloudWorkspace pipes.Workspace
	if identity.Type == "user" {
		cloudWorkspace, _, err = client.UserWorkspaces.Get(ctx, identityHandle, workspaceHandle).Execute()
	} else {
		cloudWorkspace, _, err = client.OrgWorkspaces.Get(ctx, identityHandle, workspaceHandle).Execute()
	}

	if error_helpers.IsInvalidWorkspaceDatabaseArg(err) {
		return nil, sperr.New("Invalid 'workspace-database' argument '%s'.\nPlease check the workspace name and try again.", workspaceDatabaseString)
	} else if error_helpers.IsInvalidCloudToken(err) {
		return nil, error_helpers.InvalidCloudTokenError()
	}

	workspaceHost := cloudWorkspace.GetHost()
	databaseName := cloudWorkspace.GetDatabaseName()

	actor, _, err := client.Actors.Get(ctx).Execute()
	if err != nil {
		return nil, error_helpers.InvalidCloudTokenError()
	}

	password, _, err := client.Users.GetDBPassword(ctx, actor.GetHandle()).Execute()
	if err != nil {
		return nil, sperr.Wrap(err)
	}

	connectionString := fmt.Sprintf("postgresql://%s:%s@%s-%s.%s:9193/%s", actor.Handle, password.Password, identityHandle, workspaceHandle, workspaceHost, databaseName)

	pipesMetadata := &steampipeconfig.PipesMetadata{
		Actor: &steampipeconfig.ActorMetadata{
			Id:     actor.Id,
			Handle: actor.Handle,
		},
		Identity: &steampipeconfig.IdentityMetadata{
			Id:     cloudWorkspace.IdentityId,
			Type:   identity.Type,
			Handle: identityHandle,
		},
		Workspace: &steampipeconfig.WorkspaceMetadata{
			Id:     cloudWorkspace.Id,
			Handle: cloudWorkspace.Handle,
		},

		ConnectionString: connectionString,
	}

	return pipesMetadata, nil
}
