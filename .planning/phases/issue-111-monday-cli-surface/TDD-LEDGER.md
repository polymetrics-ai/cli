# TDD Ledger — issue #111 Monday CLI surface metadata

## Manual GSD fallback

Programming-loop command is unavailable in the repo-local GSD registry; manual red → green → refactor loop is active and recorded here.

## Red/green ledger

| Step | Evidence | Result |
| --- | --- | --- |
| Red | `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayCLISurface' -count=1` and `go test ./cmd/connectorgen -run 'TestMondayCLISurface' -count=1` fail because Monday has no loaded `cli_surface.json`. | Captured |
| Green | Added `internal/connectors/defs/monday/cli_surface.json`; `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayCLISurface' -count=1` and `go test ./cmd/connectorgen -run 'TestMondayCLISurface' -count=1` pass. | Captured |
| Refactor | Full defs validation via `go run ./cmd/connectorgen validate internal/connectors/defs --json` passes with 547 connectors, 0 findings, 0 warnings. | Captured |

## Expected failing tests before green

```bash
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayCLISurface' -count=1
go test ./cmd/connectorgen -run 'TestMondayCLISurface' -count=1
```
