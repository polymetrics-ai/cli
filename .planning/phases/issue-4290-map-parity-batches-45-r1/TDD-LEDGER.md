# Issue #4290 TDD Ledger

## Red

- 2026-08-19: none of the twenty assigned bundles had a connector-local operation source lock or declaration-disposition ledger. A complete-map inventory assertion cannot pass until both files exist and agree exactly with `api_surface.json`.
- Required failure assertions are: source lock exists; disposition map exists; normalized method/path sets are equal; each row has one valid parity class; every source DELETE appears; an unauthored row is not recorded as `foundation-gap`; absent `sync_transport.json` is recorded as ETL declaration-pending; typed write actions are enabled `direct_write`; and their reverse-ETL eligibility attribute uses the exact generic typed-destination executor gap with file/line evidence and minimal change.

## Green

- 2026-08-19 Batch 4: `node .planning/phases/issue-4290-map-parity-batches-45-r1/materialize-parity-maps.mjs write batch4` materialized ten public-source locks and exact crosswalk maps. The matching `check batch4` passed after asserting exact method/path inventory equality, required SHA-256/byte pins, DELETE coverage, enabled typed `direct_write` rows, and the nested reverse-ETL foundation attribute. `go run ./cmd/connectorgen validate` and `go run ./cmd/connectorgen surface-sync --check` passed.
- Pending Batch 5: materialize the next ten locks/maps, with explicit `skipped: no-public-api-description` source-lock evidence for the TikTok and eBay browser failures, then repeat the inventory assertion and generation checks.

## Refactor

- Pending: deterministic JSON ordering, explicit runtime scope metadata, and review of disabled/pending vocabulary.
