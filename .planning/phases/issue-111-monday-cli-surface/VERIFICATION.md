# Verification — issue #111 Monday CLI surface metadata

## Targeted commands

```bash
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayCLISurface' -count=1
go test ./cmd/connectorgen -run 'TestMondayCLISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/monday --json
```

## Broader follow-up

```bash
go test ./internal/connectors/commandrunner ./internal/cli -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

## Results

- `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayCLISurface' -count=1` — pass.
- `go test ./cmd/connectorgen -run 'TestMondayCLISurface' -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs/monday --json` — known command-shape caveat: this validate command treats child `schemas/` and `fixtures/` as connector dirs when pointed at a single bundle path; use parent-root validation instead.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass: 547 connectors, 0 findings, 0 warnings; no Monday findings.
