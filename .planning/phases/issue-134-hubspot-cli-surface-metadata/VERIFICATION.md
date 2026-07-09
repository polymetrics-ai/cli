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

Passed:

```bash
gofmt -w cmd/connectorgen/main_test.go cmd/connectorgen/validate.go
go test ./cmd/connectorgen -run 'TestValidate_CLISurfaceImplementedBinaryOperationPasses|TestValidate_CLISurfaceImplementedBinaryRequiresTypedOperation|TestValidate_HubSpotCLISurfaceMetadata' -count=1
go test ./cmd/connectorgen -run 'CLISurface|HubSpot' -count=1
go test ./internal/connectors/engine -run 'CLISurface|HubSpot' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen ./internal/connectors/engine
go run ./cmd/pm help docs
go run ./cmd/pm docs validate --connectors-dir docs/connectors
go run ./cmd/pm help connectors
go run ./cmd/pm connectors inspect hubspot --json
```

`connectorgen validate` result: 548 connector(s) checked, 0 findings.
Full connectorgen+engine package tests passed.
Docs validation passed after updating the HubSpot manual/skill and catalog entries.

Broad gates passed after the targeted metadata slice:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Notes:
- First broad `go test ./...` exposed hard-coded connector counts (`551`/`547`) and a cold-load timeout in `internal/connectors/certify` after adding the 548th bundle. Fixed by updating counts and caching declarative bundle loads in `bundleregistry.New()` while returning fresh registries.
- Final broad gate command passed; `make verify` also ran `go test -timeout 20m ./...`, docs validation, smoke, golangci-lint scoped checks, and `connectorgen validate`.
