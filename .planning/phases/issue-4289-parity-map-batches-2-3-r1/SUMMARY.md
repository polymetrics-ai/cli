# Issue #4289 — Delivery Summary

Refs #4289

Two committed map slices now cover all nineteen requested connector bundles. Source-lock remediation replaces the prior self-referential `declared_percent` with `operations_found`, root `counts.total`, detailed per-kind/method counts, and a confidence basis for every lock. The complete source-accounted inventory is now 5,127 operations. Each connector's old/new `api_surface` count and provider basis are published in `SOURCE-LOCK-VERIFICATION.json`; the surface is no longer bounded by the earlier connector data.

The held-PR repair replaced the incomplete two-route PayPal Transaction Search pin with all thirteen official PayPal REST OpenAPI documents, yielding 115 exact PayPal operations. It also restores the Batch-2-shaped root count object for every batch-3 source lock and makes that root denominator a verifier requirement.

The classification correction is deliberate: typed write actions are enabled `direct_write`; `reverse_etl` is represented only as nested and action-level eligibility metadata. All 621 existing schema-backed write actions are semantically eligible and individually representable by the closed generic destination. Reverse-ETL is still foundation-pending because persisted App/CLI dispatch does not select that already-composed destination (`internal/app/transport_dispatch.go:53-67`) and one descriptor can select only one named action per sync mode (`internal/connectors/sync_transport.go:388-415`). ETL remains a real endpoint class but is connector-local `declaration-pending` until a source `sync_transport.json` meets the PR #4286 contract. No destination binding, selector, or action was invented.

Commits:

- `11c5d8132 chore(connectors): map parity batch 2 Refs #4289`
- `3353fa94e chore(connectors): map parity batch 3 Refs #4289`

Local verification is recorded in `VERIFICATION.md`; data-focused review is recorded in `REVIEW.md`. Final scoped gates, rebase, and the direct-PR review workflow remain before a PR may be opened.
