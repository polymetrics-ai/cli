# Verification: Freshdesk Full Operation Implementation

Date: 2026-07-10

## Current Red Baseline

- Freshdesk implemented coverage check fails with 5 covered rows and 165 blocked operation rows.

## Required Green Commands

```bash
go test ./internal/connectors/commandrunner -run DirectRead
go test ./internal/connectors/engine -run DirectRead
go test ./cmd/connectorgen -run CLISurface
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner
go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

## Broader Handoff Commands

```bash
gofmt -w cmd internal
go vet ./...
go test ./... -timeout 20m
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Credentialed Freshdesk checks are not allowed unless explicitly requested.
