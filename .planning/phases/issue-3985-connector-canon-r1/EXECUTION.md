# Execution — Issue 3985 connector canon

## Delivery mode

`scripts/gsd prompt execute-phase issue-3985-connector-canon-r1` resolved successfully. The
canonical single-worker contract forbids spawning GSD roles, so execution used the documented inline
fallback. The task also required autonomous completion and prohibited connector implementation work.

## Completed slices

1. Imported source reports/captain rulings with SHA-256 pins; built the canon index, procedure,
   remote-reproducibility report, and recoverable archive entry points.
2. Added the real commandrunner preflight sweep as `make connector-runtime-preflight`; `test` covers
   the same sweep once in each `make verify` and `make verify-parallel` run. Added
   `make connector-canon-check` to check the source pins, archive location, current procedure,
   corrections, and public status wording.
3. Reconciled the branch onto `origin/main` at `4df0b0416`, preserving its source-pinned GitHub
   parity artifacts and distinguishing those current generated counts from the void wrong-branch
   historical gap map.
4. Per captain addendum, audited #3972 and its eleven children through GitHub REST, created/attached
   #3987, revised #3972/#3978, and added the preserved-scope amendments recorded in
   `POSTGRES-PARITY-AUDIT.md`. No PostgreSQL or GitHub implementation branch was touched.
5. Replaced the now-stale r1 PostgreSQL execution tree with the r2 current report, preserving r1
   unchanged in the data archive with an explicit supersession marker.

## TDD evidence

- **Red:** `make connector-runtime-preflight` and `make connector-canon-check` initially failed
  because neither target existed.
- **Green:** both targets pass; aggregate verification reaches the real
  `TestEveryImplementedCommandPassesRuntimePreflight` sweep once through `test`, rather than a
  copied validator.
- **Captain addendum red:** the live r1 parent had 11 children and no owner for a warehouse-only
  four-flow/seven-mode contract.
- **Green:** #3987 exists as child 12, #3978 depends on it, and the pinned r2 report plus canon check
  preserve the resulting graph.

## External-state evidence

All GitHub work used `gh-axi` REST operations. Read-back reported `12 of 12` sub-issues under #3972.
#3974 was assessed correct and was not edited while Wave A remained active.
