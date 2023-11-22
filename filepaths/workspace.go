package filepaths

import (
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
	"path"
	"path/filepath"
	"strings"

	"github.com/turbot/pipe-fittings/constants/runtime"
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

func ModFilePath(modFolder string) string {
	modFilePath := filepath.Join(modFolder, app_specific.ModFileName)
	return modFilePath
}
