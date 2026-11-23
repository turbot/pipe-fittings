---
status: pending
complexity: high
model: opus
depends_on: [1]
parallel_group: null
surface: none
kind: null
canonical_reference: backend/pgx_connector.go
family_pattern_contract: null
flaky: null
isolation: null
---
# Task 2: Lock every DuckDB the backend opens to the permitted folders

## Objective

With the setting set, every DuckDB opened by `DuckDBBackend.Connect` and
`DuckDBInitBackend.Connect` has `allowed_directories` = the permitted folders,
`enable_external_access = false` and `lock_configuration = true` in force on every pooled
connection before any caller SQL runs, and the open fails if that cannot be done. SQL on such a
database cannot read, write, attach or install anything outside the folders, and nothing in the
database value lifts the limit. Unset: today's open, unchanged.

## Context

- Issue #833 bullet 4 and success criteria 2 and 5. Rules: `decision.md` decisions 1, 6, 8, 9 (the
  order of extension setup, lock and init script is decided there; follow it).
- This repo imports no database driver (`go.mod` has only pgx); the host registers `"duckdb"`.
  So the per-connection hook must be driver-agnostic. Canonical shape: `backend/pgx_connector.go`
  (a `driver.Connector` wrapper running `AfterConnectFunc` on every new `driver.Conn`), used at
  `backend/postgres_backend.go:62-70` with `sql.OpenDB(connector)`.
- Today's opens:
  - `backend/duckdb_backend.go:33-50` `Connect`: `sql.Open("duckdb", b.connectionString)`, pool
    settings, then `installAndLoadDuckDbExtensions(db)` (:77-91) through the pool (so on one
    connection only).
  - `backend/duckdb_init_backend.go:50-77` `Connect`: `sql.Open("duckdb", "")`,
    `SetMaxOpenConns(1)`, extensions, then `executeInitScript` (:93-103) runs the script file's SQL.
- `marcboeker/go-duckdb/v2` v2.4.3 implements `driver.DriverContext`
  (`duckdb.go:39 func (Driver) OpenConnector(dsn string)`); its connector holds one DuckDB instance
  and each `Connect` is a new connection to it, so the three settings, being DuckDB globals, are
  shared by every connection of that instance. Its DSN `?key=value` parameters become DuckDB
  config before the instance opens.
- `allowed_directories` entries are prefixes: write each root with a trailing separator. Add each
  root's symlink-resolved form too (decision 1 says how, if `permitted_paths` needs a helper).
- Getting the registered driver without a driver import: `sql.Open(name, dsn)` does not connect,
  and `(*sql.DB).Driver()` returns the driver; for a `DriverContext` driver `sql.Open` already
  calls `OpenConnector`, so do not open the real file twice (a DuckDB file lock). One way:
  `sql.Open(name, "")` (in-memory), take `.Driver()`, close it, then build the wrapper from the
  driver with the real DSN. Choose and say why in a comment.
- Tests must not need the network: `installAndLoadDuckDbExtensions` installs `inet` from the
  internet into `filepaths.EnsurePipesDuckDbExtensionsDir()`. Make the extension list (or the
  setup function) a package-level seam the tests replace, and set `app_specific.InstallDir` to a
  temp folder in tests.
- Test helper pattern: `funcs/permitted_paths_test.go:80-104` (`setPermitted`).

## Files to modify

