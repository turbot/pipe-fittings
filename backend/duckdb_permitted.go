package backend

import (
	"context"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/turbot/pipe-fittings/v2/sperr"
)

// connectDuckDBPermitted opens the DuckDB dsn locked to the permitted folders: every connection the pool opens
// runs duckDBPermittedInit before it is handed out. configure sets the pool's limits. One connection is opened
// here, so a lock that cannot be put in force fails the open, not the first query; on any failure the pool is
// closed and the error returned.
func connectDuckDBPermitted(ctx context.Context, dsn string, configure func(*sql.DB)) (*sql.DB, error) {
	init, err := duckDBPermittedInit(duckDBFilePath(dsn))
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to duckdb backend")
	}
	connector, err := NewConnInitConnector("duckdb", dsn, init)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to duckdb backend")
	}
	db := sql.OpenDB(connector)
	configure(db)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, sperr.WrapWithMessage(err, "could not connect to duckdb backend")
	}
	return db, nil
}

// duckDBPermittedInit returns the init that locks each DuckDB connection to the permitted folders.
//
// The settings it leaves in force are DuckDB globals, shared by every connection to one database instance:
//   - temp_directory: a folder of its own inside the first permitted folder it can be made in (duckDBTempDir), so DuckDB's temp
//     files, and the folder it adds to allowed_directories for them, lie inside the folders; or empty when no
//     permitted folder can hold one, so DuckDB does not spill to disk and adds no folder;
//   - allowed_directories: every permitted folder, as written and with its symlinks resolved, each ending in a
//     separator, as DuckDB matches them as prefixes (so /data/a never admits /data/ab);
//   - allowed_paths: empty, so no single file outside the folders is let through (a DSN could set it);
//   - allow_persistent_secrets = false: DuckDB keeps persistent secrets in <HOME>/.duckdb/stored_secrets,
//     outside the folders, and external access does not cover them, so SQL could otherwise write a secret
//     there, or replace or drop one the user stored;
//   - enable_external_access = false: no file outside those folders is read, written or attached, and no
//     extension is installed or loaded;
//   - lock_configuration = true: no later SQL can change any setting back.
//
// Order. On the instance's first connection (lock_configuration still false) allow_persistent_secrets is
// turned off first, as DuckDB will not change it once anything has used its secrets; then pipe-fittings' own
// extension setup runs, because DuckDB refuses to install or load an extension once external access is off; then
// the temp folder and the two lists, which DuckDB will not change once external access is off; then external access; then the
// lock. Only after this does any other SQL run: the caller's queries, and a duckdbinit script, which its
// Connect runs after the pool's first connection is locked. A later connection finds the lock in force and
// does not set anything. Every connection, first or later, then reads the settings back and fails unless all
// are as required, so a DSN that set lock_configuration before the init could (the driver applies DSN
// settings when it opens the instance) fails the open instead of leaving it unlocked.
//
// DuckDB itself adds its own files when external access is turned off: its temp_directory (where it spills to
// disk) to allowed_directories, and the database file and its .wal to allowed_paths. The check accepts those
// only as duckDBOwnFiles says, and fails unless temp_directory lies inside the folders or is empty, so neither DuckDB's
// default (.tmp in the working folder, or <file>.tmp, wherever those lie) nor a DSN's temp_directory can open
// a folder outside them to SQL.
//
// DuckDB matches allowed_directories as text and follows symlinks, so a symlink inside the folders that leads
// outside them would let SQL read and write the files it leads to. SQL cannot make a symlink, so the folders
// are searched once here, before any connection opens, and a symlink that leads outside fails the open
// (duckDBOutwardSymlink). The temp folder lies inside a permitted folder, so that search covers it.
// dbPath is the database file the DSN opens, or "" for an in-memory database.
func duckDBPermittedInit(dbPath string) (ConnInitFunc, error) {
	roots, err := permitted_paths.ResolvedRoots()
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, fmt.Errorf("%s names no folder", permitted_paths.Setting())
	}
	if err := duckDBOutwardSymlink(permitted_paths.Roots()); err != nil {
		return nil, err
	}
	dirs := duckDBAllowedDirectories(roots)
	tempDir, err := duckDBTempDir(roots)
	if err != nil {
		return nil, err
	}
	own := duckDBOwnFiles{dbPath: dbPath}

	return func(ctx context.Context, conn driver.Conn) error {
		locked, err := duckDBSettingBool(ctx, conn, "lock_configuration")
		if err != nil {
			return fmt.Errorf("could not read the DuckDB configuration lock: %w", err)
		}
		var setErr error
		if !locked {
			setErr = duckDBLock(ctx, conn, tempDir, dirs)
		}
		if err := duckDBVerifyLock(ctx, conn, dirs, own); err != nil {
			// another connection to the same instance may have locked it first, so a failed set is only
			// reported when the lock is not in force after it
			if setErr != nil {
				return fmt.Errorf("could not limit DuckDB to the folders %s permits: %w", permitted_paths.Setting(), setErr)
			}
			return fmt.Errorf("DuckDB is not limited to the folders %s permits: %w", permitted_paths.Setting(), err)
		}
		return nil
	}, nil
}

