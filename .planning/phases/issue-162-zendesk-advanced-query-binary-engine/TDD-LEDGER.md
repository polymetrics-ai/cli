# TDD Ledger: Zendesk Advanced Query / Binary Engine

## Cycle 1 — binary manifest output policy

- Red target: direct reads with `output_policy=binary_manifest` return response metadata and never expose body bytes.
- Red evidence: `go test ./internal/connectors/engine -run 'DirectRead.*Binary|BundleLoadZendeskDirectReadCommandCoverage' -count=1` failed because `binary_manifest` was unsupported.
- Green implementation: pending.

## Cycle 2 — Zendesk binary command coverage

- Red target: `cli_surface.json` has executable typed commands for all 37 Zendesk binary/file GET operations; the planned `binary candidates` placeholder is removed.
- Red evidence: `go test ./internal/connectors/engine -run 'DirectRead.*Binary|BundleLoadZendeskDirectReadCommandCoverage' -count=1` failed because Zendesk still exposed the planned `binary candidates` placeholder.
- Green implementation: pending.

## Manual GSD fallback

`scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`); proceeding with manual TDD loop.
