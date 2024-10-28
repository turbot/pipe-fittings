package modinstaller

import (
	"context"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/utils"
)

func UninstallWorkspaceDependencies(ctx context.Context, opts *InstallOpts) (_ *InstallData, err error) {
	utils.LogTime("cmd.UninstallWorkspaceDependencies")
	defer func() {
		utils.LogTime("cmd.UninstallWorkspaceDependencies end")
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// uninstall workspace dependencies

	// set update strategy to minimal
	opts.UpdateStrategy = constants.ModUpdateMinimal
	installer, err := NewModInstaller(opts)
	if err != nil {
		return nil, err
	}

	if err := installer.UninstallWorkspaceDependencies(ctx); err != nil {
		return nil, err
	}

	return installer.installData, nil

}
