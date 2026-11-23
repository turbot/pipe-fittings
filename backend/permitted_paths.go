package backend

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/turbot/pipe-fittings/v2/permitted_paths"
)

// postgresFileOptions are the connection string options that make the Postgres driver read a file
var postgresFileOptions = []string{"sslcert", "sslkey", "sslrootcert", "passfile", "servicefile"}

// ConfineModConnectionString limits a database connection string that mod content supplies to what
// <APP>_PERMITTED_ROOT_PATHS permits. It returns the string to use, which may differ from cs, or an error
// that says what is refused. With the setting unset it returns cs unchanged and refuses nothing.
//
// It is for every value a mod or a .ppc file can supply: a mod's or a resource's `database`, a `require`
// block's `database` override, and every .ppc connection (a mod can ship its own .ppc, and a connection
// cannot tell a .ppc in the mod folder from one in the install folder). It is not applied to a string the
// host builds in its own code (a command-line value, or the string `tailpipe connect` returns): the host
// chose those, not the mod, and a host that keeps its own database files lists their folders in the setting.
//
// With the setting set:
//   - DuckDB and SQLite: an in-memory value passes; anything with `?` options or a URI or scheme form
//     (`file:`, `md:`, `s3://`) is refused; any other value is a file, checked, and rewritten to its
//     absolute path with symlinks resolved, so the file opened is the file checked
//   - duckdbinit:// is refused whatever follows
//   - Postgres: each of sslcert, sslkey, sslrootcert, passfile and servicefile is checked and rewritten
//     to its resolved path; a string that cannot be parsed is refused
//   - MySQL: an allowAllFiles option is refused whatever its value
//
// A refusal names the option or the path, never the whole string, which can carry a password.
func ConfineModConnectionString(cs string) (string, error) {
	if !permitted_paths.Enabled() {
		return cs, nil
	}
	trimmed := strings.TrimSpace(cs)
	switch {
	case IsDuckDBInitConnectionString(trimmed):
		return "", fmt.Errorf("database: %s values are not allowed while %s is set", DuckDBInitConnectionStringPrefix, permitted_paths.Setting())
	case IsDuckDBConnectionString(trimmed):
		return confineFileDatabase(cs, trimmed, DuckDBConnectionStringPrefix)
	case IsSqliteConnectionString(trimmed):
		return confineFileDatabase(cs, trimmed, SqliteConnectionStringPrefix)
	case IsPostgresConnectionString(trimmed):
		return confinePostgres(trimmed)
	case IsMySqlConnectionString(trimmed):
		return cs, confineMysql(trimmed)
	}
	// no known prefix: the backend rejects it later, as it does today
	return cs, nil
}

// confineFileDatabase checks the file a duckdb:// or sqlite:// value names
func confineFileDatabase(original, trimmed, prefix string) (string, error) {
	rest := strings.TrimPrefix(trimmed, prefix)
	if strings.Contains(rest, "?") {
		return "", fmt.Errorf("database: option \"?…\" is not allowed while %s is set", permitted_paths.Setting())
	}
	if rest == "" || strings.HasPrefix(rest, ":memory:") {
		return original, nil
	}
	if filepath.VolumeName(rest) == "" {
		// a scheme before the first slash (file:, md:, s3://) is not a plain file path
		before, _, _ := strings.Cut(rest, "/")
		if strings.Contains(before, ":") {
			return "", fmt.Errorf("database: %q is not a plain file path, which is all %s permits", rest, permitted_paths.Setting())
		}
	}
	resolved, err := permitted_paths.Check(rest)
	if err != nil {
		return "", fmt.Errorf("database: %w", err)
	}
	return prefix + resolved, nil
}

// confinePostgres checks the files a postgres:// value makes the driver read, and rewrites each to its
// resolved path. Only the query is rebuilt: every other part of the string is kept as written.
func confinePostgres(trimmed string) (string, error) {
	u, err := url.Parse(trimmed)
	if err != nil {
		// the error from url.Parse quotes the string, which can carry a password
		return "", fmt.Errorf("database: the postgres connection string cannot be parsed, so %s cannot be applied to it", permitted_paths.Setting())
	}
	if u.RawQuery == "" {
		return trimmed, nil
	}
	pairs := strings.Split(u.RawQuery, "&")
	for i, pair := range pairs {
		rawKey, rawValue, hasValue := strings.Cut(pair, "=")
		if !hasValue {
			continue
		}
		key, err := url.QueryUnescape(rawKey)
		if err != nil || !isPostgresFileOption(key) {
			if err != nil {
				// a key that cannot be decoded could still name one of the options
				return "", fmt.Errorf("database: an option of the postgres connection string cannot be parsed, so %s cannot be applied to it", permitted_paths.Setting())
			}
			continue
		}
		value, err := url.QueryUnescape(rawValue)
		if err != nil {
			return "", fmt.Errorf("database: option %q cannot be parsed, so %s cannot be applied to it", key, permitted_paths.Setting())
		}
		if value == "" {
			continue
		}
		resolved, err := permitted_paths.Check(value)
		if err != nil {
			return "", fmt.Errorf("database: option %q: %w", key, err)
		}
		pairs[i] = rawKey + "=" + url.QueryEscape(resolved)
	}
	// url.Parse splits the query at the first ?, so the string is rebuilt around it, never re-encoded
	at := strings.Index(trimmed, "?")
	return trimmed[:at+1] + strings.Join(pairs, "&") + trimmed[at+1+len(u.RawQuery):], nil
}

func isPostgresFileOption(key string) bool {
	for _, option := range postgresFileOptions {
		if strings.EqualFold(key, option) {
			return true
		}
	}
	return false
}

// confineMysql refuses an allowAllFiles option, which lets the server read any local file
func confineMysql(trimmed string) error {
	// the options follow the first ? after the last /, as the driver reads them
	rest := trimmed
	if i := strings.LastIndex(rest, "/"); i >= 0 {
		rest = rest[i:]
	}
	_, query, found := strings.Cut(rest, "?")
	if !found {
		return nil
	}
	for _, pair := range strings.Split(query, "&") {
		rawKey, _, _ := strings.Cut(pair, "=")
		key, err := url.QueryUnescape(rawKey)
		if err != nil {
			key = rawKey
		}
		if strings.EqualFold(key, "allowAllFiles") {
			return errors.New("database: option \"allowAllFiles\" is not allowed while " + permitted_paths.Setting() + " is set")
		}
	}
	return nil
}
