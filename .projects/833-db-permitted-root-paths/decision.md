# Decisions: #833 PERMITTED_ROOT_PATHS also covers databases and in-process SQL

No decisions were pre-locked; the planner chose these from the issue, #832's code and the hosts'
dependencies. Executors do not re-open them; a fact that contradicts one is a `BLOCKED-task-K.md`.

1. **One checker: #832's `permitted_paths` package.** Every path test calls
   `permitted_paths.Check` (and opens or writes the resolved path it returns), `Enabled()` gates
   every new code path, `Roots()` / `Setting()` feed the DuckDB settings and the messages. Nothing
   reads `<APP>_PERMITTED_ROOT_PATHS` again. A bad setting (`Enabled()` true, `Check` erroring)
   fails closed: the value is refused and a DuckDB open fails. If the DuckDB lock needs each root's
   symlink-resolved form, add one exported helper to `permitted_paths` (for example
   `ResolvedRoots()`, wrapping the existing `resolveRoots`); do not copy the resolver.

2. **The mod-content checks live in `backend/permitted_paths.go`** as
   `ConfineModConnectionString(cs string) (string, error)`. `backend` already owns the prefixes
   (`DuckDBConnectionStringPrefix`, `SqliteConnectionStringPrefix`,
   `DuckDBInitConnectionStringPrefix`, `postgresConnectionStringPrefixes`,
   `mysqlConnectionStringPrefix`) and must not import `parse`, `modconfig` or `connection`
   (`connection` imports `backend`). `permitted_paths` cannot hold it: `backend` will import
   `permitted_paths`, so the reverse is a cycle.

3. **Provenance decides which refusals apply.** "From mod content" means a plain-string `database`
   value decoded by `parse.ResolveConnectionString` (resource and mod `database`, which hosts call
   for every resource) and the `database` override in a `require { mod … { database = … } }` block
   (`ModVersionConstraint.Initialise`). Those go through `ConfineModConnectionString`. A
   `connection.<type>.<name>` reference, a connection object, a CLI argument (`inputvars`) and a
   host-built string (tailpipe's `duckdbinit://`) are the host's own and are not subject to the
   mod-content refusals; the runtime lock (decisions 6 and 7) still applies to every DuckDB and
   SQLite open.

4. **DuckDB and SQLite values from mod content, with the setting set:**
   - The remainder after the prefix carrying `?` is refused (options on the value).
   - An in-memory value is allowed: empty, `:memory:` or `:memory:<name>`. The runtime lock covers it.
   - A remainder with a URI or scheme form (`file:`, `md:`, `s3://`, any `<scheme>:` before the
     first `/`) is refused: it is not a plain file path.
   - Anything else is a file path, made absolute against the process working directory (as the
     driver opens it), passed to `permitted_paths.Check`, and the returned connection string is
     rewritten to `<prefix><resolved path>` so the file opened is the file checked.
   - `duckdbinit://` from mod content is refused outright, whatever follows.

5. **Postgres and MySQL values from mod content, with the setting set:**
   - Postgres (`postgres://`, `postgresql://`): parse with `net/url`, never `pgconn.ParseConfig`
     (it reads the passfile, service file and certificates while parsing). Each of `sslcert`,
     `sslkey`, `sslrootcert`, `passfile`, `servicefile` (key compared case-insensitively, every
     value of a repeated key) must pass `permitted_paths.Check`; the value is rewritten to the
     resolved path. Other parameters pass through. An unparseable URL is refused.
   - MySQL (`mysql://`): an `allowAllFiles` parameter (case-insensitive key) is refused whatever its
     value.
   - Error messages name the option and its path, never the whole connection string: a Postgres or
     MySQL string can carry a password.

6. **Every DuckDB the backend opens is locked, per connection, with no driver import.** When
   `Enabled()`, `DuckDBBackend.Connect` and `DuckDBInitBackend.Connect` open through a
   driver-agnostic connector (`backend/conn_init_connector.go`, modelled on
   `backend/pgx_connector.go`): it takes the registered driver by name (its `driver.DriverContext`
   connector when it has one, else `Open(dsn)`), and runs an init function on every new
   `driver.Conn` before database/sql hands it out. The DuckDB init, on each connection, leaves in
   force: `allowed_directories` = every root (and its resolved form) with a trailing separator, so
   `/data/a` never admits `/data/ab`; `enable_external_access = false`;
   `lock_configuration = true`. It verifies the three are in force (reading `current_setting`), and
   any failure fails the connection and so the open: `Connect` closes the pool and returns the
   error. Because the three are DuckDB global settings shared by every connection of one instance,
   a later connection verifies rather than re-sets them. Paths in SQL are single-quote escaped.

   **Order.** pipe-fittings' own extension setup (`set extension_directory`, `install`/`load` of
   `constants.DuckDbExtensions`) runs inside the init before the three settings, because DuckDB
   refuses extension install and load once external access is off; the duckdbinit script runs
   after them. So no SQL from a mod, a host or an init script ever reaches an unlocked connection,
   which is what the issue's "before any other SQL" protects. Settings in a host's DSN are applied
   by the driver before the init; the init then overrides them, and a DSN that pre-locks the
   configuration makes the init fail, so the open fails closed.

7. **SQLite ATTACH outside the folders is refused per connection.** Task 3 first proves (test,
   unset) that SQLite SQL can ATTACH a file outside. Then, when `Enabled()`,
   `SqliteBackend.Connect` opens through the same connector, and the init registers the driver's
   authorizer through a locally declared interface
   (`interface{ RegisterAuthorizer(func(int, string, string, string) int) }`, which
   `mattn/go-sqlite3`'s `*SQLiteConn` has). The authorizer denies `SQLITE_ATTACH` (24) whose file
   name is a URI (`file:`) or fails `permitted_paths.Check`; an empty name and `:memory:` stay
   allowed; every other action is allowed. A driver connection without that method fails the open.

8. **Test drivers are test-only, pinned to the hosts' versions.** powerpipe and tailpipe (their
   develop `go.mod` today) use `github.com/marcboeker/go-duckdb/v2 v2.4.3` and
   `github.com/mattn/go-sqlite3 v1.14.52`. Those exact versions are added to `go.mod`, imported
   only from `_test.go` files carrying `//go:build cgo`, so a library consumer never builds cgo
   because of pipe-fittings, and MVS never raises a host's driver version. No non-test package
   imports a driver: `go list -deps ./...` stays free of both. Tests need no network: the DuckDB
   extension install is a seam the tests empty.

9. **Unset means unchanged.** With `Enabled()` false: `ConfineModConnectionString` returns its
   input unchanged and refuses nothing; each `Connect` takes today's exact `sql.Open` path and
   statements. Each runtime task proves its unset arm.

10. **Messages follow #832.** A refusal wraps `*permitted_paths.OutsideError` (so callers can
    `errors.As` it) or states the reason with the setting's name from `permitted_paths.Setting()`,
    for example `database: option "?…" is not allowed while POWERPIPE_PERMITTED_ROOT_PATHS is set`.

11. **Ship as a stack.** `gh stack --help` works on this worker; each task is one layer, landing in
    order.
