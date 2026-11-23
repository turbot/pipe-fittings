// Package permitted_paths owns the <APP>_PERMITTED_ROOT_PATHS setting (for example
// POWERPIPE_PERMITTED_ROOT_PATHS): a comma-separated list of absolute folders. When it is set, the readers
// that follow it read only inside those folders. When it is unset or empty there is no limit.
//
// The setting comes from the environment only and is read once, by Load at start-up (or on first use when
// an app never calls Load). A reader of a path that mod or config content names calls Check on the path
// before it opens it, and Filter on a listing before it returns it.
//
// The setting also covers databases. backend.ConfineModConnectionString checks every database value that mod
// or .ppc content supplies (every connection type that builds a Postgres, DuckDB, SQLite or MySQL string,
// steampipe included): a DuckDB or SQLite file must be inside the folders and takes no options (only an empty
// value or exactly :memory: is in-memory), duckdbinit:// is refused, the files a Postgres value names
// (sslcert, sslkey, sslrootcert, passfile) are checked, Postgres servicefile and service are refused (the
// service file can name files of its own), and MySQL allowAllFiles is refused. The backend then locks every
// DuckDB it opens to the folders (no access outside them, and the lock cannot be lifted), and refuses to open
// one while a symlink in the folders leads outside them, as DuckDB would follow it. It makes SQLite refuse an
// ATTACH of a file outside them, or of a name it cannot judge (an expression or a bound parameter). So SQL a
// mod runs cannot read, write, attach or install outside either. A host that opens its own DuckDB files (for
// example tailpipe's database and the folders its init script reads) must list their folders in the setting,
// or the lock refuses them.
package permitted_paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/turbot/pipe-fittings/v2/app_specific"
)

// state is what Load read from the environment
type state struct {
	// key is the name of the env var, for the messages
	key string
	// roots are the permitted folders, absolute and clean; empty when there is no limit
	roots []string
	// err is why the setting could not be read; every check fails while it is set
	err error
}

var (
	mu     sync.Mutex
	loaded *state
)

// Load reads the app's PERMITTED_ROOT_PATHS env var, replacing anything read before. An app calls it once at
// start-up and stops on the error: a relative entry is refused, and the error names it.
// A bad setting is also remembered, so a check made after it fails closed, never open.
func Load() error {
	mu.Lock()
	defer mu.Unlock()
	loaded = read(app_specific.EnvPermittedRootPaths)
	return loaded.err
}

// Reset forgets what Load read, so the next use reads the env var again. It is for tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	loaded = nil
}

func current() *state {
	mu.Lock()
	defer mu.Unlock()
	if loaded == nil {
		loaded = read(app_specific.EnvPermittedRootPaths)
	}
	return loaded
}

// Enabled reports whether the setting limits the paths that are read: it is set, or it is set to something
// that could not be read (so every check fails)
func Enabled() bool {
	s := current()
	return len(s.roots) > 0 || s.err != nil
}

// Setting returns the name of the env var (for example POWERPIPE_PERMITTED_ROOT_PATHS), for messages
func Setting() string {
	return current().key
}

// Roots returns the permitted folders as written, absolute and clean
func Roots() []string {
	return append([]string(nil), current().roots...)
}

// ResolvedRoots returns each permitted folder as written and again with its symlinks resolved, for a reader
// that limits paths by prefix itself (DuckDB's allowed_directories). A setting that could not be read is its
// error, so the reader fails closed. With no limit set it returns none.
func ResolvedRoots() ([]string, error) {
	s := current()
	if s.err != nil {
		return nil, s.err
	}
	return resolveRoots(s.roots), nil
}

func read(key string) *state {
	s := &state{key: key}
	if key == "" {
		// the app has not set its env var keys: nothing names the setting
		return s
	}
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return s
	}
	for _, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if !filepath.IsAbs(entry) {
			s.roots = nil
			s.err = fmt.Errorf("%s: %q is not an absolute path; every folder in the list must be absolute", key, entry)
			return s
		}
		s.roots = append(s.roots, filepath.Clean(entry))
	}
	if len(s.roots) == 0 {
		s.err = fmt.Errorf("%s: %q names no folder", key, value)
	}
	return s
}

// OutsideError is a path that lies outside every permitted folder
type OutsideError struct {
	// Path is the path as the caller gave it
	Path string
	// Setting is the name of the env var
	Setting string
}

// Reason says what is wrong without the path, for a caller that words the path itself (to hide a sensitive one)
func (e *OutsideError) Reason() string {
	return fmt.Sprintf("path is outside the folders %s permits", e.Setting)
}

func (e *OutsideError) Error() string {
	return fmt.Sprintf("%q: %s", e.Path, e.Reason())
}