- `backend/duckdb_backend.go`
- `backend/duckdb_init_backend.go`
- `go.mod`
- `go.sum`
- `permitted_paths/permitted_paths.go` (only if decision 1's resolved-roots helper is needed)

## Files to create

- `backend/conn_init_connector.go`
- `backend/duckdb_permitted.go`
- `backend/duckdb_permitted_test.go`

## Sub-agent prompt

> Read `.projects/833-db-permitted-root-paths/task-2-duckdb-lockdown.md` (from the repo root) for
> full context, and `.projects/833-db-permitted-root-paths/decision.md` decisions 1, 6, 8 and 9,
> which are locked.
> Branch: `issue-833-db-permitted-root-paths`. Working dir: the repo root.
>
> Implement the task:
> 1. Create `backend/conn_init_connector.go`: a driver-agnostic `driver.Connector` that wraps the
>    driver registered under a name (its `OpenConnector(dsn)` when it is a `driver.DriverContext`,
>    else `Open(dsn)` per connection), runs an init `func(ctx context.Context, conn driver.Conn) error`
>    on every new connection, closes the connection and returns the error if init fails, and
>    closes the inner connector on `Close` when it is an `io.Closer`. Include a helper that runs one
>    statement and reads one value on a raw `driver.Conn` (through `driver.ExecerContext` /
>    `driver.QueryerContext`, or `Prepare` as a fallback). No driver package is imported.
> 2. Create `backend/duckdb_permitted.go` with the DuckDB init: on a connection where
>    `lock_configuration` is not yet true, run pipe-fittings' extension setup, then set
>    `allowed_directories` (every root and resolved root, trailing separator, single quotes
>    escaped), `enable_external_access = false`, `lock_configuration = true`; on every connection,
>    first or later, read the three back with `current_setting` and fail unless all are as
>    required. Write a comment stating the order and why (decision 6).
> 3. In both `Connect` methods, when `permitted_paths.Enabled()`, open through the connector with
>    that init (keep the pool settings and the duckdbinit `SetMaxOpenConns(1)`), run the duckdbinit
>    script only after, and on any failure close the pool and return the error. Force one
>    connection open inside `Connect` (for example `db.PingContext`) so a failed lock fails the
>    open, not the first query. When not enabled, leave today's code path byte-for-byte.
> 4. Add `github.com/marcboeker/go-duckdb/v2 v2.4.3` to `go.mod` (decision 8; exactly that
>    version) and `go mod tidy`. Import it only from `backend/duckdb_permitted_test.go`, which
>    carries `//go:build cgo`.
> 5. Write `backend/duckdb_permitted_test.go` (no network: replace the extension seam), with tests
>    whose names start `TestDuckDBPermitted`:
>    - `TestDuckDBPermittedLockdown`: setting = one temp folder, a CSV inside and one outside,
>      `WithConfig` with `MaxOpenConns` 4. Hold four connections at once (`db.Conn`) and on each
>      assert: reading the inside CSV works; `read_csv`/`read_text` of the outside file, `COPY … TO`
>      outside, `ATTACH` of an outside file and `INSTALL` from an outside path each fail;
>      `SET enable_external_access = true` and `SET allowed_directories = …` fail. Then let a
>      connection expire (`MaxConnLifeTime` tiny) and assert a fresh one is locked too.
>    - `TestDuckDBPermittedDatabaseFile`: a DuckDB file inside the folder opens and is queryable;
>      a sibling folder named with the root as a prefix (`<root>x`) is outside.
>    - `TestDuckDBPermittedValueCannotLift`: a host DSN with `?enable_external_access=true` is still
>      locked; a DSN with `?lock_configuration=true` makes `Connect` return an error.
>    - `TestDuckDBPermittedInit`: the duckdbinit backend with a script inside the folder runs it,
>      and a script line reading an outside file fails the open.
>    - `TestDuckDBPermittedUnset`: with the setting unset, reading the outside CSV works and
>      `current_setting('enable_external_access')` is true, as today.
>    - `TestDuckDBPermittedBadSetting`: a relative entry in the setting makes `Connect` fail.
> 6. Run the `## Verification` commands. The first DuckDB build links a large static library; let
>    it finish.
>
> Set the frontmatter `status: in-progress` when you begin and `status: done` when
> the acceptance criteria are met. Commit on the parent branch with subject
> `833-db-permitted-root-paths: Task 2 — duckdb-lockdown`, staging the files this task names by
> path, never `.projects/`.
>
> Where `AGENT_FLEET_DECIDE` is `on`, research a choice and decide it (`decide-by-default.md`);
> else on a design question write `BLOCKED-task-2.md` (`## Kind:` first) into the `Project dir:`
> from your dispatch prompt, set `status: blocked`, and exit.

## Acceptance criteria

- [ ] With the setting set, SQL on a DuckDB opened by either backend cannot read, write, attach or
      install outside the folders on any pooled connection, including one opened after another
      expired.
- [ ] No setting in the database value lifts the limit; a DSN that blocks the lock fails the open.
- [ ] A failed lock (or a bad setting) makes `Connect` return an error and leaves no open pool.
- [ ] With the setting unset, both `Connect` methods behave as today.
- [ ] No non-test package imports a DuckDB driver.
- [ ] Tests pass: `go test ./backend/ -run '^TestDuckDBPermitted'`.

## Verification

```bash
go build ./backend/... ./permitted_paths/...
go vet ./backend/ ./permitted_paths/
bash -o pipefail -c "go test ./backend/ -run '^TestDuckDBPermitted' -count=1 -v | grep -- '--- PASS: TestDuckDBPermittedLockdown'"
bash -o pipefail -c "go test ./backend/ -run '^TestDuckDBPermitted' -count=1 -v | grep -- '--- PASS: TestDuckDBPermittedUnset'"
go list -deps ./backend/... ./permitted_paths/... ./parse/... ./modconfig/... | grep -E 'go-duckdb|go-sqlite3' && exit 1 || true
```

`go list -deps` without `-test` lists only non-test dependencies, so it fails if a library package
imports the driver. The pipefail runs fail if any `TestDuckDBPermitted*` test fails or the named
one did not run (for example with cgo off).

## Out of scope

- The mod-content refusals (task 1); SQLite (task 3, which reuses `conn_init_connector.go`); the
  test mod (task 4).
- Parsing SQL: DuckDB's own settings do that part (issue Approach).

## Definition of done

With the setting set, every DuckDB the backend opens is locked to the folders on every connection
or does not open; unset, nothing changes.
