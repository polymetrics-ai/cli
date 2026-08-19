## Task Delivery Header

- Issue: Refs #4289 — chore(connectors): map parity batches 2 and 3
- Base branch: main
- Merges into: main
- Delivery: A pull request against `main` from `fm/cli-map-batch23-r1`, with each selected bundle holding a public source lock and six-class disposition ledger; no credentials or provider calls are used.
- Working branch: fm/cli-map-batch23-r1
- Task: Produce source-locked, complete six-class maps for batch 2 (`grafana`, `trello`, `slack`, `n8n`, `google-calendar`, `gmail`, `twilio`, `amazon-sqs`, `elasticsearch`) and batch 3 (`gong`, `google-ads`, `facebook-marketing`, `linkedin-ads`, `aircall`, `xero`, `paypal-transaction`, `gocardless`, `amazon-seller-partner`, `miro`). Every documented source operation is bound to exactly one API-surface row and parity class. The task deliberately excludes Gitea and touches no engine or unrelated connector path.
- Verification: source-lock/disposition integrity checks; `go run ./cmd/connectorgen validate`; `go run ./cmd/connectorgen surface-sync --check`; targeted conformance and commandrunner tests; `connector-boundary` as a detached capture; repository generated/snapshot checks applicable to definition data.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every selected connector has auditable public provenance | live | Its source lock has the official artifact URL, capture date, SHA-256, byte count, `counts.total`, per-kind/method counts, and explicit `coverage_confidence` with a basis. The local verifier checks that captured inventory against the ledger without refetching mutable provider pages or discovery documents. A partial source is a hold, not a complete-map claim. |
| Every documented operation is accounted for exactly once | live | The integrity check proves source-inventory count equals disposition count and API-surface bindings, while primary parity-class totals equal that denominator. |
| Disabled operations have an honest reason | live | The verifier rejects missing state/rejection fields, treats un-authored contracts as `declaration-pending`, and requires file/line plus a minimal change for every `foundation-gap`. |
| No credentialed or live provider execution is claimed | live | Every ledger reports `live_certification: pending`; all source artifacts are public descriptions and all validation is local structural/fixture evidence. |

## Scope and Ownership

This is a map-only exception to the one-connector implementation lane: the issue owns nineteen disjoint connector bundle trees and their common planning evidence. Production edits are restricted to `internal/connectors/defs/<listed-connector>/sources/**`; no `operations.json`, executable command, schema, engine, or shared-tool change is in scope. The working directory name is `paypal-transaction`, which is the bundle corresponding to the issue's PayPal Transactions label.

## GSD Lifecycle and Inline Fallback

The adapter was checked with `scripts/gsd doctor`; its command sources were resolved for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`; generated prompts were reviewed. The canonical contract forbids role spawning and this task is running in a non-Pi inline worker, so the lifecycle is executed manually in this phase record. `go run ./cmd/agentcontractgen check` passed before planning.

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`. These are used only to assess existing engine contracts and validation truth; no Go implementation is added.

## Batch 2 — Red / Green / Refactor

1. **Red:** the nine target bundles lack `sources/<connector>-operation-source-lock.json` and `sources/<connector>-declaration-disposition.json`; therefore no provenance or six-class complete-map assertion can pass.
2. **Green:** credential-free public descriptions are pinned, parsed into method/path inventories, and mechanically reconciled to the existing `api_surface.json`. Every mismatch remains an explicit disabled disposition; no request, response, pagination, or body schema is inferred.
3. **Refactor:** normalize each ledger to the corrected batch-1 row shape: `method`, `path`, `parity_class`, `api_surface`, `source`, `state`, `foundation`, `rejection`, and `declaration`. An absent connector-local command or contract is `declaration-pending`, not `foundation-gap`. Preserve real engine gaps only with concrete refusal file/line and smallest safe change.
4. Commit batch 2 after its local map integrity and connector validation gates are green.

## Batch 3 — Red / Green / Refactor

1. **Red:** the ten target bundles likewise lack complete connector-local source locks and corrected six-class disposition ledgers.
2. **Green:** repeat the public-source provenance and exact source/API-surface crosswalk for `gong`, `google-ads`, `facebook-marketing`, `linkedin-ads`, `aircall`, `xero`, `paypal-transaction`, `gocardless`, `amazon-seller-partner`, and `miro`; each provider DELETE is declared or explicitly disabled.
3. **Refactor:** apply the same vocabulary and transport assessment as batch 2. PR #4286 makes a declaration-owned ETL source possible, so an absent source `sync_transport.json` is connector-local `declaration-pending`. A typed write action remains an enabled `direct_write`; its nested `declaration.reverse_etl_eligibility` records the real `generic-typed-destination-executor` foundation gap with `internal/app/issue_label_warehouse_transport.go:85-95` evidence and the required connector-neutral typed-destination minimal change. No `transport_binding` action or destination descriptor is invented.
4. Commit batch 3 after its local integrity and connector validation gates are green.

## Safety and Classification Rules

- Public API descriptions only; no credential, token, live provider request, provider write, or certification execution.
- `requires-elevated-scope` is enabled with source-backed runtime scope metadata, never a disabled reason.
- `unsafe-to-exercise` applies only to genuinely destructive/irreversible actions outside user intent. Documented deletes are required map rows.
- A `foundation-gap` requires engine refusal file/line plus minimal change. A missing command, operation contract, CLI surface, or transport declaration is `declaration-pending`.
- Endpoint classes are mutually exclusive: direct read, direct write (including delete), ETL, binary read, or binary write. `reverse_etl` is eligibility metadata nested under a typed `direct_write`, never an endpoint class: it remains foundation-blocked by `generic-typed-destination-executor` because `internal/app/issue_label_warehouse_transport.go:85-95` only registers and enforces the GitHub issue-label destination factory. ETL uses the merged PR #4286 declaration contract and remains `declaration-pending` until connector-local source evidence exists.

## Planned Gates and Checkpoints

1. Source/artifact capture and local source-lock/disposition integrity assertion for batch 2.
2. Batch-2 `connectorgen validate` and `surface-sync --check`, then commit and push checkpoint.
3. Repeat capture/integrity assertion and validation for batch 3, then commit and push checkpoint.
4. Run targeted fixture/conformance and commandrunner preflight tests, generated/check gates, detached `connector-boundary`, `verify-work`, and data-focused code review.
5. Rebase on `main`, push without force, open a Conventional Commit PR using `Refs #4289`, and read back its API-reported base.

## Source-Lock Completeness Remediation — 2026-08-19

Captain review found a fleet-wide defect: a source lock without `counts.total`, or a landing-page inventory that divides its own count by itself, cannot prove coverage. This phase replaces `declared_percent` with `operations_found` and `coverage_confidence`/basis everywhere, and makes `counts.total` plus per-kind/method counts mandatory. Gong, Google Ads, Miro, PayPal Transaction Search, and all Amazon Selling Partner models now use parsed machine-readable source inventories. Facebook Marketing and LinkedIn Ads are explicitly `partial` until their complete public provider references can be materialised; the branch must not be pushed or opened as a PR while either remains partial.
