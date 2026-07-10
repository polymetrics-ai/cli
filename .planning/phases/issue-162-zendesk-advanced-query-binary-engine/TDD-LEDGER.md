# TDD Ledger: Zendesk Advanced Query / Binary Engine

## Cycle 1 — binary manifest output policy

- Red target: direct reads with `output_policy=binary_manifest` return response metadata and never expose body bytes.
- Red evidence: pending.
- Green implementation: pending.

## Cycle 2 — Zendesk binary command coverage

- Red target: `cli_surface.json` has executable typed commands for all 37 Zendesk binary/file GET operations; the planned `binary candidates` placeholder is removed.
- Red evidence: pending.
- Green implementation: pending.

## Manual GSD fallback

`scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`); proceeding with manual TDD loop.
