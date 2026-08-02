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

## Review fix round 1 targets

- Ashby catalogs and bundle manifests expose full-refresh modes only; saved timestamp state never filters records, and caller-supplied `syncToken` is blocked on `ashby-sync-token-checkpoint-foundation`.
- Checks, stream reads, direct reads, and writes reject malformed or false Ashby success envelopes without returning response secrets.
- `create_application.applicationHistory` rejects undeclared nested properties while documented map-valued request fields remain open only where the OpenAPI explicitly models a map.
- A repeated `nextCursor` fails before the cursor is requested again.
- Ashby stream commands expose no `string_array` flags and name `connector-stream-repeatable-array-foundation` for the withheld variants.

## Review fix round 1 evidence

- 2026-08-02 red: focused Ashby regression command failed on nested arbitrary-field acceptance, timestamp-state filtering, unblocked `syncToken`, repeated page cursor acceptance, advertised incremental modes, executable stream array flags, missing success acceptance, and sibling check/direct-read/write false-success acceptance; the documented map-field and successful-envelope controls remained green.
- 2026-08-02 focused verification first exposed that the existing schema compiler accepts only boolean `additionalProperties`; documented Ashby map fields were adjusted to explicit `true` while every other modeled object remains closed, without changing shared engine code.
- 2026-08-02 green: `go test ./internal/connectors/native/ashby` passed.
