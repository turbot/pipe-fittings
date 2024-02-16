package test_init

import (
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
)

func SetAppSpecificConstants() {
	installDir, err := files.Tildefy("~/.flowpipe")
	if err != nil {
		panic(err)
	}

	app_specific.AppName = "flowpipe"
	// TODO unify version logic with steampipe
	//app_specific.AppVersion
	app_specific.AutoVariablesExtensions = []string{".auto.fpvars"}
	//app_specific.ClientConnectionAppNamePrefix
	//app_specific.ClientSystemConnectionAppNamePrefix
	app_specific.DefaultInstallDir = installDir
	app_specific.DefaultVarsFileName = "flowpipe.fpvars"
	//app_specific.DefaultDatabase
	//app_specific.EnvAppPrefix
	app_specific.EnvInputVarPrefix = "FP_VAR_"
	//app_specific.InstallDir
	app_specific.ConfigExtension = ".fpc"
	app_specific.ModDataExtensions = []string{".fp"}
	app_specific.VariablesExtensions = []string{".fpvars"}
	//app_specific.ServiceConnectionAppNamePrefix
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	app_specific.WorkspaceDataDir = ".flowpipe"
}
