---
status: pending
complexity: medium
model: sonnet
depends_on: [1, 2, 3]
parallel_group: null
surface: none
kind: null
canonical_reference: backend/duckdb_permitted_test.go
family_pattern_contract: null
flaky: null
isolation: null
---
# Task 4: One test mod and one test that try every route, set and unset

## Objective

Add a test mod under `workspace/test_data/permitted_db_routes/` and one table-driven test,
`TestPermittedDatabaseRoutes`, that drives each route from the issue through the mod parser and
the backend: with the setting set each is refused, with it unset each works. A later reader adds a
route by adding one row. Update the setting's doc comments to say it now covers databases.

## Context

- Issue #833 success criterion 6 (and, end to end, 1-5). Rules: `decision.md` decisions 3, 8, 9.
- What tasks 1-3 built: `backend.ConfineModConnectionString` called from
  `parse.ResolveConnectionString` (task 1); the DuckDB lock in both DuckDB backends (task 2);
  the SQLite ATTACH authorizer (task 3); the cgo-only test drivers in `go.mod` (tasks 2 and 3).
- The routes, one row each:
  1. mod `database` = DuckDB file outside the folders (set: refused at parse; unset: opens and a
     query runs) and inside (works in both);
  2. the same for a SQLite file;
  3. `?` options on a DuckDB value;
  4. `duckdbinit://` from mod content;
  5. each Postgres file option (`sslcert`, `sslkey`, `sslrootcert`, `passfile`, `servicefile`)
     naming an outside file;
  6. MySQL `allowAllFiles=true`;
  7. mod SQL on DuckDB reading, writing (`COPY … TO`), attaching and installing outside;
  8. mod SQL on SQLite attaching an outside file;
  9. a `.ppc` file in the mod's folder defining a DuckDB and a SQLite connection whose
     `file_name` is outside, named by the mod's `database` as `connection.<type>.<name>`
     (decision 3). Load the `.ppc` with the config loader the hosts use when a test can drive it;
     else build the connection as the decoder does and say so in the batch result. Set: the
     connection's `GetConnectionString()` is refused; unset: it returns the outside path unchanged.
  For Postgres and MySQL there is no server in the unit gate, so "works" with the setting unset
  means the parsed value reaches `GetConnectionString()` unchanged; say so in the test's comment.
- Driving a mod: `parse.ParseModDefinition(modFilePath, evalCtx)` (`parse/mod.go:45`) decodes the
  mod block and its `database` through `decodeMod` → `ResolveConnectionString`. It calls
  `modconfig.NewMod`, which needs `modconfig.AppSpecificNewModResourcesFunc`
  (`modconfig/app_specific.go:3`) set; set a minimal one in the test and restore it in
  `t.Cleanup`. `mod.GetDatabase().GetConnectionString()` gives the string;
  `backend.FromConnectionString(ctx, cs)` (`backend/backend.go:52`) then `Connect` opens DuckDB or
  SQLite. Do not call `FromConnectionString` on Postgres or MySQL values (it connects).
- The fixture must not hard-code machine paths: copy it into `t.TempDir()`, make that copy the
  permitted folder, and supply the route's database value (and the outside path for the SQL routes)
  through the eval context (for example `var.database`, `var.outside`) or by templating the copied
  files. The mod's SQL routes live in the fixture as `.sql` files the test reads.
- Test placement: `backend/permitted_routes_test.go`, package `backend_test` (an external test
  package may import `parse` without a cycle), `//go:build cgo`, importing the test drivers as
  task 2's and task 3's tests do. `go test ./...` runs it.
