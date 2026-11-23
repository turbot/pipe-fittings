Status: held at self-review (fixing in this run)

## Blocking

1. backend/sqlite_permitted.go:80 — SQLite ATTACH by expression or bound parameter: the authorizer gets no file name, the empty name is allowed, and `ATTACH '/outside/' || 'x.db'` or `ATTACH ? AS p` creates a file outside the folders. Criterion 4. Probe: TestProbeSqliteAttachExpr.
2. backend/permitted_paths.go:63 — a value starting with `:memory:` skips the path check, but SQLite reads `:memory:/../../outside/x.sqlite` as a file path. Criterion 1. Probe: TestProbeSqliteMemoryPrefixPath.
3. backend/permitted_paths.go:98 — Postgres `servicefile` indirection: a service file inside the folders names an `sslrootcert`, `sslkey` or `passfile` outside them, and the checked string passes. Criterion 3. Probe: TestProbePostgresServiceFileIndirection.
4. connection/steampipe_pg.go:99 — the steampipe connection type's GetConnectionString skips backend.ConfineModConnectionString, so `sslkey`, `sslcert` and `passfile` outside the folders reach pgx. Criterion 3 and decision 3; a copy of the postgres type (kind 8). Probe: TestProbeSteampipeConnectionStringUnconfined.
5. backend/duckdb_permitted.go:100 — DuckDB matches allowed_directories as text, so a symlink inside the folders reads and writes outside files. Criterion 2. Batch 2 recorded it as open, but the body's Still open says None. Probe: TestProbeDuckDBSymlinkEscape.

## Tidy

1. PR body § Proof — cites the gitignored evidence/task-5.log by hand; render it from the record.
2. permitted_paths/permitted_paths.go:13 — the package doc must stay true once items 1 to 4 are fixed.

## Probe patches

Kept by the reviewer under $TMPDIR (self-review-probe.lv4E3Y-probe-{1..5}.patch), not in the repo.
