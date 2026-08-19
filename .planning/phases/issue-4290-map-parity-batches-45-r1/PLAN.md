## Task Delivery Header

- Issue: Refs #4290 — chore(connectors): map parity batches 4 and 5
- Base branch: main
- Merges into: main
- Delivery: A pull request against `main`, with the two ten-connector parity-map increments committed separately, source locks and six-class disposition ledgers present for every assigned bundle, and the named local validation gates green.
- Working branch: fm/cli-map-batch45-r1
- Task: Produce source-locked, complete operation disposition maps for Batch 4 (`salesforce`, `hubspot`, `pipedrive`, `mailchimp`, `zendesk-support`, `quickbooks`, `bamboo-hr`, `airtable`, `google-analytics-data-api`, `woocommerce`) and Batch 5 (`pinterest`, `tiktok-marketing`, `linear`, `buildkite`, `sonar-cloud`, `launchdarkly`, `fastly`, `squarespace`, `ebay-fulfillment`, `shipstation`). Every normalized, documented `api_surface.json` endpoint must occur exactly once and use the proven batch-1 row shape.
- Verification: Source-lock/map inventory assertions; `go run ./cmd/connectorgen validate`; `go run ./cmd/connectorgen surface-sync --check`; the connector-boundary check; `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/engine`; and generated/docs checks that apply to definitions-only changes.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every Batch 4 normalized documented endpoint has one map row | live | A local inventory assertion compares the method/path multiset in each Batch 4 `api_surface.json` with its source lock and disposition map; removing a row makes the assertion fail. |
| Every Batch 5 normalized documented endpoint has one map row | live | The same assertion compares each Batch 5 bundle; it detects duplicate and missing rows. |
| Row vocabulary distinguishes unauthored work from engine incapability | live | The assertion rejects any `foundation-gap` rejection without engine file/line evidence and minimal change, and checks unauthored rows use `declaration-pending`. |
| Existing connector definitions remain valid and synchronized | live | `connectorgen validate` and `surface-sync --check` parse the real affected bundles and report no drift. |
| No credentialed provider operation is performed | fake | This is a documentation-only mapping lane; locks are public descriptions retrieved without credentials, while all provider behavior remains unexercised. |

## Scope and Ownership Guard

- Allowed production paths are only `internal/connectors/defs/{salesforce,hubspot,pipedrive,mailchimp,zendesk-support,quickbooks,bamboo-hr,airtable,google-analytics-data-api,woocommerce,pinterest,tiktok-marketing,linear,buildkite,sonar-cloud,launchdarkly,fastly,squarespace,ebay-fulfillment,shipstation}/sources/`.
- This phase also owns its required `.planning/phases/issue-4290-map-parity-batches-45-r1/` evidence. No engine, generator, schema, source-operation contract, CLI surface, or other connector files may change.
- The issue body has three trailing connector names after the explicit Batch 5 ten. The launch brief's exact twenty-connector list is authoritative; `gocardless`, `amazon-seller-partner`, and `miro` are deliberately not touched.
- `bamboohr` and `sonarcloud` resolve to the repository-owned `bamboo-hr` and `sonar-cloud` definition directories.

## Locked Disposition Rules

1. `foundation-gap` means the current engine refuses the documented shape. It must name the refusal file and line plus the smallest viable engine change. Missing operation contracts, commands, or CLI surfaces are `declaration-pending`, never foundation gaps.
2. `requires-elevated-scope` is enabled with source/runtime scope metadata; it is never a rejection or disabled state.
3. Every DELETE is represented. A delete requested by this issue is not `unsafe-to-exercise` merely because it mutates data.
4. Source maps reproduce the batch-1 shape: `method`, `path`, `parity_class`, `api_surface`, `source`, `state`, `foundation`, `rejection`, and `declaration`. Existing normalized `api_surface.json` endpoint inventory is the connector-local crosswalk; each lock records its public documentation source and exact endpoint inventory.
5. No provider credential, tenant discovery, API request, or write is used. Salesforce tenant-defined object/field dependencies remain named runtime metadata, not discovered facts.

## TDD Delivery Plan

### Red — complete-map invariants

Before materialization, every target lacked both `sources/<connector>-operation-source-lock.json` and `sources/<connector>-declaration-disposition.json`; an inventory assertion therefore has no source lock/map to compare and fails. The test first requires source-lock/map presence, exact method/path multiset equality with `api_surface.json`, valid six-class values, DELETE coverage, and vocabulary correctness.

### Green — Batch 4

Pin public description evidence and materialize the ten Batch 4 locks/maps from their normalized provider inventories. Preserve every `covered_by` binding; classify streams as `etl`, typed write bindings as `reverse_etl`, bounded operation bindings as direct read/write, and documented binary endpoints as binary read/write. A missing terminal declaration is pending, not a gap. Commit only the Batch 4 source directories after its assertions and connector generation gates pass.

### Green — Batch 5

Repeat the same source-lock and full-inventory materialization for Batch 5, including GraphQL/native-hook inventory facts without inventing GraphQL documents or tenant schemas. Commit only the Batch 5 source directories after its assertions and connector generation gates pass.

### Refactor / verification

Review generated JSON for deterministic ordering and no source data invented beyond the already-normalized provider inventory. Run the required generated-file, boundary, focused test, and source-map assertions; document any check the local per-command limit cannot complete.

## Lifecycle Record

- Required lifecycle resolved and generated inline: `scripts/gsd prompt discuss-phase 4290 --auto`; `plan-phase 4290 --tdd --skip-research --auto`; `execute-phase 4290 --interactive --auto`; `verify-work 4290 --auto`; `code-review 4290 --auto`.
- Inline/manual fallback: this environment is not a compatible Pi isolated-agent runtime, the task requires a single autonomous worker, and the canonical contract forbids role spawning. The GSD decisions, plan, red/green ledger, verification, and review are therefore recorded in this phase directory.
- Skills loaded: `github-issue-first-delivery`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.
- CLI help/manual/website parity: not applicable. This phase adds source-map evidence only; it does not change a CLI command, flag, help topic, docs surface, or generated manual.

## Commit and Push Checkpoints

1. Planning evidence checkpoint.
2. Batch 4 source locks/maps and green assertions.
3. Batch 5 source locks/maps and green assertions.
4. Verification/review evidence and any review-fix checkpoint.

