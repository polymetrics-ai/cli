# Issue #4289 — Data-Focused Code Review

## Scope reviewed

The two mapping commits add source locks and declaration-disposition ledgers for only the nineteen issue-owned connector definition trees, update their source-derived API-surface blocking metadata where required, and add the phase evidence. No engine, CLI, operation contract, executable command, schema, credential, or `sync_transport.json` was changed.

## Findings

The source-lock delivery blocker is resolved. Facebook Marketing is now pinned to the complete provider-published Business SDK code-generation archive (1,445 named owner-type/method/node-edge declarations, with explicit instance-dependent Graph-path basis); LinkedIn Ads is pinned to every Marketing page in the provider sitemap (272 operations after removing literal examples covered by documented templates). `SOURCE-LOCK-VERIFICATION.json` records old/new API-surface counts, provider basis, total operation count, and confidence for all nineteen connectors; every lock is `complete`.

- The verifier proves 3,348 currently found source inventory rows have exactly one corrected Batch-1-shaped disposition and a bound API-surface endpoint. It also requires `counts.total`, per-kind counts, and a confidence basis for every lock.
- Typed actions now remain enabled `direct_write`. Reverse-ETL is nested eligibility metadata, not an endpoint parity class; every candidate carries only `generic-typed-destination-executor` with the required `internal/app/issue_label_warehouse_transport.go:85-95` evidence and minimal change.
- ETL remains an endpoint class but is declaration-pending until its connector authors the #4286 source `sync_transport.json`; the maps do not fabricate a destination binding or transport action.
- All documented DELETE rows were inspected by the verifier/report. They are either enabled where a typed direct-write binding already exists or explicitly disabled as declaration-pending.
- `rg transport_binding` over all nineteen connector directories found no synthetic transport binding.

## Review evidence

`git diff --check`, source-map verification, `connectorgen validate`, and `surface-sync --check` pass after the complete-source remediation. Targeted conformance, commandrunner runtime preflight, and the long-running `connector-boundary` scanner passed before remediation and will be repeated against the final commit before PR progression.
