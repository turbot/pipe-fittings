package modinstaller

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/funcs"
	"github.com/turbot/pipe-fittings/v2/parse"
	"github.com/turbot/pipe-fittings/v2/versionmap"
	"github.com/zclconf/go-cty/cty"
)

const modInstallLockFileName = "mod_install.lock"

// how often a command that has to wait tries the lock again
const modLockPollInterval = 100 * time.Millisecond

// where the message about waiting goes
var modLockWaitOutput io.Writer = os.Stderr

// modInstallLockPath is the lock file of a mod location. It is a file of its own in the app's data folder:
// mod.pp and .mod.cache.json are replaced on every write, and a lock on a replaced file protects nothing.
func modInstallLockPath(workspacePath string) string {
	return filepath.Join(workspacePath, app_specific.WorkspaceDataDir, modInstallLockFileName)
}

// lockModLocation takes an exclusive lock on the mod location and returns the function that releases it.
//
// A command that finds the location locked says so once, on stderr, and waits for the other to finish. If ctx is
// cancelled while it waits, it returns the context's error and has changed nothing.
func lockModLocation(ctx context.Context, workspacePath string) (unlock func(), err error) {
	lockPath := modInstallLockPath(workspacePath)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("could not create the folder for the mod install lock: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644) //nolint:gosec // not a secret, and not read
	if err != nil {
		return nil, fmt.Errorf("could not open the mod install lock '%s': %w", lockPath, err)
	}

	told := false
	for {
		locked, lockErr := tryLockFile(f.Fd())
		if lockErr != nil {
			_ = f.Close()
			return nil, fmt.Errorf("could not lock '%s': %w", lockPath, lockErr)
		}
		if locked {
			// closing the file releases the lock
			return func() { _ = f.Close() }, nil
		}
		if !told {
			told = true
			_, _ = fmt.Fprintf(modLockWaitOutput, "Waiting for another %s mod command on '%s' to finish...\n", app_specific.AppName, workspacePath)
		}
		select {
		case <-ctx.Done():
			_ = f.Close()
			return nil, ctx.Err()
		case <-time.After(modLockPollInterval):
		}
	}
}

// lockAndRefresh locks the mod location, then reads again everything the installer worked out from disk before it
// had the lock. A command that waited for another one to finish would otherwise go on with a mod file and a cache
// as they were before that command wrote them, and write over its changes.
//
// A dry run writes nothing, so it takes no lock and keeps what it has.
func (i *ModInstaller) lockAndRefresh(ctx context.Context) (unlock func(), err error) {
	if i.dryRun {
		return func() {}, nil
	}
	unlock, err = lockModLocation(ctx, i.workspacePath)
	if err != nil {
		return nil, err
	}
	if err := i.refreshFromDisk(); err != nil {
		unlock()
		return nil, err
	}
	return unlock, nil
}

// refreshFromDisk re-reads the require block of the mod file and the lock file, and works out the target mods
// again from the command's args. It is called with the mod location locked.
func (i *ModInstaller) refreshFromDisk() error {
	// the require block, and where the mod block ends (a new require block goes in there)
	if modFilePath := i.workspaceMod.GetFilePath(); modFilePath != "" {
		if _, err := os.Stat(modFilePath); err == nil {
			evalCtx := &hcl.EvalContext{
				Functions: funcs.ContextFunctions(i.workspacePath),
				Variables: make(map[string]cty.Value),
			}
			current, res := parse.ParseModDefinition(modFilePath, evalCtx)
			if res.Diags.HasErrors() {
				return error_helpers.HclDiagsToError("Failed to load mod", res.Diags)
			}
			i.workspaceMod.Require = current.Require
			i.workspaceMod.DeclRange = current.DeclRange
			i.oldRequire = nil
			if require := i.workspaceMod.Require; require != nil {
				i.oldRequire = require.Clone()
			}
		}
	}

	workspaceLock, err := versionmap.LoadWorkspaceLock(i.workspacePath)
	if err != nil {
		return err
	}
	i.installData = NewInstallData(workspaceLock, i.workspaceMod)

	targetMods, err := i.GetRequiredModVersionsFromArgs(i.modArgs)
	if err != nil {
		return err
	}
	i.targetMods = targetMods
	return nil
}
