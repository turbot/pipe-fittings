package backend

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/fs"
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
//   - allowed_directories: every permitted folder, as written and with its symlinks resolved, each ending in a
//     separator, as DuckDB matches them as prefixes (so /data/a never admits /data/ab);
//   - allowed_paths: empty, so no single file outside the folders is let through (a DSN could set it);
//   - enable_external_access = false: no file outside those folders is read, written or attached, and no
//     extension is installed or loaded;
//   - lock_configuration = true: no later SQL can change any setting back.
//
// Order. On the instance's first connection (lock_configuration still false) pipe-fittings' own extension
// setup runs first, because DuckDB refuses to install or load an extension once external access is off; then
// the two lists, which DuckDB will not change once external access is off; then external access; then the
// lock. Only after this does any other SQL run: the caller's queries, and a duckdbinit script, which its
// Connect runs after the pool's first connection is locked. A later connection finds the lock in force and
// does not set anything. Every connection, first or later, then reads the settings back and fails unless all
// are as required, so a DSN that set lock_configuration before the init could (the driver applies DSN
// settings when it opens the instance) fails the open instead of leaving it unlocked.
//
// DuckDB itself adds its own files when external access is turned off: its temp_directory (where it spills to
// disk) to allowed_directories, and the database file and its .wal to allowed_paths. The check accepts those
// only as duckDBOwnFiles says, so a DSN that pointed temp_directory elsewhere cannot open a folder to SQL.
//
// DuckDB matches allowed_directories as text and follows symlinks, so a symlink inside the folders that leads
// outside them would let SQL read and write the files it leads to. SQL cannot make a symlink, so the folders
// are searched once here, before any connection opens, and a symlink that leads outside fails the open
// (duckDBOutwardSymlink).
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
	own := duckDBOwnFiles{dbPath: dbPath}

	return func(ctx context.Context, conn driver.Conn) error {
		locked, err := duckDBSettingBool(ctx, conn, "lock_configuration")
		if err != nil {
			return fmt.Errorf("could not read the DuckDB configuration lock: %w", err)
		}
		var setErr error
		if !locked {
			setErr = duckDBLock(ctx, conn, dirs)
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
//
// What DuckDB cannot read through is skipped, as it cannot lead DuckDB anywhere: a folder this process cannot
// enter, and a symlink loop whose first step stays inside the folders. A folder that does not exist has nothing
// in it. Any other error fails, so a folder that can be entered but not listed (where DuckDB could reach a
// symlink by name) fails, as it cannot be shown free of such a symlink.
func duckDBOutwardSymlink(roots []string) error {
	for _, root := range roots {
		resolved, err := permitted_paths.Check(root)
		if err != nil {
			return err
		}
		err = filepath.WalkDir(resolved, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if path == resolved && errors.Is(err, fs.ErrNotExist) {
					return fs.SkipAll
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
		})
		if err != nil {
			return err
		}
	}
	return nil
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

// duckDBLock runs the extension setup and sets the lock's settings, in that order
func duckDBLock(ctx context.Context, conn driver.Conn, dirs []string) error {
	exec := func(query string) error { return execConn(ctx, conn, query) }
	if err := duckDbExtensionSetup(exec); err != nil {
		return err
	}
	for _, stmt := range []string{
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
	got, err := duckDBSettingList(ctx, conn, "allowed_directories")
	if err != nil {
		return err
	}
	tempDir, err := duckDBSettingString(ctx, conn, "temp_directory")
	if err != nil {
		return err
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

// duckDBDefaultTempDir is the temp_directory DuckDB uses for an in-memory database: .tmp in the working folder
const duckDBDefaultTempDir = ".tmp"

// duckDBOwnFiles says which of the entries DuckDB adds itself to its allow lists are accepted
type duckDBOwnFiles struct {
	// dbPath is the database file, or "" for an in-memory database
	dbPath string
}

// tempDirAllowed reports whether dir, an allowed_directories entry, is DuckDB's own temp directory and that is
// where DuckDB puts it by default (.tmp for an in-memory database, <file>.tmp for a file) or inside the folders.
// A temp_directory set anywhere else, for example by a DSN, would open that folder to SQL.
func (o duckDBOwnFiles) tempDirAllowed(dir, tempDir string) bool {
	if tempDir == "" || duckDBNormPath(dir) != duckDBNormPath(tempDir) {
		return false
	}
	if tempDir == duckDBDefaultTempDir || (o.dbPath != "" && sameFilePath(tempDir, o.dbPath+".tmp")) {
		return true
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
