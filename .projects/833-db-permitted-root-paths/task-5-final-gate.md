---
status: pending
complexity: medium
model: sonnet
depends_on: [1, 2, 3, 4]
parallel_group: null
surface: none
kind: final-gate
canonical_reference: null
family_pattern_contract: null
flaky: null
isolation: null
---
# Task 5: Final gate

## Objective

Run the repo's full boundary checks once over the whole project and fix what they surface.

## Context

- `gates.yml` (find it with `node "${CLAUDE_CONFIG_DIR:-$HOME/.claude}/scripts/fleet-files.cjs" find gates .`)
  has no top-level `final:`, so the gate is one `agent-fleet gates run <check>` per check the
  touched kinds owe: the `go` kind (`**/*.go`, `go.mod`, `go.sum`) owes `build`, `fast` and
  `lint`; the `testdata` kind (`workspace/test_data/**`) owes `fast`.
- The new DuckDB and SQLite tests need cgo; `fast` runs `go test ./... -timeout 120s`, and the
  first link of the DuckDB library is slow but outside the test timeout.
- `never_run` in `gates.yml` bans `make test` and `go clean*`; do not run them.

## Files to modify

- Whatever the gate surfaces, within this project's files.

## Files to create

- None.

## Sub-agent prompt

> Read `.projects/833-db-permitted-root-paths/task-5-final-gate.md` (from the repo root) for
> full context.
> Branch: `issue-833-db-permitted-root-paths`. Working dir: the repo root.
>
> From the repo root, in order, fixing anything that surfaces:
> 1. The repo's formatter or autofix, if it has one: `gofmt -l .` on the files this project
>    changed, and `gofmt -w` on any it lists.
> 2. The repo's gate: `agent-fleet gates run build`, then `agent-fleet gates run fast`, then
>    `agent-fleet gates run lint`. Fix every failure the gate surfaces and nothing wider. A failure
>    already on the project's base branch follows `reference/inherited-failures.md`: one in a test
>    covering a file the project changed is yours to fix; any other is noted in the batch result,
>    not fixed.
> 3. Commit. For a failure you left, say in the batch result how you know it is on the base (never
>    `git stash` or `reset` to test).
>
> Set the frontmatter `status: in-progress` when you begin and `status: done` when
> the gate passes. Commit on the parent branch with subject
> `833-db-permitted-root-paths: Task 5 — final-gate`, staging the files you changed by path, never
> `.projects/`.
>
> Where `AGENT_FLEET_DECIDE` is `on`, research a choice and decide it (`decide-by-default.md`);
> else on a design question write `BLOCKED-task-5.md` (`## Kind:` first) into the `Project dir:`
> from your dispatch prompt, set `status: blocked`, and exit.

## Acceptance criteria

- [ ] `agent-fleet gates run build`, `fast` and `lint` all exit 0.

## Verification

```bash
agent-fleet gates run build
agent-fleet gates run fast
agent-fleet gates run lint
```

## Out of scope

- New behaviour or tests beyond what the gate asks for.

## Definition of done

The repo's build, unit tests and lint pass on the project branch.
