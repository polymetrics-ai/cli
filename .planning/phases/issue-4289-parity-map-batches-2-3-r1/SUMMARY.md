# Issue #4289 — Delivery Summary

Refs #4289

Two committed map slices now cover all nineteen requested connector bundles. Source-lock remediation replaces the prior self-referential `declared_percent` with `operations_found`, `counts.total`, per-kind/method counts, and a confidence basis for every lock. The current source-accounted inventory is 3,348 operations; it is not yet a final complete-map delivery because Facebook Marketing and LinkedIn Ads remain explicitly `partial`.

The classification correction is deliberate: typed write actions are enabled `direct_write`; `reverse_etl` is represented only as nested eligibility metadata and is foundation-blocked by `generic-typed-destination-executor` at `internal/app/issue_label_warehouse_transport.go:85-95`. ETL is a real endpoint class but remains connector-local `declaration-pending` until a source `sync_transport.json` meets the PR #4286 contract. No destination binding/action was invented.

Commits:

- `11c5d8132 chore(connectors): map parity batch 2 Refs #4289`
- `3353fa94e chore(connectors): map parity batch 3 Refs #4289`

Local verification is recorded in `VERIFICATION.md`; data-focused review is recorded in `REVIEW.md`. No PR is to be opened until the two partial source references become complete.
