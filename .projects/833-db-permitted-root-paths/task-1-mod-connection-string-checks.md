---
status: pending
complexity: medium
model: sonnet
depends_on: []
parallel_group: null
surface: none
kind: null
canonical_reference: funcs/permitted_paths.go
family_pattern_contract: null
flaky: null
isolation: null
---
# Task 1: Refuse database values from mod content that leave the permitted folders

## Objective

Add `backend.ConfineModConnectionString(cs string) (string, error)` and route every database
value from mod content through it, so with `<APP>_PERMITTED_ROOT_PATHS` set a DuckDB or SQLite
file must be inside the folders with no options, `duckdbinit://` is refused, Postgres file options
must be inside the folders and MySQL `allowAllFiles` is refused. Unset: the value passes through
unchanged.

## Context

- Issue #833 bullets 1-3 and success criteria 1, 3 and 5. Rules: `decision.md` decisions 1-5, 9, 10.
- #832's checker: `permitted_paths/permitted_paths.go` (`Enabled` :64, `Setting` :70, `Check`
  :131, `OutsideError` :108). `Check` returns the absolute, symlink-resolved path to open.
- Pattern to mirror: `funcs/permitted_paths.go` (`confinePath` :128-140 checks and returns the
  resolved path; :159-167 shapes the error with `errors.As(err, &outside)`), and its test helper
  `setPermitted` / `permit` in `funcs/permitted_paths_test.go:80-104` (sets the app prefix with
  `app_specific.SetAppSpecificEnvVarKeys`, `t.Setenv`, `permitted_paths.Load()`, resets in
  `t.Cleanup`).
- Where mod content becomes a connection string:
  - `parse/decode.go:72` `ResolveConnectionString`; the plain-string branch is :106-107
    (`csp = connection.NewConnectionString(dbValue.AsString())`). Hosts call it for every
    resource's `database`; `decodeMod` (:53) calls it for the mod's.
  - `modconfig/mod_version_constraint.go:27` `Database *string` (a `require { mod … { database } }`
    override), initialised in `Initialise(block)` at :98.
  - The `connection.<type>.<name>` branch (:92-100) and the connection-object branch (:108-140) are
    the host's own and stay unchecked (decision 3).
- Prefixes: `backend/duckdb_backend.go:15`, `backend/sqlite_backend.go:14`,
  `backend/duckdb_init_backend.go:16`, `backend/postgres_backend.go:23`,
  `backend/mysql_backend.go:17`. Classification today is case-sensitive `strings.HasPrefix` on the
  untrimmed string (`backend/backend.go:121-152`); the constructors `TrimSpace` first. Trim before
  classifying in the new function so `" duckdb:///x"` cannot slip past.
- Import graph: `backend` imports neither `parse`, `modconfig` nor `connection`, and must not
  (`connection` imports `backend`). `backend` may import `permitted_paths`.

## Files to modify

- `parse/decode.go`
- `modconfig/mod_version_constraint.go`

## Files to create

- `backend/permitted_paths.go`
- `backend/permitted_paths_test.go`
- `parse/resolve_connection_string_permitted_test.go`
- `modconfig/mod_version_constraint_permitted_test.go`

## Sub-agent prompt

