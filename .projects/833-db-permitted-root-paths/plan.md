# Project: PERMITTED_ROOT_PATHS also covers databases and in-process SQL

**Issue:** #833
**Branch:** `issue-833-db-permitted-root-paths`
**Base branch:** `develop`
**Ship as:** stack
**Topic:** `833-db-permitted-root-paths`
**Repo key:** `turbot/pipe-fittings`
**Project dir:** `.projects/833-db-permitted-root-paths/`
**Contract version:** 2026-09-27
**Scheduling:** edge-gated
**Expected wallclock:** ~30 min x 5 deep = ~150 min

## Goal

When `<APP>_PERMITTED_ROOT_PATHS` (#832) is set, databases and in-process SQL stay inside its
folders. A mod's `database` value that names a DuckDB or SQLite file must point inside the
folders and carry no options; `duckdbinit://` from mod content is refused; Postgres file options
must point inside the folders and MySQL's `allowAllFiles` is refused. Every DuckDB the backend
opens is locked to the folders (`allowed_directories`, `enable_external_access=false`,
`lock_configuration=true`) on every pooled connection, and SQLite SQL cannot ATTACH a file outside
them. With the setting unset, every database behaves as it does today. One table-driven test over
one test mod tries every route, set and unset.

Success is the issue's six success criteria, each met by the task named in its row below, and the
repo's build, unit tests and lint clean at the end.

| Success criterion (issue #833) | Task |
|---|---|
| DuckDB or SQLite file outside the folders fails, inside works | 1 (refusal), 4 (end to end) |
| DuckDB SQL cannot read, write, attach or install outside, on any pooled connection; nothing in the value lifts it | 2, 4 |
| `duckdbinit://`, Postgres file options outside, MySQL `allowAllFiles` refused | 1, 4 |
| SQLite SQL cannot attach outside | 3, 4 |
| Unset: every database behaves as today | 1, 2, 3 (each proves its unset arm), 4 |
| A test mod tries each route; one test, set and unset | 4 |

## Decisions (locked)

See `decision.md`. In short: reuse `permitted_paths.Check` (#832), never a second parser of the
setting; the mod-content checks live in `backend/` and run where mod content becomes a connection
string (`parse.ResolveConnectionString`, `ModVersionConstraint.Initialise`, and each `.ppc`
connection type's `GetConnectionString`, since a mod can ship its own `.ppc`); DuckDB and SQLite are
locked at open through a driver-agnostic per-connection connector, with no driver imported by
non-test code; test-only drivers pinned to the versions powerpipe and tailpipe use; every new path
sits behind `permitted_paths.Enabled()`.

## Status

| Task | Status | Model | Group | Notes |
|------|--------|-------|-------|-------|
| task-1-mod-connection-string-checks | verified | sonnet | - | commit 044107b |
| task-2-duckdb-lockdown | verified | opus | - | commit 1942ca2 |
| task-3-sqlite-attach | verified | sonnet | - | commit 39598bc |
| task-4-routes-test-mod | verified | sonnet | - | commit 5270c74 |
| task-5-final-gate | verified | sonnet | - | gates build, fast, lint exit 0 (evidence/task-5.log) |

## Parallel groups

None. Every task writes into the `backend` package (one package-wide compile and test gate), and
tasks 2 and 3 both write `go.mod`, `go.sum` and the shared connector, so the plan is one chain.

## Tasks

1. **task-1-mod-connection-string-checks.md** — `backend.ConfineModConnectionString`: the refusals
   for a database value from mod content (DuckDB/SQLite file inside, no options; `duckdbinit://`;
   Postgres file options; MySQL `allowAllFiles`), wired into `parse.ResolveConnectionString` and
   `ModVersionConstraint.Initialise`.
2. **task-2-duckdb-lockdown.md** — a driver-agnostic per-connection init connector; every DuckDB
   the duckdb and duckdbinit backends open is locked to the folders on every pooled connection, and
   the open fails if the lock fails. Adds the test-only DuckDB driver.
3. **task-3-sqlite-attach.md** — prove SQLite SQL can ATTACH outside today, then refuse it through
   the same connector with the driver's authorizer. Adds the test-only SQLite driver.
4. **task-4-routes-test-mod.md** — the test mod and the one table-driven test that tries every
   route with the setting set (refused) and unset (works); the setting's docs name databases.
5. **task-5-final-gate.md** — the repo's build, unit tests and lint, once.

## Dependencies

```
1
└── 2                artifact — 2's DuckDB tests open values 1's check produced; compile — both in package backend
    └── 3            artifact — 3 reuses 2's connector; compile — both write go.mod, go.sum, backend/conn_init_connector.go
        └── 4        artifact — 4's routes test drives 1, 2 and 3 end to end (depends on 1, 2, 3)
            └── 5    artifact — final gate over every task (depends on 1, 2, 3, 4)
```

Pairwise write-set check (every pair compared):

- 1 and 2: share the `backend` package compile and test gate → edge (artifact, also covers compile).
- 1 and 3: share the `backend` package gate → covered by 3 → 2 → 1.
- 2 and 3: both write `go.mod`, `go.sum`, `backend/conn_init_connector.go` → edge 3 → 2.
- 1, 2, 3 and 4: 4 writes `backend/permitted_routes_test.go` (package gate) and
  `permitted_paths/permitted_paths.go` (which 2 may also touch) → 4 depends on 1, 2, 3.
- 5 depends on every task.

No `ordering` edges.

## Execution

Sequential. Each task is its own batch and its own stack layer, landing in order.

## Starting states (step 4a)

The project changes what a connection string resolves to (a relative DuckDB/SQLite path becomes
its resolved absolute path) and adds refusals, only when the setting is set. Walked:

- Fresh install, current develop, the previous release (v2.9.5, which has no
  `PERMITTED_ROOT_PATHS` at all): the setting is unset, so every new path is skipped and the
  connection string and open sequence are byte-for-byte today's. Tasks 1, 2 and 3 each prove the
  unset arm with a test.
- A host that sets the setting: refusals and the DuckDB lock apply from the next start. Nothing is
  stored or migrated; nothing is deleted; no manual step. A host that opens its own DuckDB files
  (tailpipe's database, an init script's parquet folders) lists their folders, as the issue says;
  task 4 writes that into the setting's doc comment.

## Insights

(The reduce step appends here after each batch. Empty is normal.)

- task 1: `ConfineModConnectionString` rebuilds only the Postgres query; the rest of the connection string, including the password, is kept as written.
- task 2: once external access is off DuckDB adds its own `.tmp` temp folder and the database file plus its `.wal` to the allow list; the check accepts only their defaults or paths inside the permitted folders, so a DSN `temp_directory` outside them fails the open.
- task 2: `go mod tidy` raised `goccy/go-yaml` v1.11.2 to v1.17.1 (go-duckdb v2.4.3 needs it); later tasks that touch go.mod inherit that bump.
- task 3: the `sqlite3` driver has no `driver.DriverContext`, so the connector's `dsnConnector` fallback covers it; `VACUUM INTO` and `load_extension` outside/with the setting set are refused by SQLite, but no test in `TestSqlitePermitted*` pins them (task 4 should cover them if it owns SQLite routes).
- task 4: `TestPermittedDatabaseRoutes` replaces the DuckDB extension setup through the seam in `backend/export_test.go`; task 5 must keep that seam in the build (not behind a build tag) or the test would try to install extensions over the network.

## Definition of done

- `agent-fleet gates run build`, `fast` and `lint` clean (final-gate task).
- All tasks verified.
- One PR per layer against `develop`, linking #833.
