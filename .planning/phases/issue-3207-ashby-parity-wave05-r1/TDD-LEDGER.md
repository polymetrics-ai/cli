# TDD Ledger — Ashby parity wave05-r1

## Red targets before production edits

- Inventory script should report current official Ashby OpenAPI REST + webhook counts and fail if any operation is unclassified or duplicated.
- `connectorgen validate internal/connectors/defs/ashby` should fail on the current 4-row quarantine ledger because it is incomplete against the refreshed inventory (manual inventory check, not existing validator).
- Ashby connector tests should fail until native metadata/catalog/command surface matches expanded bundle counts.

## Green targets

- Generated Ashby bundle loads with complete api_surface coverage, schemas, write schemas, CLI surface, operations, and docs.
- Native Ashby read fixture mode emits records for every declared stream and preserves context cancellation/limit behavior.
- Typed direct read and reverse-ETL write paths validate/dry-run through existing engine contracts without provider calls.
- Generated docs/skills/manual inspect surfaces include current Ashby streams/actions/commands and safety notes.

## Evidence log

- 2026-08-01: GSD programming-loop command unavailable; manual GSD/TDD fallback recorded in PLAN.md.
- 2026-08-01 red inventory: public Ashby ReadMe embedded OpenAPI reports REST=185 plus webhooks=27 (total=212); current `api_surface.json` has 4 rows (gap=208).
- 2026-08-01 green generation: Ashby parity generator produced REST=185, webhooks=27, streams=72, writes=101, direct_reads=7, blocked=32 with no duplicate api_surface rows.
- 2026-08-01 green validation: `go run ./cmd/connectorgen validate internal/connectors/defs` passed with 550 connectors and 0 findings.
- 2026-08-01 green tests: Ashby conformance, native/hook tests, CLI `Connector|Dynamic|Golden`, vet/build/boundary/diff-check, and full `make verify` all passed (see VERIFICATION.md for exact commands/results).
