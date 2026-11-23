---
status: verified
complexity: medium
model: sonnet
depends_on: [2, "1:compile"]
parallel_group: null
surface: none
kind: null
canonical_reference: backend/conn_init_connector.go
family_pattern_contract: null
flaky: null
isolation: null
---
# Task 3: SQLite SQL cannot ATTACH a file outside the permitted folders

## Objective

Prove with a test that SQLite SQL can ATTACH (and so create or read) a file outside the folders
today, then, with the setting set, refuse that on every connection `SqliteBackend.Connect` opens.
Unset: today's open, unchanged.

## Context

- Issue #833 bullet 5 and success criteria 4 and 5. Rules: `decision.md` decisions 1, 7, 8, 9.
- Task 2 created `backend/conn_init_connector.go` (a driver-agnostic connector running an init on
  every new `driver.Conn`); reuse it, do not write a second one. If it needs a small change for a
  driver without `driver.DriverContext`, make it there.
- Today's open: `backend/sqlite_backend.go:33-43` `Connect`: `sql.Open("sqlite3", …)` and pool
  settings, no init SQL. The `"sqlite3"` name is `mattn/go-sqlite3`, which hosts use at v1.14.52.
- `mattn/go-sqlite3` v1.14.52: `func (c *SQLiteConn) RegisterAuthorizer(callback func(int, string, string, string) int)`
  (`sqlite3.go:730`), `SQLITE_ATTACH` (`sqlite3.go:403`, value 24 in SQLite), `SQLITE_DENY` = 1,
  `SQLITE_OK` = 0. For `SQLITE_ATTACH` the first string argument is the file name as written. In a
  non-cgo build (`static_mock.go:32`) `RegisterAuthorizer` is a no-op, but such a build cannot open
  SQLite at all.
- Test helper pattern: `funcs/permitted_paths_test.go:80-104` (`setPermitted`); task 2's
  `backend/duckdb_permitted_test.go` for the backend-level test shape.

## Files to modify

- `backend/sqlite_backend.go`
- `backend/conn_init_connector.go` (only if a non-`DriverContext` driver needs it)
- `go.mod`
- `go.sum`

## Files to create

- `backend/sqlite_permitted.go`
- `backend/sqlite_permitted_test.go`

## Sub-agent prompt

> Read `.projects/833-db-permitted-root-paths/task-3-sqlite-attach.md` (from the repo root) for
> full context, and `.projects/833-db-permitted-root-paths/decision.md` decisions 1, 7, 8 and 9,
> which are locked.
> Branch: `issue-833-db-permitted-root-paths`. Working dir: the repo root.
>
> Implement the task:
> 1. Add `github.com/mattn/go-sqlite3 v1.14.52` to `go.mod` (exactly that version) and
>    `go mod tidy`. Import it only from `backend/sqlite_permitted_test.go`, which carries
>    `//go:build cgo`.
> 2. Write `TestSqlitePermittedUnset` first and run it before any code change: with the setting
>    unset, `SqliteBackend.Connect` on a database inside a temp folder, then
>    `ATTACH '<outside>/x.db' AS o` and `CREATE TABLE o.t(a)` both succeed and the outside file
>    exists. Record in the batch result that SQLite could attach outside before the change.
> 3. Create `backend/sqlite_permitted.go`: the SQLite init for task 2's connector. It asserts the
>    raw connection to a locally declared
>    `interface{ RegisterAuthorizer(func(int, string, string, string) int) }` (fail the connection
>    when it does not have it) and registers an authorizer that, for action 24 (`SQLITE_ATTACH`),
>    denies a file name that starts with `file:` or fails `permitted_paths.Check`, allows an empty
>    name and `:memory:`, and allows every other action. Declare the action and result codes as
>    named constants with a comment citing SQLite's values.
> 4. In `SqliteBackend.Connect`, when `permitted_paths.Enabled()`, open through the connector with
>    that init, keep the pool settings, force one connection (`PingContext`) so a failure fails the
>    open, and on failure close the pool and return the error. Unset: today's code path unchanged.
> 5. Add `TestSqlitePermittedAttach` to the test file: setting = one temp folder,
>    `MaxOpenConns` 3, three connections held at once; on each, `ATTACH` of an outside file fails
>    and leaves no outside file, `ATTACH` of `file:<outside>?mode=rwc` fails, `ATTACH` through a
>    symlink inside that points outside fails, `ATTACH` of a file inside the folder works, and
>    `ATTACH ':memory:'` works. Add `TestSqlitePermittedBadSetting`: a relative entry makes
>    `Connect` fail. Add `TestSqlitePermittedNoAuthorizer`: the SQLite init called on a fake
>    `driver.Conn` that lacks `RegisterAuthorizer` returns an error.
> 6. Run the `## Verification` commands.
>
> Set the frontmatter `status: in-progress` when you begin and `status: done` when
> the acceptance criteria are met. Commit on the parent branch with subject
> `833-db-permitted-root-paths: Task 3 — sqlite-attach`, staging the files this task names by path,
> never `.projects/`.
>
> Where `AGENT_FLEET_DECIDE` is `on`, research a choice and decide it (`decide-by-default.md`);
> else on a design question write `BLOCKED-task-3.md` (`## Kind:` first) into the `Project dir:`
> from your dispatch prompt, set `status: blocked`, and exit.

## Acceptance criteria

- [ ] A test shows SQLite SQL can ATTACH outside the folders with the setting unset.
- [ ] With the setting set, ATTACH of a file outside the folders (plain path, `file:` URI, outward
      symlink) fails on every pooled connection; ATTACH inside and of `:memory:` works.
- [ ] A driver without an authorizer, or a bad setting, fails the open.
- [ ] No non-test package imports a SQLite driver.
- [ ] Tests pass: `go test ./backend/ -run '^TestSqlitePermitted'`.

## Verification

```bash
go build ./backend/...
go vet ./backend/
bash -o pipefail -c "go test ./backend/ -run '^TestSqlitePermitted' -count=1 -v | grep -- '--- PASS: TestSqlitePermittedAttach'"
bash -o pipefail -c "go test ./backend/ -run '^TestSqlitePermitted' -count=1 -v | grep -- '--- PASS: TestSqlitePermittedUnset'"
go list -deps ./backend/... ./permitted_paths/... ./parse/... ./modconfig/... | grep -E 'go-duckdb|go-sqlite3' && exit 1 || true
```

The ATTACH test cannot see SQLite's other ways to write a file (for example `VACUUM INTO`); the
issue asks for ATTACH, and the batch result names any other route the executor finds open, for
the reviewer.

## Out of scope

- The SQLite main file from mod content (task 1 checks it); DuckDB (task 2); the test mod (task 4).

## Definition of done

With the setting set, SQLite SQL on a backend-opened database cannot attach a file outside the
folders, on any connection; unset, nothing changes.
