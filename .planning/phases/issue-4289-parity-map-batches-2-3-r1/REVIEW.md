# Issue #4289 — Data-Focused Code Review

## Scope reviewed

The two mapping commits add source locks and declaration-disposition ledgers for only the nineteen issue-owned connector definition trees, update their source-derived API-surface blocking metadata where required, and add the phase evidence. No engine, CLI, operation contract, executable command, schema, credential, or `sync_transport.json` was changed.

## Findings

The source-lock delivery blocker was reopened by the held-PR check and repaired. Facebook Marketing is pinned to the complete provider-published Business SDK code-generation archive (1,445 named owner-type/method/node-edge declarations, with explicit instance-dependent Graph-path basis); LinkedIn Ads is pinned to every Marketing page in the provider sitemap (272 operations after removing literal examples covered by documented templates); and PayPal is now pinned to all thirteen provider-published REST OpenAPI documents (115 declarations rather than the incomplete two-route Transaction Search document). `SOURCE-LOCK-VERIFICATION.json` records old/new API-surface counts, root totals, provider basis, and confidence for all nineteen connectors; every lock is `complete`.

- The verifier proves 5,127 currently found source inventory rows have exactly one corrected Batch-1-shaped disposition and a bound API-surface endpoint. It requires root `counts.total`, nested REST per-kind counts, an old/new API-surface report, and a confidence basis for every lock.
- Typed actions remain enabled `direct_write`. Reverse-ETL is eligibility metadata, not an endpoint parity class; all 621 schema-backed actions have an explicit semantic eligibility disposition. The ledger also marks the absent exact stream-to-required-input binding as `declaration-pending` and separates the persisted App/CLI dispatch hold (`internal/app/transport_dispatch.go:53-67`) from the one-action-per-mode selection limit (`internal/connectors/sync_transport.go:388-415`), so no risk, scope, or destructive classification becomes an exclusion and no one-action proof is treated as connector completion.
- ETL remains an endpoint class but is declaration-pending until its connector authors the #4286 source `sync_transport.json`; the maps do not fabricate a destination binding or transport action.
- All documented DELETE rows were inspected by the verifier/report. They are either enabled where a typed direct-write binding already exists or explicitly disabled as declaration-pending.
- `rg transport_binding` over all nineteen connector directories found no synthetic transport binding.

## Review evidence

`git diff --check`, source-map verification, `connectorgen validate`, `surface-sync --check`, targeted conformance, commandrunner runtime preflight, `internal/cli`, build, vet, generated/snapshot gates, and the detached `connector-boundary` scanner all pass after the complete-source remediation.

## GSD lifecycle note

`scripts/gsd sources code-review` and `scripts/gsd prompt code-review` were resolved. The shell session has no compatible Pi runtime for `/gsd-code-review`, so the documented inline/manual fallback was used: review covered the source parser/crawler paths, public-source accounting, baseline-binding preservation, row schema, six-class/reverse-ETL invariants, and the final gate evidence above. No Critical, Warning, or Info finding remains.
