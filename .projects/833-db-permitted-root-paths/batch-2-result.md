---
# Batch 2 result — 2026-11-23 19:48

Tasks: 1 (parallel_group: null; task 3 waits on the compile edge from task 2)
Started: 2026-11-23T19:41:28Z
Finished: 2026-11-23T19:48:12Z
Wallclock: 7 (orchestrator-side)

---

Task 2: duckdb-lockdown
Status: verified
Files touched: backend/conn_init_connector.go, backend/duckdb_permitted.go, backend/duckdb_permitted_test.go, backend/duckdb_backend.go, backend/duckdb_init_backend.go, go.mod, go.sum, permitted_paths/permitted_paths.go
Tests run: go build ./backend/... ./permitted_paths/..., go vet ./backend/ ./permitted_paths/, go test ./backend/ -run '^TestDuckDBPermitted' (Lockdown and Unset PASS), go list -deps driver-free check; five commands, each exit 0
Verification output: |
  5 `exit=` lines in evidence/task-2.log, all exit=0 (accepted by the step-9 evidence gate)
Evidence log: .projects/833-db-permitted-root-paths/evidence/task-2.log
Commit: 1942ca2d235b1d3bde26322990065f169c0c1c52 "833-db-permitted-root-paths: Task 2 — duckdb-lockdown"
Notes: The init also empties `allowed_paths`. Symlinks inside a permitted folder that point outside it are still readable from SQL (DuckDB checks paths as text): not fixed, outside this task's scope.

---

Insights promoted:
- task 2: once external access is off DuckDB adds its own `.tmp` temp folder and the database file plus its `.wal` to the allow list; the check accepts only their defaults or paths inside the permitted folders, so a DSN `temp_directory` outside them fails the open.
- task 2: `go mod tidy` raised `goccy/go-yaml` v1.11.2 to v1.17.1 (go-duckdb v2.4.3 needs it); later tasks that touch go.mod inherit that bump.
