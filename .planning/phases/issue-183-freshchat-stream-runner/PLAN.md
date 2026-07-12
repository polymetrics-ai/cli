# Issue #183 — Freshchat stream runner

Parent issue: #180
Sub-issue: #183
Branch: `feat/183-freshchat-stream-runner`

## Scope

Implement/verify Freshchat stream runner parity for all implemented ETL command-surface entries. This slice stays read-only and fixture-backed:

- every Freshchat `cli_surface.json` command with `intent=etl`, `availability=implemented`, and `stream` must map to a declared stream;
- every mapped stream must have replay fixtures;
- the real declarative engine must read at least one fixture record for every mapped stream;
- pagination/cursor/PK metadata stays static and validated; no live credentials or writes.

## GSD/TDD

- GSD plan prompt: `scripts/gsd prompt plan-phase issue-183-freshchat-stream-runner --skip-research`.
- Programming loop adapter command attempted: `scripts/gsd prompt programming-loop init --phase issue-183-freshchat-stream-runner --dry-run`.
- Adapter result: unavailable (`unknown GSD command: programming-loop`); manual programming-loop fallback applies.

## Required skills used

- gsd-core
- golang-how-to
- golang-cli
- golang-testing
- golang-error-handling
- golang-security
- golang-safety
- golang-documentation

## Plan

1. Add a red conformance regression test that loads the on-disk Freshchat bundle and replay-runs every implemented ETL command stream.
2. Add missing stream fixtures for the current 18 Freshchat ETL streams.
3. Run focused conformance/connectorgen gates.
4. Run full local verification before opening a stacked PR.
5. Preserve safety: no credentialed Freshchat calls, no reverse ETL execution, no binary/upload path, and no raw HTTP escape hatch.