// Check returns the path made absolute with every symlink resolved, or an error when that lies outside every
// permitted folder: an *OutsideError that names the path and the setting. The caller opens the returned path,
// not the one it passed, so the file it opens is the one that was checked. With no limit set the path is
// returned as it is.
//
// A path whose lookup fails outside the folders is refused as outside, so a lookup error never tells the
// caller what exists there. A path that cannot be followed inside them (a loop, or a failure under ..) is an error.
func Check(path string) (string, error) {
	s := current()
	if s.err != nil {
		return "", s.err
	}
	if len(s.roots) == 0 {
		return path, nil
	}
	outside := &OutsideError{Path: path, Setting: s.key}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	roots := resolveRoots(s.roots)

	// the path is judged where it resolves to, so a folder written through a symlink counts as the folder
	resolved, err := resolvePath(abs)
	if err != nil {
		// a lookup that failed outside the folders is reported as outside, so its error cannot tell the
		// caller what exists there
		var unresolvable *unresolvableError
		if errors.As(err, &unresolvable) {
			if !isWithinAny(roots, unresolvable.at) {
				return "", outside
			}
		} else if !isWithinAny(roots, abs) {
			return "", outside
		}
		return "", err
	}
	if !isWithinAny(roots, resolved) {
		return "", outside
	}
	return resolved, nil
}

// Filter returns the paths of a listing that lie inside the permitted folders, in order; the ones whose
// symlinks lead outside, or that cannot be resolved, are left out. With no limit set it returns them all.
func Filter(paths []string) []string {
	s := current()
	if s.err == nil && len(s.roots) == 0 {
		return paths
	}
	var kept []string
	for _, p := range paths {
		if _, err := Check(p); err == nil {
			kept = append(kept, p)
		}
	}
	return kept
}

// Within reports whether path, as written, lies inside a permitted folder, without looking at the file
// system: the check for a glob pattern, whose matches Filter then checks one by one. With no limit set it is true.
func Within(path string) bool {
	s := current()
	if s.err != nil {
		return false
	}
	if len(s.roots) == 0 {
		return true
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return isWithinAny(resolveRoots(s.roots), abs)
}

// resolveRoots returns each folder as written and again with its symlinks resolved, so a path counts as inside
// whether it is written through the link or the real folder. A folder that cannot be resolved counts as written.
func resolveRoots(roots []string) []string {
	all := make([]string, 0, len(roots)*2)
	for _, root := range roots {
		all = append(all, root)
		if resolved, err := resolvePath(root); err == nil && resolved != root {
			all = append(all, resolved)
		}
	}
	return all
}

func isWithinAny(roots []string, path string) bool {
	for _, root := range roots {
		if isWithin(root, path) {
			return true
		}
	}
	return false
}

// unresolvableError is a path resolvePath cannot follow past the part at, as the OS cannot either
type unresolvableError struct {
	at  string
	err error
}

func (e *unresolvableError) Error() string { return e.err.Error() }
func (e *unresolvableError) Unwrap() error { return e.err }

// maxSymlinks bounds the symlinks resolvePath follows, as the OS does, so a symlink loop ends
const maxSymlinks = 255

// resolvePath resolves the symlinks in an absolute, clean path one part at a time, following each symlink
// by its target whether or not that target exists. Once a part cannot be looked up (it is missing, or a
// folder on the way is not a folder or cannot be searched), the rest is appended as written: the OS cannot
// follow anything past that part either. So a missing file inside the folder can still be checked (and is
// reported missing by the reader itself), and a symlink that points outside is refused whether or not its
// target exists or can be reached.
func resolvePath(path string) (string, error) {
	sep := string(filepath.Separator)
	vol := filepath.VolumeName(path)
	resolved := vol + sep
	pending := strings.Split(path[len(vol):], sep)
	links := 0
	for len(pending) > 0 {
		part := pending[0]
		pending = pending[1:]
		switch part {
		case "", ".":
			continue
		case "..":
			resolved = filepath.Dir(resolved)
			continue
		}
		next := filepath.Join(resolved, part)
		fi, err := os.Lstat(next)
		if err != nil {
			// a ".." still to come (from a symlink target) would be cleaned away with the part the OS stops
			// at, leaving a path the OS can open and this function has not resolved: report the failure
			for _, rest := range pending {
				if rest == ".." {
					return "", &unresolvableError{at: next, err: err}
				}
			}
			return filepath.Join(append([]string{next}, pending...)...), nil
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			resolved = next
			continue
		}
		links++
		if links > maxSymlinks {
			return "", fmt.Errorf("too many levels of symbolic links")
		}
		target, err := os.Readlink(next)
		if err != nil {
			return "", err
		}
		if filepath.IsAbs(target) {
			targetVol := filepath.VolumeName(target)
			resolved = targetVol + sep
			target = target[len(targetVol):]
		}
		// a relative target is read from the folder holding the symlink, which is resolved
		pending = append(strings.Split(target, sep), pending...)
	}
	return resolved, nil
}

// isWithin reports whether path, once cleaned, is root or lies under it
func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
