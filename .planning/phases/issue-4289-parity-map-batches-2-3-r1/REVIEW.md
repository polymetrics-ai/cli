# Issue #4289 — Data-Focused Code Review

## Scope reviewed

The two mapping commits add source locks and declaration-disposition ledgers for only the nineteen issue-owned connector definition trees, update their source-derived API-surface blocking metadata where required, and add the phase evidence. No engine, CLI, operation contract, executable command, schema, credential, or `sync_transport.json` was changed.

## Findings

One delivery blocker is open: the Facebook Marketing and LinkedIn Ads provider sources are marked `coverage_confidence: partial`. Their prior landing-page inventories (36 and 10 operations) are no longer presented as complete; no PR may be opened until complete provider references are materialised and mapped.

- The verifier proves 3,348 currently found source inventory rows have exactly one corrected Batch-1-shaped disposition and a bound API-surface endpoint. It also requires `counts.total`, per-kind counts, and a confidence basis for every lock.
- Typed actions now remain enabled `direct_write`. Reverse-ETL is nested eligibility metadata, not an endpoint parity class; every candidate carries only `generic-typed-destination-executor` with the required `internal/app/issue_label_warehouse_transport.go:85-95` evidence and minimal change.
- ETL remains an endpoint class but is declaration-pending until its connector authors the #4286 source `sync_transport.json`; the maps do not fabricate a destination binding or transport action.
- All documented DELETE rows were inspected by the verifier/report. They are either enabled where a typed direct-write binding already exists or explicitly disabled as declaration-pending.
- `rg transport_binding` over all nineteen connector directories found no synthetic transport binding.

## Review evidence

`git diff --check`, source-map verification, `connectorgen validate`, `surface-sync --check`, targeted conformance, and commandrunner runtime preflight all passed before the source-lock remediation. The long-running `connector-boundary` scanner completed cleanly in its captured trace. Those structural passes do not close the two source-completeness holds.
