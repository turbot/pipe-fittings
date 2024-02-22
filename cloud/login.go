package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/utils"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

var UnconfirmedError = "Not confirmed"

// WebLogin POSTs to ${envBaseUrl}/api/latest/login/token to retrieve a login is
// it then opens the login webpage and returns th eid
func WebLogin(ctx context.Context) (string, error) {
	client := newSteampipeCloudClient(viper.GetString(constants.ArgPipesToken))

	tempTokenReq, _, err := client.Auth.LoginTokenCreate(ctx).Execute()
	if err != nil {
		return "", sperr.WrapWithMessage(err, "failed to create login token")
	}
	id := tempTokenReq.Id
	// add in id query string
	browserUrl := fmt.Sprintf("%s?r=%s", getLoginTokenConfirmUIUrl(), id)

	fmt.Println()                                  //nolint:forbidigo // acceptable
	fmt.Printf("Verify login at %s\n", browserUrl) //nolint:forbidigo // acceptable

	if err = utils.OpenBrowser(browserUrl); err != nil {
		slog.Info("failed to open login web page")
	}

	return id, nil
}

// GetLoginToken uses the login id and code and retrieves an authentication token
func GetLoginToken(ctx context.Context, id, code string) (string, error) {
	client := newSteampipeCloudClient("")
	tokenResp, _, err := client.Auth.LoginTokenGet(ctx, id).Code(code).Execute()
	if err != nil {
		var apiErr steampipecloud.GenericOpenAPIError
		if errors.As(err, &apiErr) {
			var body = map[string]any{}
			if err := json.Unmarshal(apiErr.Body(), &body); err == nil {
				return "", sperr.New("%s", body["detail"])
			}
		}
		return "", sperr.Wrap(err)
	}
	if tokenResp.GetToken() == "" && tokenResp.GetState() == "pending" {
		return "", sperr.New("login request has not been confirmed - select 'Verify' and enter the verification code")
	}
	return tokenResp.GetToken(), nil
}

// SaveToken writes the token to  ~/.steampipe/internal/{cloud-host}.tptt
func SaveToken(token string) error {
	tokenPath := tokenFilePath(viper.GetString(constants.ArgPipesHost))
	return sperr.Wrap(os.WriteFile(tokenPath, []byte(token), 0600))
}

func LoadToken() (string, error) {
	tokenPath := tokenFilePath(viper.GetString(constants.ArgPipesHost))
	if !filehelpers.FileExists(tokenPath) {
		return "", nil
	}
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", sperr.WrapWithMessage(err, "failed to load token file '%s'", tokenPath)
	}
	return string(tokenBytes), nil
}

func GetUserName(ctx context.Context, token string) (string, error) {
	client := newSteampipeCloudClient(token)
	actor, _, err := client.Actors.Get(ctx).Execute()
	if err != nil {
		return "", sperr.Wrap(err)
	}
	return getActorName(actor), nil
}

func getActorName(actor steampipecloud.User) string {
	if name, ok := actor.GetDisplayNameOk(); ok {
		return *name
	}
	return actor.Handle
}

func tokenFilePath(pipesHost string) string {
	tokenPath := path.Join(filepaths.EnsurePipesInternalDir(), fmt.Sprintf("%s%s", pipesHost, constants.TokenExtension))
	return tokenPath
}