- Docs to update: `app_specific/env.go:70-72` (the comment says "the HCL file functions so far")
  and the package comment in `permitted_paths/permitted_paths.go:1-7`: name the database checks
  and the DuckDB and SQLite lock, and that a host which opens its own DuckDB files (for example
  tailpipe's database and the folders its init script reads) must list their folders.

## Files to modify

- `app_specific/env.go`
- `permitted_paths/permitted_paths.go` (package comment only)

## Files to create

- `workspace/test_data/permitted_db_routes/mod.pp`
- `workspace/test_data/permitted_db_routes/connections.ppc`
- `workspace/test_data/permitted_db_routes/` route `.sql` files (one per SQL route)
- `backend/permitted_routes_test.go`

## Sub-agent prompt

> Read `.projects/833-db-permitted-root-paths/task-4-routes-test-mod.md` (from the repo root) for
> full context, and `.projects/833-db-permitted-root-paths/decision.md`.
> Branch: `issue-833-db-permitted-root-paths`. Working dir: the repo root.
>
> Implement the task:
> 1. Create the test mod `workspace/test_data/permitted_db_routes/`: a `mod.pp` whose mod block's
>    `database` comes from the eval context, a `connections.ppc` for route 9, and one `.sql` file
>    per SQL route (DuckDB read, write,
>    attach, install outside; SQLite attach outside) that names the outside path through a
>    placeholder the test fills. Add a short comment at the top of `mod.pp` saying what the mod is
>    for and that a new route is one row in `TestPermittedDatabaseRoutes`.
> 2. Create `backend/permitted_routes_test.go` with `TestPermittedDatabaseRoutes`: one table, one
>    row per route in the task file's Context list (each Postgres option its own row), each row
>    run as subtest `set` (expect refused; the error names the setting) and `unset` (expect it
>    works as the Context defines). Every row goes through `parse.ParseModDefinition` on a copy of
>    the fixture; DuckDB and SQLite rows then open through `backend.FromConnectionString` and
>    `Connect` and run the row's SQL. Replace task 2's extension seam so the test needs no network.
>    Put a comment above the table: "A new database route is one row here."
> 3. Update the doc comments in `app_specific/env.go` and the package comment of
>    `permitted_paths/permitted_paths.go` as the Context says. Change no code in either file.
> 4. Run the `## Verification` commands.
>
> Set the frontmatter `status: in-progress` when you begin and `status: done` when
> the acceptance criteria are met. Commit on the parent branch with subject
> `833-db-permitted-root-paths: Task 4 — routes-test-mod`, staging the files this task names by
> path, never `.projects/`.
>
> Where `AGENT_FLEET_DECIDE` is `on`, research a choice and decide it (`decide-by-default.md`);
> else on a design question write `BLOCKED-task-4.md` (`## Kind:` first) into the `Project dir:`
> from your dispatch prompt, set `status: blocked`, and exit.

## Acceptance criteria

- [ ] One test mod under `workspace/test_data/permitted_db_routes/` tries every route the issue
      names.
- [ ] `TestPermittedDatabaseRoutes` shows each route refused with the setting set and working with
      it unset, through the mod parser and (for DuckDB and SQLite) the real backend open.
- [ ] The table carries the "a new route is one row here" comment.
- [ ] The setting's doc comments name the database checks and the host's duty to list its own
      DuckDB folders.
- [ ] Tests pass: `go test ./backend/ -run '^TestPermittedDatabaseRoutes$'`.

## Verification

```bash
go vet ./backend/
test -f workspace/test_data/permitted_db_routes/mod.pp
bash -o pipefail -c "go test ./backend/ -run '^TestPermittedDatabaseRoutes$' -count=1 -v | grep -- '--- PASS: TestPermittedDatabaseRoutes '"
bash -o pipefail -c "go test ./backend/ -run '^TestPermittedDatabaseRoutes$' -count=1 -v | grep -c -- '--- PASS: TestPermittedDatabaseRoutes/.*/unset' | awk '\$1 >= 12 {ok=1} END {exit !ok}'"
```

The last command counts the passing `unset` arms (at least 12: the eight routes with the five
Postgres options as separate rows), so a table that lost rows fails; the per-row `set` arms are
proved by the top-level PASS, which needs every subtest to pass.

## Out of scope

- Any change to the checks or the lock themselves (tasks 1-3); a failure found here is fixed in
  the file that owns it and named in the batch result.
- CHANGELOG: release prep writes it.

## Definition of done

One test over one test mod proves every route of the issue refused with the setting set and
working with it unset.
