package ociinstaller

import (
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
)

func MediaTypeConfig() string {
	return fmt.Sprintf("application/vnd.turbot.%s.config.v1+json", app_specific.AppName)
}

// deprecate this....
func MediaTypePluginConfig() string {
	return fmt.Sprintf("application/vnd.turbot.%s.plugin.config.v1+json", app_specific.AppName)
}

func MediaTypePluginLicenseLayer() string {
	return fmt.Sprintf("application/vnd.turbot.%s.plugin.license.layer.v1+text", app_specific.AppName)
}

func MediaTypePluginDocsLayer() string {
	return fmt.Sprintf("application/vnd.turbot.%s.plugin.docs.layer.v1+tar", app_specific.AppName)
}

func MediaTypePluginSpcLayer() string {
	return fmt.Sprintf("application/vnd.turbot.%s.plugin.spc.layer.v1+tar", app_specific.AppName)
}
