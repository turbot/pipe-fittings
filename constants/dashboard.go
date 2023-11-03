package constants

import (
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
)

// DashboardListenAddresses is an arrays is listen addresses which Steampipe accepts
var DashboardListenAddresses = []string{"localhost", "127.0.0.1"}

const (
	DashboardServerDefaultPort    = 9194
	DashboardAssetsImageRefFormat = "us-docker.pkg.dev/steampipe/steampipe/assets:%s"
)

func DashboardAssetsImageRef() string {
	return fmt.Sprintf(DashboardAssetsImageRefFormat, app_specific.AppVersion.String())
}
