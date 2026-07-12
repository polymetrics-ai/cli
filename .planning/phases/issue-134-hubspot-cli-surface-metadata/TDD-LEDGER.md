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

Pending.

## Green evidence

Pending.

## Refactor evidence

Pending.

## Safety assertions to preserve

- No HubSpot command may use `intent: "raw_api"`.
- No HubSpot command may use `intent: "direct_write"`.
- Binary/file commands use `intent: "binary"` plus typed operation metadata and remain non-executable until a bounded binary policy lands.
- Mutations use `intent: "reverse_etl"` and planned/blocked typed metadata until writes are modeled as named actions in `writes.json`.
- No credentialed checks or live writes.
