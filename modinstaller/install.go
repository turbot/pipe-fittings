package modinstaller

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
)

func InstallWorkspaceDependencies[T modconfig.ResourceMapsI](ctx context.Context, opts *InstallOpts[T]) (_ *InstallData, err error) {
	utils.LogTime("cmd.InstallWorkspaceDependencies")
	defer func() {
		utils.LogTime("cmd.InstallWorkspaceDependencies end")
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// install workspace dependencies
	installer, err := NewModInstaller(opts)
	if err != nil {
		return nil, err
	}

	if err := installer.InstallWorkspaceDependencies(ctx); err != nil {
		return nil, err
	}

	return installer.installData, nil
}
