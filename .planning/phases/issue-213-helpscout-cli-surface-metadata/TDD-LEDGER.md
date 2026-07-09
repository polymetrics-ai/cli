# TDD Ledger: Help Scout CLI Surface Metadata

Sub-issue: #213
Parent issue: #212

## Manual GSD Fallback

`programming-loop` is unavailable in `scripts/gsd`; manual GSD loop is active. Plan, red validation, green metadata, verification, and checkpoint commits are recorded here.

## Red Evidence Captured Before Production Edits

- Existing `internal/connectors/defs/help-scout/api_surface.json` is not full-surface: it lists 8 rows, while the official Inbox API docs navigation currently exposes 146 endpoint pages / 145 unique method-path pairs.
- Existing bundle has no `cli_surface.json`, so provider-style command/help metadata is absent.
- Existing docs claim 4 stream-backed endpoint groups and broad out-of-scope exclusions; this conflicts with the full-surface safety target that sensitive/admin/destructive operations should become typed, blocked/gated operations rather than blanket exclusions.

## Planned Red Tests / Validation

1. `go run ./cmd/connectorgen validate internal/connectors/defs` after initial metadata edits.
   - Expected red if any current stream is missing a covered surface row.
   - Expected red if `cli_surface.json` references a blocked/excluded endpoint as executable.
   - Expected red if `cli_surface.json` contains a secret-shaped literal.
2. `go test ./cmd/connectorgen -run CLISurface` if validator behavior changes are required.
3. `go test ./internal/connectors/engine -run CLISurface` if loader behavior changes are required.

## Green Evidence

Pending.

## Refactor Notes

Pending.
