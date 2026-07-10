# TDD Ledger: Zendesk Advanced Query / Binary Engine

## Cycle 1 — binary manifest output policy

- Red target: direct reads with `output_policy=binary_manifest` return response metadata and never expose body bytes.
- Red evidence: `go test ./internal/connectors/engine -run 'DirectRead.*Binary|BundleLoadZendeskDirectReadCommandCoverage' -count=1` failed because `binary_manifest` was unsupported.
- Green implementation: Added `binary_manifest` direct-read policy in the engine/command runner/schema/validator. It returns metadata only (`body_redacted`, `content_type`, `content_length`, `bytes_read`, `truncated`) and never emits binary body bytes.

## Cycle 2 — Zendesk binary command coverage

- Red target: `cli_surface.json` has executable typed commands for all 37 Zendesk binary/file GET operations; the planned `binary candidates` placeholder is removed.
- Red evidence: `go test ./internal/connectors/engine -run 'DirectRead.*Binary|BundleLoadZendeskDirectReadCommandCoverage' -count=1` failed because Zendesk still exposed the planned `binary candidates` placeholder.
- Green implementation: Converted all 37 Zendesk binary/file-like GETs into `binary ...` direct-read commands with `output_policy=binary_manifest`, API coverage, path flags, and operation ledger max-byte policy; removed planned read/write/destructive placeholders.

## Manual GSD fallback

`scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`); proceeding with manual TDD loop.
