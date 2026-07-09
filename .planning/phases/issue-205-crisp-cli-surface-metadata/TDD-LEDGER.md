# TDD Ledger — Issue #205 Crisp CLI surface metadata

## Manual GSD/TDD fallback

`gsd-programming-loop` is not registered in the current `scripts/gsd` command registry. This issue follows the manual universal programming loop: plan, capture red validation, implement smallest metadata scaffold, run targeted validation, then update verification evidence.

## Ledger

| Step | Command | Expected | Actual | Status |
|---|---|---|---|---|
| Red validation | `go run ./cmd/connectorgen validate internal/connectors/defs/crisp` | fails because bundle absent | pending | planned |
| Implement metadata scaffold | write Crisp `metadata.json`, `spec.json`, empty `streams.json`, `api_surface.json`, `cli_surface.json`, `docs.md` | files added | pending | planned |
| Targeted green | `go run ./cmd/connectorgen validate internal/connectors/defs/crisp` | pass | pending | planned |
| Fleet green | `go run ./cmd/connectorgen validate internal/connectors/defs` | pass | pending | planned |

## Behavior-change note

#205 is metadata/validation only. It must not add executable streams, direct-read dispatch, write actions, binary downloads, or reverse-ETL execution. Later issues own behavior tests.
