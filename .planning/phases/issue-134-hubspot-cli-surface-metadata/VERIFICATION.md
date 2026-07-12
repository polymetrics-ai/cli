# Verification: Issue #134 HubSpot CLI Surface Metadata

Date: 2026-07-10

## Pre-production planning gate

Created before production edits:

- `PLAN.md`
- `TDD-LEDGER.md`
- `VERIFICATION.md`
- `RUN-STATE.json`

## Targeted red/green commands

```bash
go test ./cmd/connectorgen -run TestValidate_CLISurfaceImplementedBinaryOperationPasses -count=1
go test ./cmd/connectorgen -run TestValidate_CLISurfaceImplementedBinaryRequiresTypedOperation -count=1
go test ./cmd/connectorgen -run TestValidate_HubSpotCLISurfaceMetadata -count=1
```

## Targeted verification after implementation

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run 'CLISurface|HubSpot'
go test ./internal/connectors/engine -run 'CLISurface|HubSpot'
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Broader verification before handoff

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity evidence

Pending. This slice is metadata-first; runtime dispatcher/help checks are expected to be not applicable unless #135/#136 work is pulled in.

## Results

Pending.
