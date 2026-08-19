# Issue #4290 TDD Ledger

## Red

- 2026-08-19 source-lock defect: the first implementation used existing `api_surface.json` as the operation boundary. Provider totals showed that a self-consistent map can omit most of a public API. The replacement assertions require the source-derived inventory to be explicit, pin a complete artifact/reference where possible, state dynamic/unavailable cases with a null count, and reject legacy `declared_percent`.
- 2026-08-19: none of the twenty assigned bundles had a connector-local operation source lock or declaration-disposition ledger. A complete-map inventory assertion cannot pass until both files exist and agree exactly with `api_surface.json`.
- Required failure assertions are: source lock exists; disposition map exists; normalized method/path sets are equal; each row has one valid parity class; every source DELETE appears; an unauthored row is not recorded as `foundation-gap`; absent `sync_transport.json` is recorded as ETL declaration-pending; typed write actions are enabled `direct_write`; and their reverse-ETL eligibility attribute uses the exact generic typed-destination executor gap with file/line evidence and minimal change.

## Green

- In progress — source-derived API-surface regeneration: the earlier Batch 4/5 Green evidence is superseded until each of the 20 provider inventories is verified against a complete public source, its `api_surface.json` regenerated where understated, and its map rematerialized.
- 2026-08-19 Mailchimp source recovery: the ordinary client fetched the public Swagger root but received repeat Akamai 503 responses for child path documents. The captain-approved Chrome fallback retrieved all 181 `$ref` path items serially. `mailchimp-browser-reference-crawl.json` pins each browser-retrieved source body; its 295 unique method/path operations are distinct from the pre-regeneration 298-row bundle. The resulting source inventory, regenerated 323-row surface (including 28 explicit local compatibility bindings), and map pass the connector-local materializer check.
- 2026-08-19 Batch 4: `node .planning/phases/issue-4290-map-parity-batches-45-r1/materialize-parity-maps.mjs write batch4` materialized ten public-source locks and exact crosswalk maps. The matching `check batch4` passed after asserting exact method/path inventory equality, required SHA-256/byte pins, DELETE coverage, enabled typed `direct_write` rows, and the nested reverse-ETL foundation attribute. `go run ./cmd/connectorgen validate` and `go run ./cmd/connectorgen surface-sync --check` passed.
- 2026-08-19 Batch 5 Red: `materialize-parity-maps.mjs check batch5` failed on the absent Pinterest source lock (`ENOENT`), proving the same complete-map invariant before materialization.
- 2026-08-19 Batch 5 Green: `write batch5` created all ten locks/maps and `check batch5` passed. TikTok Marketing and eBay Fulfillment retain the exact Chrome failures as `skipped: no-public-api-description`, with no SHA-256 or byte-count invented; all other Batch 5 source locks carry a SHA-256 and positive byte count.

## Refactor

- In progress: deterministic source-derived ordering, explicit runtime scope metadata, and review of disabled/pending vocabulary. The map checker rejects `reverse_etl` as a parity class and requires the generic typed-destination gap only as an eligibility attribute on enabled typed direct writes.
- 2026-08-19 engine rebase input: PR #4297 landed on `main`, closing the prior per-operation REST pagination and HEAD response executor gaps. The final rebase and declaration materialization must prove no retained row cites either retired gap; no connector-name branching is introduced here.
- 2026-08-19 captain merge freeze: issue #4303 is the required connector-neutral typed reverse-ETL destination. Until it lands, this lane must retain `generic-typed-destination-executor` only beneath already-enabled typed direct writes, identify all remaining provider mutations as `direct_write`, and never manufacture typed actions or `transport_binding` declarations without provider request contracts.
