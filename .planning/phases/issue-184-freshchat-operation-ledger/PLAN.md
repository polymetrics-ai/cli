# Plan — Issue #184 Freshchat operation ledger

Refs #184, parent #180.

## GSD / skills

- `scripts/gsd prompt plan-phase issue-184-freshchat-operation-ledger --skip-research` generated successfully.
- `scripts/gsd prompt programming-loop init --phase issue-184-freshchat-operation-ledger --dry-run` is unavailable (`unknown GSD command: programming-loop`); manual programming-loop fallback is active.
- Required skills used: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-context, golang-concurrency, golang-documentation, golang-lint.

## Scope

- Convert Freshchat `api_surface.json` to operation-ledger mode (`operation_ledger_version: 1`).
- Keep implemented stream/write endpoint coverage intact.
- Replace legacy `excluded` rows with blocked operation rows using the closed operation vocabulary.
- Add narrow validation tests proving the Freshchat ledger accounts for the official 34-endpoint baseline.

## Non-goals

- No Freshchat credentialed checks.
- No direct-read executor work (#185).
- No multipart/binary upload executor work (#186).
- No help/docs renderer changes (#182).
- No reverse ETL execution.

## Implementation slices

1. Red test: add Freshchat api-surface operation-ledger metrics test expecting ledger version 1, 34 endpoints, 18 stream rows, 13 write rows, and 3 blocked operation rows.
2. Metadata: update `internal/connectors/defs/freshchat/api_surface.json` from legacy `excluded` rows to operation ledger rows.
3. Validation: run targeted connectorgen/Freshchat tests and full connector definition validation.
4. Handoff: commit/push stacked PR against `feat/180-freshchat-cli-parity`.

## Safety decisions

- `POST /users/fetch` is blocked as `direct_read`: read-like request-body criteria are not executable until bounded direct-read body support exists.
- `POST /files/upload` and `POST /images/upload` are blocked as `disallowed`: multipart/binary upload execution is not exposed without explicit binary/file policy work.
- All blocked rows stay `blocked_by_default: true` with a reason and source/notes when required by validation.
