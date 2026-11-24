package modinstaller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/parse"
)

// syncBuffer is a bytes.Buffer that the installer and the test may use from different goroutines
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// setupModApp sets what an app sets at start-up, as far as a mod install reads it
func setupModApp(t *testing.T) {
	t.Helper()
	prevExt, prevDataDir, prevName := app_specific.ModDataExtensions, app_specific.WorkspaceDataDir, app_specific.AppName
	app_specific.ModDataExtensions = []string{".pp"}
	app_specific.WorkspaceDataDir = ".testapp"
	app_specific.AppName = "testapp"
	prevResources := modconfig.AppSpecificNewModResourcesFunc
	modconfig.AppSpecificNewModResourcesFunc = func(*modconfig.Mod, ...modconfig.ModResources) modconfig.ModResources {
		return modRoutesResources{}
	}
	t.Cleanup(func() {
		app_specific.ModDataExtensions, app_specific.WorkspaceDataDir, app_specific.AppName = prevExt, prevDataDir, prevName
		modconfig.AppSpecificNewModResourcesFunc = prevResources
	})
	viper.Set(constants.ArgPrune, false)
}

// modLocation is a mod location with a mod.pp of its own and local mods to install into it
type modLocation struct {
	path    string
	modFile string
	mods    map[string]string
}

func newModLocation(t *testing.T, localMods ...string) *modLocation {
	t.Helper()
	setupModApp(t)
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	l := &modLocation{path: filepath.Join(base, "w"), mods: map[string]string{}}
	l.modFile = filepath.Join(l.path, "mod.pp")
	writeFile(t, l.modFile, "mod \"local\" {\n  title = \"w\"\n}\n")
	for _, name := range localMods {
		dir := filepath.Join(base, "src", name)
		writeFile(t, filepath.Join(dir, "mod.pp"), "mod \""+name+"\" {}\n")
		l.mods[name] = dir
	}
	return l
}

// installer makes an installer the way an app's mod command does: parse the mod file, then hand the mod over
func (l *modLocation) installer(t *testing.T, command string, modArgs ...string) *ModInstaller {
	t.Helper()
	workspaceMod, err := parse.LoadModfile(l.path)
	if err != nil {
		t.Fatal(err)
	}
	i, err := NewModInstaller(&InstallOpts{WorkspaceMod: workspaceMod, Command: command, ModArgs: modArgs, UpdateStrategy: constants.ModUpdateMinimal})
	if err != nil {
		t.Fatal(err)
	}
	return i
}

// names are what a local mod is called in the require block and the cache: its path
func (l *modLocation) names(mods ...string) []string {
	var res []string
	for _, m := range mods {
		res = append(res, l.mods[m])
	}
	return res
}

func (l *modLocation) lockFile() string { return filepaths.WorkspaceLockPath(l.path) }

// requireBlocks counts the require blocks of mod.pp
func (l *modLocation) requireBlocks(t *testing.T) int {
	t.Helper()
	b, err := os.ReadFile(l.modFile)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(b), "require {")
}

// requiredMods are the mods named by the require block of mod.pp, as a fresh parse sees them (the parse fails
// when there is more than one require block)
func (l *modLocation) requiredMods(t *testing.T) []string {
	t.Helper()
	mod, err := parse.LoadModfile(l.path)
	if err != nil {
		t.Fatalf("the mod does not load: %v", err)
	}
	var names []string
	for _, m := range mod.Require.Mods {
		names = append(names, m.Name)
	}
	return names
}

// lockedMods are the mods in .mod.cache.json
func (l *modLocation) lockedMods(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(l.lockFile())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var cache map[string]map[string]json.RawMessage
	if err := json.Unmarshal(b, &cache); err != nil {
		t.Fatalf("the cache file is not valid JSON: %v", err)
	}
	var names []string
	for _, deps := range cache {
		for name := range deps {
			names = append(names, name)
		}
	}
	return names
}

func assertSameSet(t *testing.T, what string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", what, got, want)
	}
	have := map[string]bool{}
	for _, g := range got {
		have[g] = true
	}
	for _, w := range want {
		if !have[w] {
			t.Fatalf("%s = %v, want %v", what, got, want)
		}
	}
}

// runTogether runs the functions at the same time, each starting only once every one of them is ready, and
// returns what each returned
func runTogether(fns ...func() error) []error {
	errs := make([]error, len(fns))
	ready := make(chan struct{})
	var wg sync.WaitGroup
	for n, fn := range fns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			errs[n] = fn()
		}()
	}
	close(ready)
	wg.Wait()
	return errs
}