// duckDBOutwardSymlink walks every permitted folder, with its symlinks resolved, and returns an error naming
// the first symlink in it that leads outside the folders (whether or not what it leads to exists), or that
// cannot be followed. A symlink that leads inside is not followed, as the folder it leads to is walked in its
// own right.
func duckDBOutwardSymlink(roots []string) error {
	for _, root := range roots {
		resolved, err := permitted_paths.Check(root)
		if err != nil {
			return err
		}
		if err := filepath.WalkDir(resolved, duckDBSymlinkWalk(resolved)); err != nil {
			return err
		}
	}
	return nil
}

// duckDBSymlinkWalk returns the WalkDirFunc duckDBOutwardSymlink walks the permitted folder resolved with.
//
// What DuckDB cannot read through is skipped, as it cannot lead DuckDB anywhere: a folder this process cannot
// enter, and a symlink loop whose first step stays inside the folders. A folder that does not exist, or no
// longer does, has nothing in it: DuckDB removes its own temp folder (under duckDBTempDirName) when a database
// closes, which may be while this walk runs. Any other error fails, so a folder that can be entered but not
// listed (where DuckDB could reach a symlink by name) fails, as it cannot be shown free of such a symlink.
func duckDBSymlinkWalk(resolved string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				if path == resolved {
					return fs.SkipAll
				}
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if d != nil && d.IsDir() && errors.Is(err, fs.ErrPermission) && !duckDBCanEnter(path) {
				return fs.SkipDir
			}
			return fmt.Errorf("could not search %s for symlinks that lead outside the folders %s permits: %w", path, permitted_paths.Setting(), err)
		}
		if d.Type()&os.ModeSymlink == 0 {
			return nil
		}
		_, err = permitted_paths.Check(path)
		if errors.Is(err, permitted_paths.ErrSymlinkLoop) && duckDBLinkStaysInside(path) {
			// DuckDB cannot read through a loop either
			return nil
		}
		if err != nil {
			var outside *permitted_paths.OutsideError
			if errors.As(err, &outside) || errors.Is(err, permitted_paths.ErrSymlinkLoop) {
				return fmt.Errorf("%s is a symlink that leads outside the folders %s permits, which DuckDB would follow: remove it, or add the folder it leads to to the setting", path, permitted_paths.Setting())
			}
			// a symlink that cannot be followed cannot be shown to stay inside
			return fmt.Errorf("could not follow the symlink %s, so %s cannot be applied to DuckDB: %w", path, permitted_paths.Setting(), err)
		}
		return nil
	}
}

// duckDBCanEnter reports whether this process can look up a name in the folder dir, which is what DuckDB needs
// to read anything under it. Looking up dir's own "." entry needs exactly that.
func duckDBCanEnter(dir string) bool {
	_, err := os.Lstat(dir + string(filepath.Separator) + ".")
	return err == nil
}

// duckDBLinkStaysInside reports whether the symlink link's own target, read as written from the folder holding
// it, lies inside the folders. A loop that steps outside could be ended there, by a change outside the folders,
// to lead anywhere.
func duckDBLinkStaysInside(link string) bool {
	target, err := os.Readlink(link)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(link), target)
	}
	return permitted_paths.Within(target)
}

// duckDBTempDirName is the folder, inside a permitted folder, that holds each DuckDB database's temp folder
const duckDBTempDirName = ".duckdb_tmp"

