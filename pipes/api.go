package pipes

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipes-sdk-go"
)

func newPipesClient(token string) *pipes.APIClient {
	// Create a default configuration
	configuration := pipes.NewConfiguration()
	configuration.Host = viper.GetString(constants.ArgPipesHost)

	// Add your Turbot Pipes user token as an auth header
	if token != "" {
		configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// Create a client
	return pipes.NewAPIClient(configuration)
}

func getLoginTokenConfirmUIUrl() string {
	url := url.URL{
		Scheme: "https",
		Host:   viper.GetString(constants.ArgPipesHost),
		Path:   "/login/token",
	}
	return url.String()
}
