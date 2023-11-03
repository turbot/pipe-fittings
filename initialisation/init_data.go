package initialisation

import (
	"context"
	"github.com/turbot/pipe-fittings/app_specific"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_client"
	"github.com/turbot/pipe-fittings/db_common"

	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/export"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/modinstaller"
	"github.com/turbot/pipe-fittings/statushooks"
	"github.com/turbot/pipe-fittings/workspace"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe-plugin-sdk/v5/telemetry"
)

type InitData struct {
	Workspace *workspace.Workspace
	Client    *db_client.DbClient
	Result    *db_common.InitResult

	ShutdownTelemetry func()
	ExportManager     *export.Manager
}

func NewErrorInitData(err error) *InitData {
	return &InitData{
		Result: &db_common.InitResult{
			ErrorAndWarnings: *error_helpers.NewErrorsAndWarning(err),
		},
	}
}

func NewInitData() *InitData {
	i := &InitData{
		Result:        &db_common.InitResult{},
		ExportManager: export.NewManager(),
	}

	return i
}

func (i *InitData) RegisterExporters(exporters ...export.Exporter) *InitData {
	for _, e := range exporters {
		i.ExportManager.Register(e)
	}

	return i
}

func (i *InitData) Init(ctx context.Context, opts ...db_client.ClientOption) {
	defer func() {
		if r := recover(); r != nil {
			i.Result.Error = helpers.ToError(r)
		}
		// if there is no error, return context cancellation error (if any)
		if i.Result.Error == nil {
			i.Result.Error = ctx.Err()
		}
	}()

	log.Printf("[INFO] Initializing...")

	// code after this depends of i.Workspace being defined. make sure that it is
	if i.Workspace == nil {
		i.Result.Error = sperr.WrapWithRootMessage(error_helpers.InvalidStateError, "InitData.Init called before setting up Workspace")
		return
	}

	statushooks.SetStatus(ctx, "Initializing")

	// initialise telemetry
	shutdownTelemetry, err := telemetry.Init(app_specific.AppName)
	if err != nil {
		i.Result.AddWarnings(err.Error())
	} else {
		i.ShutdownTelemetry = shutdownTelemetry
	}

	// install mod dependencies if needed
	if viper.GetBool(constants.ArgModInstall) {
		statushooks.SetStatus(ctx, "Installing workspace dependencies")
		log.Printf("[INFO] Installing workspace dependencies")

		opts := modinstaller.NewInstallOpts(i.Workspace.Mod)
		// use force install so that errors are ignored during installation
		// (we are validating prereqs later)
		opts.Force = true
		_, err := modinstaller.InstallWorkspaceDependencies(ctx, opts)
		if err != nil {
			i.Result.Error = err
			return
		}
	}

	// retrieve cloud metadata
	cloudMetadata, err := getCloudMetadata(ctx)
	if err != nil {
		i.Result.Error = err
		return
	}

	// set cloud metadata (may be nil)
	i.Workspace.CloudMetadata = cloudMetadata

	// TODO KAI RETHINK THIS <MOD VALIDATION>
	// we are only validating the CLI version now - ok to do remotely?
	// no need to validate local steampipe and plugin versions for when connecting to remote steampipe database
	// ArgConnectionString is empty when connecting to local database
	//if connectionString := viper.GetString(constants.ArgConnectionString); connectionString == "" {
	// validate steampipe version
	validationWarnings := validateModRequirementsRecursively(i.Workspace.Mod)
	i.Result.AddWarnings(validationWarnings...)
	//}

	statushooks.SetStatus(ctx, "Connecting to steampipe database")
	log.Printf("[INFO] Connecting to steampipe database")
	connectionString := viper.GetString(constants.ArgConnectionString)
	if connectionString == "" {
		i.Result.Error = sperr.New("connection string is not set")
		return
	}

	client, err := db_client.NewDbClient(ctx, connectionString, opts...)
	if err != nil {
		i.Result.Error = err
		return
	}

	// TODO KAI STEAMPIPE ONLY <CACHE>
	log.Printf("[INFO] ValidateClientCacheSettings")
	//if errorsAndWarnings := db_common.ValidateClientCacheSettings(client); errorsAndWarnings != nil {
	//	if errorsAndWarnings.GetError() != nil {
	//		i.Result.Error = errorsAndWarnings.GetError()
	//	}
	//	i.Result.AddWarnings(errorsAndWarnings.Warnings...)
	//}

	i.Client = client
}

func validateModRequirementsRecursively(mod *modconfig.Mod) []string {
	var validationErrors []string

	// validate this mod
	for _, err := range mod.ValidateRequirements() {
		validationErrors = append(validationErrors, err.Error())
	}

	// validate dependent mods
	for childDependencyName, childMod := range mod.ResourceMaps.Mods {
		// TODO : The 'mod.DependencyName == childMod.DependencyName' check has to be done because
		// of a bug in the resource loading code which also puts the mod itself into the resource map
		// [https://github.com/turbot/steampipe/issues/3341]
		if childDependencyName == "local" || mod.DependencyName == childMod.DependencyName {
			// this is a reference to self - skip (otherwise we will end up with a recursion loop)
			continue
		}
		childValidationErrors := validateModRequirementsRecursively(childMod)
		validationErrors = append(validationErrors, childValidationErrors...)
	}

	return validationErrors
}

func (i *InitData) Cleanup(ctx context.Context) {
	if i.Client != nil {
		i.Client.Close(ctx)
	}
	if i.ShutdownTelemetry != nil {
		i.ShutdownTelemetry()
	}
	if i.Workspace != nil {
		i.Workspace.Close()
	}
}
