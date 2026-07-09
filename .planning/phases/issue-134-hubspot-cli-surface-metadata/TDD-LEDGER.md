# TDD Ledger: Issue #134 HubSpot CLI Surface Metadata

Date: 2026-07-10

## Planned red tests

- `go test ./cmd/connectorgen -run TestValidate_CLISurfaceImplementedBinaryOperationPasses -count=1`
  - Expected initial failure before production edits: `binary` is not allowed by `cli_surface.schema.json` or the validator treats it as an unsupported executable intent.
- `go test ./cmd/connectorgen -run TestValidate_CLISurfaceImplementedBinaryRequiresTypedOperation -count=1`
  - Expected initial failure before production edits: no validator rule exists for binary command operation references.
- `go test ./cmd/connectorgen -run TestValidate_HubSpotCLISurfaceMetadata -count=1`
  - Expected initial failure before production edits: no `internal/connectors/defs/hubspot/cli_surface.json` bundle exists.

## Red evidence

- `go test ./cmd/connectorgen -run 'TestValidate_CLISurfaceImplementedBinaryOperationPasses|TestValidate_CLISurfaceImplementedBinaryRequiresTypedOperation|TestValidate_HubSpotCLISurfaceMetadata' -count=1`
  - Failed as expected before production edits:
    - `binary` intent rejected by `cli_surface.schema.json` enum.
    - binary-without-operation test saw `meta_schema` before the intended safety rule.
    - HubSpot bundle missing: `ConnectorsChecked = 0, want 1`.

## Green evidence

- `gofmt -w cmd/connectorgen/main_test.go cmd/connectorgen/validate.go` — passed.
- `go test ./cmd/connectorgen -run 'TestValidate_CLISurfaceImplementedBinaryOperationPasses|TestValidate_CLISurfaceImplementedBinaryRequiresTypedOperation|TestValidate_HubSpotCLISurfaceMetadata' -count=1` — passed after adding binary intent validation and HubSpot metadata scaffold.
- `go test ./cmd/connectorgen -run 'CLISurface|HubSpot' -count=1` — passed.
- `go test ./internal/connectors/engine -run 'CLISurface|HubSpot' -count=1` — passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — passed: 548 connector(s) checked, 0 findings.
- `go test ./cmd/connectorgen ./internal/connectors/engine` — passed.
- `go run ./cmd/pm help docs` — read before docs command use.
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors` — passed after HubSpot generated docs/catalog updates.
- `go run ./cmd/pm help connectors` and `go run ./cmd/pm connectors inspect hubspot --json` — passed; manifest inspection is credential-free.
- Full gate command passed: `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify && go run ./cmd/connectorgen validate internal/connectors/defs`.

## Refactor evidence

- HubSpot operations metadata uses typed `rest_read`, `binary_download`, and `composite` operation kinds only. POST search remains a non-executable `composite` metadata operation until #139 adds a fixed body/query shape.
- Planned reverse ETL commands do not reference `operations.json`; future write lanes must add named `writes.json` actions before any write becomes executable.
- All HubSpot commands are `availability: planned`; no HubSpot command is executable in the metadata slice.
- `cli_surface.json` contains no `raw_api` or `direct_write` intents.
- Adding the 548th declarative bundle required updating connector-count assertions/help docs and caching declarative bundle loads in `internal/connectors/bundleregistry` so cold full-suite certification tests stay under the default `go test` package timeout while each `New()` call still returns a fresh registry.

## Safety assertions to preserve

- No HubSpot command may use `intent: "raw_api"`.
- No HubSpot command may use `intent: "direct_write"`.
- Binary/file commands use `intent: "binary"` plus typed operation metadata and remain non-executable until a bounded binary policy lands.
- Mutations use `intent: "reverse_etl"` and planned/blocked typed metadata until writes are modeled as named actions in `writes.json`.
- No credentialed checks or live writes.
