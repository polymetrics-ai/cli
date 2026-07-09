# Verification — issue #82 Monday CLI parity parent

## Required baseline

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI/help/docs parity checks

```bash
pm help monday
pm monday
pm monday --help
pm monday board list --help
pm monday item list --help
rg -n "monday|Monday" docs/cli website internal/connectors/defs/monday
```

Record not-applicable cases explicitly when a command/help surface is generated from connector metadata rather than a static docs page.

## Current results

- `scripts/gsd doctor`: pass.
- `scripts/gsd verify-pi`: pass.
- `scripts/gsd list --json`: pass (69 commands).
- `scripts/gsd prompt programming-loop ...`: unavailable; manual fallback recorded.
- #111 targeted tests: pass (`go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayCLISurface' -count=1`; `go test ./cmd/connectorgen -run 'TestMondayCLISurface' -count=1`).
- Connector validation: pass (`go run ./cmd/connectorgen validate internal/connectors/defs --json` → 547 connectors, 0 findings, 0 warnings).
- Full Go/parent gates: pending.
