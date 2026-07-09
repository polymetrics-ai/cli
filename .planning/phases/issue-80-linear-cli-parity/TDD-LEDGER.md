# TDD Ledger — Issue #80 Linear CLI parity parent

Date: 2026-07-09

## GSD / fallback record

- Parent planning prompt: `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-80-linear-cli-parity --dry-run`.
- Result: unavailable (`unknown GSD command: programming-loop`), so manual-GSD fallback is recorded and active.

## Parent TDD policy

Every behavior-changing sub-issue must create or update its own TDD ledger before production edits. The parent ledger aggregates evidence and does not replace sub-issue red/green/refactor records.

## Sub-issue evidence index

| Issue | Lane | Red evidence | Green evidence | Refactor/verification |
|---:|---|---|---|---|
| #97 | CLI surface metadata | `go test ./internal/connectors/engine -run TestLinearCLISurfaceMapsImplementedStreams -count=1` failed as expected: `linear cli surface missing` | Added `internal/connectors/defs/linear/cli_surface.json`; focused test passed | `make verify` passed; website data regenerated; website unit test blocked by missing local deps |
| #98 | Help renderer/docs | Pending | Pending | Pending |
| #99 | Stream runner | Pending | Pending | Pending |
| #100 | Operation ledger | Pending | Pending | Pending |
| #101 | Direct reads | Pending | Pending | Pending |
| #102 | GraphQL/advanced engine | Pending | Pending | Pending |
| #103 | Sensitive/admin policy | Pending | Pending | Pending |

## Safety notes

No credentialed Linear checks, no secrets, no new dependencies, no raw generic GraphQL write/read tools, and no reverse ETL execution are allowed in this parent loop.