// duckDBTempDir returns the temp folder for one DuckDB database: a folder of its own, not yet made, under
// duckDBTempDirName in the first of roots where one can be made. Folders that exist are tried first, in order,
// then the rest, each made if missing; a root this process cannot write to is passed over. When none will do
// (every permitted folder is read-only, say) it returns "": the lock then sets temp_directory empty, so DuckDB
// does not spill to disk at all and a query that outgrows memory fails, rather than the open failing for a
// database that can be read without spilling. A duckDBTempDirName that leads outside the folders still fails
// the open. Each database has its own temp folder, as DuckDB names its temp files the same way in every
// process and removes them when it closes. DuckDB makes the folder when it first spills to disk, and removes
// it again when it closes, as it made it.
func duckDBTempDir(roots []string) (string, error) {
	var existing, missing []string
	for _, r := range roots {
		if fi, err := os.Stat(r); err == nil && fi.IsDir() {
			existing = append(existing, r)
		} else {
			missing = append(missing, r)
		}
	}
	var firstErr error
	for _, root := range append(existing, missing...) {
		dir, err := duckDBTempDirIn(root)
		if err == nil {
			return dir, nil
		}
		var outside *duckDBTempDirOutsideError
		if errors.As(err, &outside) {
			// a symlink that leads out fails the open, whatever a later root holds
			return "", err
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		slog.Debug("no permitted folder can hold a DuckDB temp folder, so DuckDB will not spill to disk", "error", firstErr)
	}
	return "", nil
}

// duckDBTempDirIn returns a temp folder name under duckDBTempDirName in root, once it has shown that the folder
// can be made there (it makes it and removes it again, so DuckDB still makes and removes it itself)
func duckDBTempDirIn(root string) (string, error) {
	parent := filepath.Join(root, duckDBTempDirName)
	if !isPermitted(parent) {
		// a symlink made at that name since the folders were searched, whether or not what it leads to exists
		return "", &duckDBTempDirOutsideError{parent: parent}
	}
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return "", fmt.Errorf("could not make a DuckDB temp folder inside the folders %s permits: %w", permitted_paths.Setting(), err)
	}
	dir := filepath.Join(parent, fmt.Sprintf("%d-%s", os.Getpid(), rand.Text()))
	if !isPermitted(dir) {
		// a symlink made at that name since the folders were searched
		return "", &duckDBTempDirOutsideError{parent: parent}
	}
	if err := os.Mkdir(dir, 0o700); err != nil {
		return "", fmt.Errorf("could not make a DuckDB temp folder inside the folders %s permits: %w", permitted_paths.Setting(), err)
	}
	if err := os.Remove(dir); err != nil {
		return "", fmt.Errorf("could not remove the DuckDB temp folder %s: %w", dir, err)
	}
	return dir, nil
}

// duckDBTempDirOutsideError is a duckDBTempDirName folder that does not lie inside the folders: a symlink
// made there since the folders were searched
type duckDBTempDirOutsideError struct{ parent string }

func (e *duckDBTempDirOutsideError) Error() string {
	return fmt.Sprintf("%s does not lie inside the folders %s permits", e.parent, permitted_paths.Setting())
}

// duckDBLock turns off persistent secrets, runs the extension setup and sets the lock's settings, in that order
func duckDBLock(ctx context.Context, conn driver.Conn, tempDir string, dirs []string) error {
	exec := func(query string) error { return execConn(ctx, conn, query) }
	// DuckDB refuses to change this once anything has used its secrets, so it goes before the extension setup
	if err := exec("SET allow_persistent_secrets = false"); err != nil {
		return fmt.Errorf("SET allow_persistent_secrets = false: %w", err)
	}
	if err := duckDbExtensionSetup(exec); err != nil {
		return err
	}
	for _, stmt := range []string{
		"SET temp_directory = " + duckDBQuote(tempDir),
		fmt.Sprintf("SET allowed_directories = %s", duckDBStringList(dirs)),
		"SET allowed_paths = []",
		"SET enable_external_access = false",
		"SET lock_configuration = true",
	} {
		if err := exec(stmt); err != nil {
			return fmt.Errorf("%s: %w", stmt, err)
		}
	}
	return nil
}

// duckDBVerifyLock reads the lock's settings back and fails unless every one is as required
func duckDBVerifyLock(ctx context.Context, conn driver.Conn, dirs []string, own duckDBOwnFiles) error {
	external, err := duckDBSettingBool(ctx, conn, "enable_external_access")
	if err != nil {
		return err
	}
	if external {
		return errors.New("enable_external_access is true")
	}
	locked, err := duckDBSettingBool(ctx, conn, "lock_configuration")
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("lock_configuration is false")
	}
	persistent, err := duckDBSettingBool(ctx, conn, "allow_persistent_secrets")
	if err != nil {
		return err
	}
	if persistent {
		return errors.New("allow_persistent_secrets is true")
	}
	got, err := duckDBSettingList(ctx, conn, "allowed_directories")
	if err != nil {
		return err
	}
	tempDir, err := duckDBSettingString(ctx, conn, "temp_directory")
	if err != nil {
		return err
	}
	if tempDir != "" && !isPermitted(tempDir) {
		return fmt.Errorf("temp_directory is %q, which is outside the folders", tempDir)
	}
	if !duckDBListIs(got, dirs, func(dir string) bool { return own.tempDirAllowed(dir, tempDir) }) {
		return fmt.Errorf("allowed_directories is %q, want %q", got, dirs)
	}
	paths, err := duckDBSettingList(ctx, conn, "allowed_paths")
	if err != nil {
		return err
	}
	if !duckDBListIs(paths, nil, own.pathAllowed) {
		return fmt.Errorf("allowed_paths is %q, want none outside the folders but the database's own files", paths)
	}
	return nil
}

