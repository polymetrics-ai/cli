# TDD Ledger: Zendesk Operation Ledger

## Cycle 1 — ledger completeness

- Red target: embedded Zendesk bundle requires `operations.json` with 617 operations and total API surface accounting.
- Red evidence: `go test ./internal/connectors/engine -run 'ZendeskOperationLedger' -count=1` failed with `Zendesk Operations count = 0, want 617`.
- Green implementation: added `internal/connectors/defs/zendesk/operations.json` with 617 typed operation specs, refreshed `api_surface.json` blocked-operation reasons, and added a disk-backed ledger test that checks every surface row has a matching operation by method/path.
- Refactor notes: kept the ledger non-executable; `covered_by` remains reserved for later executable stream/direct-read/write lanes, so `api_surface.json` uses operation rows only in this slice.

## Manual GSD fallback

`scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run` failed with `unknown GSD command: programming-loop`; proceeding with manual red/green/refactor loop per AGENTS.