// two installs of different mods at once, each started from the mod file as it was before either wrote it
func TestConcurrentInstallsKeepBothMods(t *testing.T) {
	l := newModLocation(t, "alpha", "beta")
	ctx := context.Background()
	// both commands have read the mod file and the cache before either writes (the app parses the mod first)
	a := l.installer(t, "install", l.mods["alpha"])
	b := l.installer(t, "install", l.mods["beta"])

	errs := runTogether(
		func() error { return a.InstallWorkspaceDependencies(ctx) },
		func() error { return b.InstallWorkspaceDependencies(ctx) },
	)
	for _, err := range errs {
		if err != nil {
			t.Fatalf("install failed: %v", err)
		}
	}

	if n := l.requireBlocks(t); n != 1 {
		t.Errorf("mod.pp has %d require blocks, want 1", n)
	}
	assertSameSet(t, "required mods", l.requiredMods(t), l.names("alpha", "beta"))
	assertSameSet(t, "mods in the cache", l.lockedMods(t), l.names("alpha", "beta"))
}

// the same through the package-level function an app calls
func TestConcurrentInstallWorkspaceDependenciesKeepsBothMods(t *testing.T) {
	l := newModLocation(t, "alpha", "beta", "gamma")
	ctx := context.Background()
	opts := func(mod string) *InstallOpts {
		workspaceMod, err := parse.LoadModfile(l.path)
		if err != nil {
			t.Fatal(err)
		}
		return &InstallOpts{WorkspaceMod: workspaceMod, Command: "install", ModArgs: []string{l.mods[mod]}, UpdateStrategy: constants.ModUpdateMinimal}
	}
	oa, ob, oc := opts("alpha"), opts("beta"), opts("gamma")
	errs := runTogether(
		func() error { _, err := InstallWorkspaceDependencies(ctx, oa); return err },
		func() error { _, err := InstallWorkspaceDependencies(ctx, ob); return err },
		func() error { _, err := InstallWorkspaceDependencies(ctx, oc); return err },
	)
	for _, err := range errs {
		if err != nil {
			t.Fatalf("install failed: %v", err)
		}
	}
	assertSameSet(t, "required mods", l.requiredMods(t), l.names("alpha", "beta", "gamma"))
	assertSameSet(t, "mods in the cache", l.lockedMods(t), l.names("alpha", "beta", "gamma"))
}

// an install and an uninstall of another mod at once lose neither change
func TestConcurrentInstallAndUninstallLoseNeitherChange(t *testing.T) {
	l := newModLocation(t, "alpha", "beta")
	ctx := context.Background()
	if err := l.installer(t, "install", l.mods["alpha"]).InstallWorkspaceDependencies(ctx); err != nil {
		t.Fatal(err)
	}
	assertSameSet(t, "required mods before", l.requiredMods(t), l.names("alpha"))

	install := l.installer(t, "install", l.mods["beta"])
	uninstall := l.installer(t, "uninstall", l.mods["alpha"])
	errs := runTogether(
		func() error { return install.InstallWorkspaceDependencies(ctx) },
		func() error { return uninstall.UninstallWorkspaceDependencies(ctx) },
	)
	for _, err := range errs {
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}
	}

	if n := l.requireBlocks(t); n != 1 {
		t.Errorf("mod.pp has %d require blocks, want 1", n)
	}
	assertSameSet(t, "required mods", l.requiredMods(t), l.names("beta"))
	assertSameSet(t, "mods in the cache", l.lockedMods(t), l.names("beta"))
}

