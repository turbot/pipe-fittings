Status: held at self-review (round 6 of the self-review; the cap of six passes was reached with one blocking finding open)

Branch HEAD when held: 41f5d09. Base: origin/develop. No PR is open.

Rounds 1 to 5 fixed the earlier findings (commits cd93b1f, 9c16364, d5a1e86, 8138af2, 41f5d09). Each round's tests failed before its fix. Gates build, fast and lint exit 0 at 41f5d09.

## Blocking

1. backend/permitted_paths.go ~122 — `sslrootcert=system` (pgx's value for the system CA pool, which reads no file) is refused when the working folder is outside the folders, and rewritten to `<cwd>/system` (a missing file) when it is inside, so the connection fails. Criterion 3 only refuses file options outside the folders. Fix: leave `sslrootcert=system` unchanged and unchecked, with a test row. Probe: TestProbeSslRootCertSystem (review patch self-review-probe.ZUh6Aw-probe-1.patch under $TMPDIR).

## Tidy

1. PR body — text before Proof is over 150 words (193 at the last draft); trim it.
2. PR body — Proof must be rendered from the gates record, not typed by hand.
3. PR body — say a plain SQLite VACUUM is refused, and that the temp folder is per database under .duckdb_tmp.
4. parse/workspace_profile_permitted_test.go ~73 — rename the "inherited from a base" row: the base's own decode refuses it.
5. app_specific/env.go ~70 — the setting's list of what it covers should name the workspace profile's mod_location and snapshot_location.

## Notes for the person who picks this up

- The reviews kept finding new routes on the DuckDB and profile surfaces: rounds 2 to 6 each added a finding not in the round before. Whether this project should keep going round by round, or take a narrower scope, is a call for a person.
- The fourth round's probe once wrote a file to the real home folder (~/.duckdb/stored_secrets). It was removed; the tests were changed to use a temp HOME.
- A dependency side-effect was reported (goccy/go-yaml 1.11.2 to 1.17.1, through arrow-go). Confirm it is intended before the PR.
- Empty .duckdb_tmp folders are left behind after use (per database, inside the first writable folder).

## Round 7 (self-review over the whole branch, at e6483b1)

Round 6's blocking item and tidy items 4 and 5 are fixed at e6483b1. Round 7 found one blocking item
that needs a person's decision, so the ship is held again (see `BLOCKED-ship.md`).

## Blocking (round 7)

1. backend/duckdb_init_backend.go:66 — with the setting set, a host's DuckDB init script that runs `install` or `load` (tailpipe's connect script does) fails with a permission error, because the lock is applied before the script runs; listing the folders does not help. The issue's promise that listing a host's folders keeps its DuckDB files working is therefore false for this case. Needs a product decision: see BLOCKED-ship.md.

## Needs a person

- The item above (BLOCKED-ship.md).
