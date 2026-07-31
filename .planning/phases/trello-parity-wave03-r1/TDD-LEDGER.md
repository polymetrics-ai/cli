# Trello parity wave 03 TDD ledger

## Red targets before production edits

1. `cmd/connectorgen/trello_api_surface_test.go` must fail against the current Trello bundle because `api_surface.json` has only 4 rows and no `operation_ledger_version: 1`, while the re-audited official OpenAPI has 261 HTTP operations.
2. `go run ./cmd/connectorgen validate internal/connectors/defs/trello` must validate the Trello bundle directly, not nested `fixtures/` or `schemas/` directories.

## Green targets

- `go test ./cmd/connectorgen -run TrelloAPISurfaceOperationLedger -count=1` passes with 261 total rows, 219 covered rows, 42 blocked rows, and method/model counts matching the official audit.
- `go run ./cmd/connectorgen validate internal/connectors/defs/trello` passes with no findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/trello' -count=1` passes and exercises the Trello check fixture, stream fixtures, write request fixtures, and delete semantics.

## Evidence log

- Red captured: `gofmt -w cmd/connectorgen/trello_api_surface_test.go && go test ./cmd/connectorgen -run TrelloAPISurfaceOperationLedger -count=1` failed as expected on the baseline bundle with `operation_ledger_version = 0, want 1`.
- Green captured: `go test ./cmd/connectorgen -run TrelloAPISurfaceOperationLedger -count=1` passed.
- Green captured: `go run ./cmd/connectorgen validate internal/connectors/defs/trello` passed after adding single-bundle validation support and expanding Trello.
- Green captured: `go test ./internal/connectors/conformance -run 'TestConformance/trello' -count=1` passed.
- Full `make verify` initially exposed registry-load timeout pressure from modeling every GET as a stream; the implementation was adjusted to 3 fixture-backed streams plus 95 fixed direct reads, then `make verify` passed.