> Read `.projects/833-db-permitted-root-paths/task-1-mod-connection-string-checks.md` (from the
> repo root) for full context, and `.projects/833-db-permitted-root-paths/decision.md` decisions
> 1-5, 9 and 10, which are locked.
> Branch: `issue-833-db-permitted-root-paths`. Working dir: the repo root.
>
> Implement the task:
> 1. Create `backend/permitted_paths.go` with
>    `func ConfineModConnectionString(cs string) (string, error)`. With
>    `permitted_paths.Enabled()` false, return `cs` unchanged and nil. Otherwise trim, classify by
>    the backend prefixes, and apply decision 4 (DuckDB and SQLite: `?` refused; empty, `:memory:`
>    and `:memory:<name>` allowed; a `<scheme>:` form refused; anything else made absolute against
>    the working directory, passed to `permitted_paths.Check`, and the result rewritten to
>    `<prefix><resolved>`), refuse `duckdbinit://` outright, and apply decision 5 (Postgres via
>    `net/url`: `sslcert`, `sslkey`, `sslrootcert`, `passfile`, `servicefile`, keys
>    case-insensitive, every value, each checked and rewritten to its resolved path; an unparseable
>    URL refused; MySQL: an `allowAllFiles` key, case-insensitive, refused whatever its value). A
>    value with no known prefix passes through unchanged (the backend rejects it later as today).
>    Errors name the setting (`permitted_paths.Setting()`) and the option or path, wrap
>    `*permitted_paths.OutsideError` where `Check` returned one, and never include a whole Postgres
>    or MySQL string (it can carry a password). Write a doc comment saying it is for values from
>    mod content and why the host's own strings are not passed through it.
> 2. Create `backend/permitted_paths_test.go` with `TestConfineModConnectionString`: a table over
>    `t.TempDir()` folders (one permitted, one outside, and a symlink inside the permitted folder
>    that points outside) with one row per rule: DuckDB inside (rewritten to the resolved path),
>    outside, through the outward symlink, relative path, `?` options, in-memory forms, `md:` and
>    `file:` forms; SQLite the same; `duckdbinit://` refused; each of the five Postgres options
>    inside (rewritten) and outside, upper-case key, repeated key with one outside value; MySQL
>    `allowAllFiles=true`, `=false` and `ALLOWALLFILES=1` refused, a MySQL string without it passes;
>    a password-bearing Postgres string whose refusal message does not contain the password; a
>    relative entry in the setting fails closed. Run every row a second time with the setting unset
>    and assert the output equals the input and the error is nil. Write a test-local helper like
>    `funcs/permitted_paths_test.go`'s `setPermitted`.
> 3. In `parse/decode.go` `ResolveConnectionString`, pass the plain-string value through
>    `backend.ConfineModConnectionString` before `connection.NewConnectionString`, and on error
>    return an `hcl.DiagError` diagnostic whose `Subject` is `attr.Range.Ptr()`. Leave the
>    connection-reference and connection-object branches as they are. Add
>    `parse/resolve_connection_string_permitted_test.go` with
>    `TestResolveConnectionStringPermitted`: parse a small HCL body with `hclsyntax`, take
>    `PartialContent` for `database` as `decodeMod` does, and assert an outside DuckDB file gives an
>    error diagnostic with that subject, an inside one yields a provider whose
>    `GetConnectionString()` is the resolved path, and with the setting unset the outside value
>    comes back unchanged.
> 4. In `modconfig/mod_version_constraint.go` `Initialise`, when `m.Database` is set, pass it
>    through `backend.ConfineModConnectionString`; on error return a diagnostic on the block's
>    range, on success store the rewritten value. Add
>    `modconfig/mod_version_constraint_permitted_test.go` with
>    `TestModVersionConstraintDatabasePermitted` (outside refused, inside rewritten, unset unchanged).
>    If `Initialise` cannot be driven from a test without an app registration, test a small
>    unexported helper it calls instead and say so in the batch result.
> 5. Run the `## Verification` commands.
>
> Set the frontmatter `status: in-progress` when you begin and `status: done` when
> the acceptance criteria are met. Commit on the parent branch with subject
> `833-db-permitted-root-paths: Task 1 — mod-connection-string-checks`, staging the files this
> task names by path, never `.projects/`.
>
> Where `AGENT_FLEET_DECIDE` is `on`, research a choice and decide it (`decide-by-default.md`);
> else on a design question write `BLOCKED-task-1.md` (`## Kind:` first) into the `Project dir:`
> from your dispatch prompt, set `status: blocked`, and exit.

## Acceptance criteria

- [ ] With the setting set, a mod `database` naming a DuckDB or SQLite file outside the folders is
      an HCL error diagnostic on the attribute; one inside resolves to the resolved absolute path.
- [ ] `?` options on a DuckDB/SQLite value, `duckdbinit://`, a Postgres file option outside the
      folders and MySQL `allowAllFiles` are refused from mod content.
- [ ] A `require` block's `database` override gets the same checks.
- [ ] With the setting unset, every value comes back byte-for-byte unchanged.
- [ ] No refusal message contains a Postgres or MySQL password.
- [ ] Tests pass: the three named tests below.

## Verification

```bash
go build ./backend/... ./parse/... ./modconfig/...
go vet ./backend/ ./parse/ ./modconfig/
bash -o pipefail -c "go test ./backend/ -run '^TestConfineModConnectionString$' -count=1 -v | grep -- '--- PASS: TestConfineModConnectionString'"
bash -o pipefail -c "go test ./parse/ -run '^TestResolveConnectionStringPermitted$' -count=1 -v | grep -- '--- PASS: TestResolveConnectionStringPermitted'"
bash -o pipefail -c "go test ./modconfig/ -run 'Permitted' -count=1 -v | grep -- '--- PASS: '"
```

The checks cannot see whether a host calls `ResolveConnectionString` for its resources; that is
the host's existing behaviour and task 4 drives a mod through the parser end to end.

## Out of scope

- Locking DuckDB SQL at open (task 2); SQLite ATTACH (task 3); the test mod (task 4).
- Checking connection objects, `connection.<type>.<name>` references or CLI values (decision 3).

## Definition of done

Every database value from mod content passes through one checker built on `permitted_paths.Check`,
refused or rewritten when the setting is set and untouched when it is not.