// duckDBAllowedDirectories returns each root ending in exactly one separator, without repeats
func duckDBAllowedDirectories(roots []string) []string {
	sep := string(filepath.Separator)
	seen := map[string]bool{}
	var dirs []string
	for _, root := range roots {
		dir := strings.TrimRight(root, sep) + sep
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// duckDBStringList writes a DuckDB list literal of strings, each single-quote escaped
func duckDBStringList(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = duckDBQuote(v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// duckDBQuote writes a DuckDB string literal, doubling each single quote
func duckDBQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// duckDBOwnFiles says which of the entries DuckDB adds itself to its allow lists are accepted
type duckDBOwnFiles struct {
	// dbPath is the database file, or "" for an in-memory database
	dbPath string
}

// tempDirAllowed reports whether dir, an allowed_directories entry, is DuckDB's own temp directory and that lies
// inside the folders. A temp directory anywhere else (DuckDB's default .tmp in the working folder or <file>.tmp,
// or one a DSN set) would open that folder to SQL.
func (o duckDBOwnFiles) tempDirAllowed(dir, tempDir string) bool {
	if tempDir == "" || duckDBNormPath(dir) != duckDBNormPath(tempDir) {
		return false
	}
	return isPermitted(tempDir)
}

// pathAllowed reports whether path, an allowed_paths entry, is the database file or its write-ahead log, which
// DuckDB has open already, or lies inside the folders
func (o duckDBOwnFiles) pathAllowed(path string) bool {
	if o.dbPath != "" && (sameFilePath(path, o.dbPath) || sameFilePath(path, o.dbPath+".wal")) {
		return true
	}
	return isPermitted(path)
}

func isPermitted(path string) bool {
	_, err := permitted_paths.Check(path)
	return err == nil
}

func sameFilePath(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	return errA == nil && errB == nil && absA == absB
}

// duckDBFilePath returns the database file a DuckDB DSN opens, as go-duckdb reads it, or "" for an in-memory
// database
func duckDBFilePath(dsn string) string {
	if i := strings.Index(dsn, "?"); i >= 0 {
		dsn = dsn[:i]
	}
	if dsn == ":memory:" {
		return ""
	}
	return dsn
}

// duckDBNormPath writes a folder the way DuckDB stores it in allowed_directories: forward slashes, ending in one
func duckDBNormPath(p string) string {
	p = strings.ReplaceAll(p, `\`, "/")
	return strings.TrimRight(p, "/") + "/"
}

// duckDBListIs reports whether the folders DuckDB holds (got) include every one of want, and any other is one
// extra accepts
func duckDBListIs(got, want []string, extra func(string) bool) bool {
	wantSet := map[string]bool{}
	for _, p := range want {
		wantSet[duckDBNormPath(p)] = true
	}
	gotSet := map[string]bool{}
	for _, p := range got {
		gotSet[duckDBNormPath(p)] = true
		if !wantSet[duckDBNormPath(p)] && !extra(p) {
			return false
		}
	}
	for p := range wantSet {
		if !gotSet[p] {
			return false
		}
	}
	return true
}

func duckDBSettingBool(ctx context.Context, conn driver.Conn, name string) (bool, error) {
	v, err := queryConnValue(ctx, conn, fmt.Sprintf("SELECT current_setting(%s)", duckDBQuote(name)))
	if err != nil {
		return false, fmt.Errorf("could not read %s: %w", name, err)
	}
	switch b := v.(type) {
	case bool:
		return b, nil
	case string:
		switch strings.ToLower(b) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
	}
	return false, fmt.Errorf("could not read %s: unexpected value %v", name, v)
}

func duckDBSettingString(ctx context.Context, conn driver.Conn, name string) (string, error) {
	v, err := queryConnValue(ctx, conn, fmt.Sprintf("SELECT current_setting(%s)::VARCHAR", duckDBQuote(name)))
	if err != nil {
		return "", fmt.Errorf("could not read %s: %w", name, err)
	}
	switch str := v.(type) {
	case nil:
		return "", nil
	case string:
		return str, nil
	}
	return "", fmt.Errorf("could not read %s: unexpected value %v", name, v)
}

func duckDBSettingList(ctx context.Context, conn driver.Conn, name string) ([]string, error) {
	v, err := queryConnValue(ctx, conn, fmt.Sprintf("SELECT current_setting(%s)", duckDBQuote(name)))
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", name, err)
	}
	switch l := v.(type) {
	case nil:
		return nil, nil
	case []string:
		return l, nil
	case []any:
		out := make([]string, 0, len(l))
		for _, item := range l {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("could not read %s: unexpected entry %v", name, item)
			}
			out = append(out, s)
		}
		return out, nil
	}
	return nil, fmt.Errorf("could not read %s: unexpected value %v", name, v)
}
