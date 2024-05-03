package filepaths

import (
	"fmt"
	"path"
	"strings"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants/runtime"
)

// mod related constants
const (
	WorkspaceModDir             = "mods"
	WorkspaceModShadowDirPrefix = ".mods."
	WorkspaceConfigFileName     = "workspace.spc"
	WorkspaceLockFileName       = ".mod.cache.json"
)

func WorkspaceModPath(workspacePath string) string {
	return path.Join(workspacePath, app_specific.WorkspaceDataDir, WorkspaceModDir)
}

func WorkspaceModShadowPath(workspacePath string) string {
	return path.Join(workspacePath, app_specific.WorkspaceDataDir, fmt.Sprintf("%s%s", WorkspaceModShadowDirPrefix, runtime.ExecutionID))
}

func IsModInstallShadowPath(dirName string) bool {
	return strings.HasPrefix(dirName, WorkspaceModShadowDirPrefix)
}

func WorkspaceLockPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceLockFileName)
}

func DefaultVarsFilePath(workspacePath string) string {
	return path.Join(workspacePath, app_specific.DefaultVarsFileName)
}
func LegacyDefaultVarsFilePath(workspacePath string) string {
	if app_specific.LegacyDefaultVarsFileName == "" {
		return ""
	}
	return path.Join(workspacePath, app_specific.LegacyDefaultVarsFileName)
}
