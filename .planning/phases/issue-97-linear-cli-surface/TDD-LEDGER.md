# TDD Ledger — Issue #97 Linear CLI surface metadata

Date: 2026-07-09

## GSD / fallback record

- GSD programming loop attempted: `scripts/gsd prompt programming-loop init --phase issue-97-linear-cli-surface --dry-run`.
- Result: unavailable (`unknown GSD command: programming-loop`).
- Manual fallback: PLAN → RED → GREEN → REFACTOR → VERIFY.

## Red / green / refactor log

| Step | Command / artifact | Expected | Actual |
|---|---|---|---|
| RED | Add `TestLinearCLISurfaceMapsImplementedStreams`; run `go test ./internal/connectors/engine -run TestLinearCLISurfaceMapsImplementedStreams -count=1` | Fail before `cli_surface.json` exists | Failed as expected: `linear cli surface missing` |
| GREEN | Add `internal/connectors/defs/linear/cli_surface.json`; rerun focused test | Pass | Passed: `go test ./internal/connectors/engine -run TestLinearCLISurfaceMapsImplementedStreams -count=1` |
| VALIDATE | `go run ./cmd/connectorgen validate internal/connectors/defs/linear --json` | 0 findings | The single-connector path is not supported by this validator shape (it treated `fixtures`/`schemas` as bundles). Full defs validation passed: `go run ./cmd/connectorgen validate internal/connectors/defs --json` with 0 findings / 547 checked. |
| REGRESSION | `go test ./internal/connectors/engine ./internal/connectors/commandrunner -count=1` | Pass | Passed. |

## Safety assertions to preserve

- Implemented commands only map to existing streams.
- No implemented direct-read command until Linear-specific output policy/executor work is approved.
- No raw API/raw GraphQL command.
- No reverse ETL write marked implemented.
- No secrets or secret-shaped fixtures/docs.
