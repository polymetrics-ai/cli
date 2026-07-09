# Summary: Chatwoot CLI Surface Metadata

Status: implemented locally; sub-PR pending.

## Delivered

- Refreshed Chatwoot `api_surface.json` from the official Swagger source with 144 operations across 89 paths.
- Switched Chatwoot surface accounting to `operation_ledger_version: 1` with blocked-by-default operation rows for non-executable direct-read, sensitive/admin reverse-ETL, destructive, duplicate, and disallowed candidates.
- Preserved current executable coverage for 7 streams and 6 write actions.
- Added `cli_surface.json` mapping Chatwoot-shaped commands to implemented streams/writes and planned/blocked safe intents.
- Updated metadata/docs to state full-surface accounting without overclaiming unsupported direct-read, binary/multipart, or admin execution.

## Verification

- `python3 .planning/phases/issue-149-chatwoot-cli-surface-metadata/traces/verify-official-surface-count.py`: pass.
- `jq empty internal/connectors/defs/chatwoot/api_surface.json internal/connectors/defs/chatwoot/cli_surface.json internal/connectors/defs/chatwoot/metadata.json`: pass.
- `go test ./cmd/connectorgen -run CLISurface -count=1`: pass.
- `go test ./internal/connectors/engine -run CLISurface -count=1`: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass.
- `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`: pass.

## Next

1. Commit and push #149 branch.
2. Open sub-PR against `feat/148-chatwoot-cli-parity` with `Refs #149` and `Refs #148`.
3. Route automated review per stacked PR rules.