// one install on its own is as before: the mod file and the cache are written, and no lock waits
func TestSingleInstallWorks(t *testing.T) {
	l := newModLocation(t, "alpha")
	out := &syncBuffer{}
	prev := modLockWaitOutput
	modLockWaitOutput = out
	t.Cleanup(func() { modLockWaitOutput = prev })

	if err := l.installer(t, "install", l.mods["alpha"]).InstallWorkspaceDependencies(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertSameSet(t, "required mods", l.requiredMods(t), l.names("alpha"))
	assertSameSet(t, "mods in the cache", l.lockedMods(t), l.names("alpha"))
	if out.String() != "" {
		t.Errorf("an install with nothing to wait for printed %q", out.String())
	}
}

// holdLock takes the lock of the mod location, as another command would, and gives it back when the test ends
func holdLock(t *testing.T, l *modLocation) (release func()) {
	t.Helper()
	unlock, err := lockModLocation(context.Background(), l.path)
	if err != nil {
		t.Fatal(err)
	}
	var once sync.Once
	release = func() { once.Do(unlock) }
	t.Cleanup(release)
	return release
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// a command that has to wait says so once, on stderr, and goes on when the other one lets go
func TestWaitingCommandSaysSoOnceAndGoesOn(t *testing.T) {
	l := newModLocation(t, "alpha")
	out := &syncBuffer{}
	prev := modLockWaitOutput
	modLockWaitOutput = out
	t.Cleanup(func() { modLockWaitOutput = prev })

	release := holdLock(t, l)
	i := l.installer(t, "install", l.mods["alpha"])
	done := make(chan error, 1)
	go func() { done <- i.InstallWorkspaceDependencies(context.Background()) }()

	waitFor(t, "the waiting message", func() bool { return out.String() != "" })
	// several polls pass, and the message must not repeat
	time.Sleep(3 * modLockPollInterval)
	select {
	case err := <-done:
		t.Fatalf("the install did not wait for the lock (returned %v)", err)
	default:
	}
	release()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("install failed after the wait: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the install did not go on once the lock was free")
	}

	if got := strings.Count(out.String(), "Waiting for another testapp mod command"); got != 1 {
		t.Errorf("the waiting message appeared %d times, want once: %q", got, out.String())
	}
	assertSameSet(t, "required mods", l.requiredMods(t), l.names("alpha"))
}

// Ctrl-C while a command waits stops it and leaves the mod file and the cache as they were
func TestCancelWhileWaitingChangesNothing(t *testing.T) {
	l := newModLocation(t, "alpha", "beta")
	if err := l.installer(t, "install", l.mods["alpha"]).InstallWorkspaceDependencies(context.Background()); err != nil {
		t.Fatal(err)
	}
	modBefore, err := os.ReadFile(l.modFile)
	if err != nil {
		t.Fatal(err)
	}
	cacheBefore, err := os.ReadFile(l.lockFile())
	if err != nil {
		t.Fatal(err)
	}

	out := &syncBuffer{}
	prev := modLockWaitOutput
	modLockWaitOutput = out
	t.Cleanup(func() { modLockWaitOutput = prev })

	holdLock(t, l)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// --force must not turn the cancellation into a success
	i := l.installer(t, "install", l.mods["beta"])
	i.force = true
	done := make(chan error, 1)
	go func() { done <- i.InstallWorkspaceDependencies(ctx) }()

	waitFor(t, "the waiting message", func() bool { return out.String() != "" })
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the install did not stop when cancelled")
	}

	modAfter, err := os.ReadFile(l.modFile)
	if err != nil {
		t.Fatal(err)
	}
	cacheAfter, err := os.ReadFile(l.lockFile())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(modBefore, modAfter) {
		t.Errorf("mod.pp changed:\n%s", modAfter)
	}
	if !bytes.Equal(cacheBefore, cacheAfter) {
		t.Errorf("the cache changed:\n%s", cacheAfter)
	}
	assertNoTempFiles(t, l.path)
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err == nil && strings.HasSuffix(d.Name(), ".tmp") {
			t.Errorf("temporary file left behind: %s", path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// the lock is dropped when the command ends, whatever the outcome: a failed install does not hold up the next
func TestLockIsReleasedAfterAFailedInstall(t *testing.T) {
	l := newModLocation(t)
	i := l.installer(t, "install", filepath.Join(l.path, "does-not-exist"))
	_ = i.InstallWorkspaceDependencies(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	unlock, err := lockModLocation(ctx, l.path)
	if err != nil {
		t.Fatalf("the lock is still held after the command ended: %v", err)
	}
	unlock()
}

// a dry run writes nothing, so it does not wait for a command that is writing
func TestDryRunDoesNotWait(t *testing.T) {
	l := newModLocation(t, "alpha")
	holdLock(t, l)
	i := l.installer(t, "install", l.mods["alpha"])
	i.dryRun = true

	done := make(chan error, 1)
	go func() { done <- i.InstallWorkspaceDependencies(context.Background()) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("a dry run waited for the lock")
	}
	assertSameSet(t, "required mods", l.requiredMods(t), nil)
}

// mod.pp keeps its permissions, and stays a symlink, through an install
func TestInstallKeepsModFilePermissionsAndSymlink(t *testing.T) {
	l := newModLocation(t, "alpha")
	if err := os.Chmod(l.modFile, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := l.installer(t, "install", l.mods["alpha"]).InstallWorkspaceDependencies(context.Background()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(l.modFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mod.pp permissions = %v, want 0600", info.Mode().Perm())
	}

	// now with mod.pp a link to a file elsewhere
	l2 := newModLocation(t, "beta")
	real := filepath.Join(filepath.Dir(l2.path), "real-mod.pp")
	content, err := os.ReadFile(l2.modFile)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, real, string(content))
	if err := os.Remove(l2.modFile); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, l2.modFile); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	if err := l2.installer(t, "install", l2.mods["beta"]).InstallWorkspaceDependencies(context.Background()); err != nil {
		t.Fatal(err)
	}
	linkInfo, err := os.Lstat(l2.modFile)
	if err != nil {
		t.Fatal(err)
	}
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("mod.pp is no longer a symlink")
	}
	written, err := os.ReadFile(real)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "beta") {
		t.Errorf("the file the link points to was not updated:\n%s", written)
	}
}
