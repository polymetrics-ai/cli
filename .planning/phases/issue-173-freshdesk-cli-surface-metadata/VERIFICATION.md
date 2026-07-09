# Verification: Freshdesk CLI Surface Metadata

Date: 2026-07-09

## Current Verification

Red baseline only:

- API surface count check failed as expected: current Freshdesk surface has 10 entries, target baseline is 170.
- CLI surface presence check failed as expected: `cli_surface.json` is absent.

## Required Green Commands

```bash
python3 <json/count validation>
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen -run CLISurface
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Broader Handoff Commands

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Do not run credentialed Freshdesk checks unless explicitly requested.
