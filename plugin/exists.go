package plugin

import (
	"context"

	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/pipe-fittings/v2/versionfile"
)

// Exists looks up the version file and reports whether a plugin is already installed
func Exists(ctx context.Context, plugin string) (bool, error) {
	versionData, err := versionfile.LoadPluginVersionFile(ctx)
	if err != nil {
		return false, err
	}

	imageRef := ociinstaller.NewImageRef(plugin)

	// lookup in the version data
	_, found := versionData.Plugins[imageRef.DisplayImageRef()]
	return found, nil
}
