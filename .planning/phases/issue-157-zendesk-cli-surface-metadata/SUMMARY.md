# Summary: Zendesk CLI Surface Metadata

Status: verified locally; sub-PR pending.

## Scope

Initial metadata-only Zendesk bundle scaffold and command/API surface inventory from the official OAS. Added `internal/connectors/defs/zendesk/` with metadata/spec/check scaffold, full blocked-by-default `api_surface.json`, `cli_surface.json` command inventory, and docs.

## Not in this slice

- Runtime help rendering or command dispatch.
- Stream schemas/fixtures beyond minimal scaffold.
- Direct-read executors.
- Reverse-ETL write schemas or destructive confirmations.
- Credentialed Zendesk checks.

## Evidence

- Official OAS baseline matched: 617 operations across 429 paths.
- `api_surface.json` operation rows: direct_read=282, binary_read=37, sensitive_reverse_etl=210, destructive_action=85, deprecated=3.
- `cli_surface.json` command rows: 5 planned/docs-only category rows across 3 groups; full per-operation detail stays in `api_surface.json` so runtime bundle loading remains bounded.
- Targeted tests, broad `go test ./...`, `make verify`, and `connectorgen validate` passed.

## Next

Commit the green implementation slice, push the sub-issue branch, and open a stacked PR to the parent branch.
