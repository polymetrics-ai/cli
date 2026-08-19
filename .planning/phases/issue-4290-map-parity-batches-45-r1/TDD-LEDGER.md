# Issue #4290 TDD Ledger

## Red

- 2026-08-19: none of the twenty assigned bundles had a connector-local operation source lock or declaration-disposition ledger. A complete-map inventory assertion cannot pass until both files exist and agree exactly with `api_surface.json`.
- Required failure assertions are: source lock exists; disposition map exists; normalized method/path sets are equal; each row has one valid parity class; every source DELETE appears; and an unauthored row is not recorded as `foundation-gap`.

## Green

- Pending Batch 4: materialize the ten public-source locks and exact crosswalk maps, then run the inventory assertion and generation checks.
- Pending Batch 5: materialize the next ten locks/maps, then repeat the inventory assertion and generation checks.

## Refactor

- Pending: deterministic JSON ordering, explicit runtime scope metadata, and review of disabled/pending vocabulary.

