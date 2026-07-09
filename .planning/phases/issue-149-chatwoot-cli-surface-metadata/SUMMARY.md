# Summary: Chatwoot CLI Surface Metadata

Status: sub-PR open (#227); remote checks/review coverage pending.

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
- `go vet ./...`: pass.
- `go test ./...`: pass after project-timeout `make verify` populated the long-running certify result cache.
- `go build ./cmd/pm`: pass.
- `make verify`: pass.

## Next

1. Wait for PR #227 checks to finish.
2. CodeRabbit skipped PR #227 because base is non-default; record parent PR #223 fallback coverage as pending.
3. Do not merge #227 into parent until checks are green and review coverage route is satisfied or explicitly recorded as pending/provisional per stacked workflow.
