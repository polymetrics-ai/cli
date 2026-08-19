## Task Delivery Header

- Issue: Refs #4290 — chore(connectors): map parity batches 4 and 5
- Base branch: fm/cli-reverse-etl-destination-r1
- Merges into: fm/cli-reverse-etl-destination-r1 → main
- Delivery: PR #4295 targets `fm/cli-reverse-etl-destination-r1`, with source-driven inventories, rematerialized parity ledgers, and a 20-row seven-surface report. Application-level generic reverse-ETL dispatch remains an explicit foundation dependency until the final foundation head is merged and exercised through the installed App/CLI path.
- Working branch: fm/cli-map-batch45-r1
- Task: Produce source-locked, complete operation disposition maps for Batch 4 (`salesforce`, `hubspot`, `pipedrive`, `mailchimp`, `zendesk-support`, `quickbooks`, `bamboo-hr`, `airtable`, `google-analytics-data-api`, `woocommerce`) and Batch 5 (`pinterest`, `tiktok-marketing`, `linear`, `buildkite`, `sonar-cloud`, `launchdarkly`, `fastly`, `squarespace`, `ebay-fulfillment`, `shipstation`). A pre-existing `api_surface.json` is not an inventory boundary: validate each against the provider's complete public source, regenerate every understated surface, then map every resulting operation exactly once in the proven batch-1 row shape.
- Verification: Source-lock/map inventory assertions; `go run ./cmd/connectorgen validate`; `go run ./cmd/connectorgen surface-sync --check`; the connector-boundary check; `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/engine`; and generated/docs checks that apply to definitions-only changes.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every Batch 4 documented endpoint has one map row | live | The rebuilt method/path multiset in each Batch 4 `api_surface.json` must match its provider-source lock and disposition map exactly; removing or retaining an unsupported row makes the assertion fail. |
| Every Batch 5 documented endpoint has one map row | live | The same source-to-surface-to-map assertion compares each Batch 5 bundle; it detects duplicate and missing rows. |
| The provider source is complete enough to support its count | live | Every lock records `counts.total`, method/protocol counts, `operations_found`, and coverage confidence/basis. Machine-readable artifacts are SHA-256/byte pinned; rendered references record the complete crawl; genuinely dynamic or inaccessible surfaces retain explicit `null` totals and their basis. |
| Row vocabulary distinguishes unauthored work from engine incapability | live | The assertion rejects any endpoint-level `foundation-gap` without engine file/line evidence and minimal change, checks unauthored rows use `declaration-pending`, and verifies that reverse-ETL is only an eligibility attribute carrying the verified generic typed-destination limitation. |
| Existing connector definitions remain valid and synchronized | live | `connectorgen validate` and `surface-sync --check` parse the real affected bundles and report no drift. |
| Every assigned connector has an honest seven-surface status | live | The 20-row machine-readable ledger is regenerated from each bundle, source map, and transport declaration; its count and connector set checks fail on a missing or duplicate row. A zero classification count is explicitly not an N/A claim. |
| Every enumerable provider operation has fail-closed pre-merge evidence | live, blocked pending evidence | `HARD-PREMERGE-GATE.json` joins each source-record ID to exactly one canonical map row and records exact source URL/document hash/bytes/version status, runtime reachability, CLI/website/fixture evidence, control metadata, and separate six-surface dispositions. It rejects duplicate/missing joins and a disabled row unless its source says provider-deprecated/absent; unenumerable dynamic/unavailable inventories remain explicit blockers, never silent omissions. |
| Open shared foundation gaps are visible, deduplicated, and non-merge-ready | live, blocked pending #4304 | `FOUNDATION-GAP-LEDGER.json` has one stable gap definition per shared capability plus one exact provider-operation row for every affected action. It reports 345 enumerated provider operations and 49 retained compatibility bindings across 8 connectors, batch and portfolio rollups, owner #4304, runtime evidence, and deterministic closure commands. Each affected reverse-ETL surface has `enabled_on_affected_surfaces: false` and `merge_ready: false`; it is never disabled or N/A. |
| No credentialed provider operation is performed | fake | This is a documentation-only mapping lane; locks are public descriptions retrieved without credentials, while all provider behavior remains unexercised. |

## Scope and Ownership Guard

- Allowed production paths are only `internal/connectors/defs/{salesforce,hubspot,pipedrive,mailchimp,zendesk-support,quickbooks,bamboo-hr,airtable,google-analytics-data-api,woocommerce,pinterest,tiktok-marketing,linear,buildkite,sonar-cloud,launchdarkly,fastly,squarespace,ebay-fulfillment,shipstation}/{api_surface.json,sources/}`.
- This phase also owns its required `.planning/phases/issue-4290-map-parity-batches-45-r1/` evidence. No engine, schema, source-operation contract, CLI surface, or other connector files may change.
- The issue body has three trailing connector names after the explicit Batch 5 ten. The launch brief's exact twenty-connector list is authoritative; `gocardless`, `amazon-seller-partner`, and `miro` are deliberately not touched.
- `bamboohr` and `sonarcloud` resolve to the repository-owned `bamboo-hr` and `sonar-cloud` definition directories.

## Locked Disposition Rules

1. `foundation-gap` means the current engine refuses the documented shape. It must name the refusal file and line plus the smallest viable engine change. Missing operation contracts, commands, or CLI surfaces are `declaration-pending`, never foundation gaps.
2. `requires-elevated-scope` is enabled with source/runtime scope metadata; it is never a rejection or disabled state.
3. Every DELETE is represented. A delete requested by this issue is not `unsafe-to-exercise` merely because it mutates data.
4. `docs/sync-transport-definition.md` is the transport contract. ETL is connector-neutral where a bundle can truthfully supply the declarative source contract; absence of a connector-owned `sync_transport.json` is `declaration-pending`, never the retired #4093 gap. A typed write action retains the canonical `direct_write` endpoint classification, including DELETE. At #4304 foundation SHA `c6f03c937`, a generic destination DefinitionFactory exists, but `internal/app/transport_dispatch.go:40-74` still selects only the legacy issue-label/warehouse routes. Reverse-ETL eligibility therefore records `application-generic-destination-dispatch` until the foundation's App/CLI dispatch integration lands and is exercised. An affected action is `foundation-gap`, not enabled for reverse ETL, and never merge-ready until that closure; it is neither disabled nor N/A. Its minimal change belongs to #4304: admit any definition-preflighted generic destination through the real App dispatch path; do not create a connector workaround. Every typed write action has an explicit eligibility disposition; single-action proof cannot settle connector completion when other representable record-driven actions remain.
5. Main now includes PR #4297. `head-response-less-operation-executor` and `operation-scoped-rest-pagination` are closed runtime capabilities, so a fresh materialization must not preserve either as a foundation gap; an operation that otherwise has a complete connector-owned declaration is enabled.
6. Merge is frozen until #4304 completes the provider-neutral App/CLI generic-destination dispatch integration across the portfolio. Until then, retain `application-generic-destination-dispatch` solely as a nested reverse-ETL eligibility gap on each connector-owned typed direct write, and project it into `FOUNDATION-GAP-LEDGER.json` as a non-enabled affected surface. Do not fabricate a `transport_binding` or author a typed action without a provider-declared request contract; all other mutations stay identified as `direct_write` and `declaration-pending` until they have one.
7. Source maps reproduce the batch-1 shape: `method`, `path`, `parity_class`, `api_surface`, `source`, `state`, `foundation`, `rejection`, and `declaration`. The regenerated, provider-source-derived `api_surface.json` inventory is the connector-local crosswalk; each lock records its public documentation source, provider total, confidence, and exact endpoint inventory. Never report `declared_percent`.
8. No provider credential, tenant discovery, API request, or write is used. Salesforce tenant-defined object/field dependencies remain named runtime metadata, not discovered facts.
9. The captain pre-merge gate is fail-closed. For every enumerable provider operation, preserve the locked source URL/document revision status/hash, canonical mapping, runtime reachability, generated CLI and website status, executable fixture/conformance status, controls, and separate ETL/reverse-ETL/direct-read/direct-write/binary-download/binary-upload reconciliation. `pending`, `not-asserted`, and an unmaterialized source version are never N/A. N/A is available only when the provider source itself proves the capability absent. Scope, tier, destructive, risk, and safety policy are typed runtime metadata/confirmation, never a reason to remove command reachability.

## TDD Delivery Plan

### Red — source-bounded inventory defect

The earlier source lock only compared a map with the pre-existing `api_surface.json`, which hid provider operations not already in the bundle. This is a failing acceptance condition: a well-known provider's old count must not be accepted as a complete count merely because the source map agrees with it. Rebuild every retrievable source from its provider artifact/reference and compare the rebuilt surface with the exact artifact operations.

### Green — source-derived API surfaces

For every connector with a complete machine-readable source, pin the exact bytes and digest, derive its unique request operations, and materialize `api_surface.json` version-2 provenance. Preserve an executable stream/write binding only when the source method/path is identical; record all other provider operations as blocked operation-ledger rows without fabricating body, response, pagination, or schema contracts. For rendered sources, crawl the complete reference before producing the same inventory. For Salesforce's instance-dependent object surface and unavailable browser-only sources, retain an explicit dynamic/unavailable basis and an unknown total rather than a fabricated count.

### Red — complete-map invariants

Before materialization, every target lacked both `sources/<connector>-operation-source-lock.json` and `sources/<connector>-declaration-disposition.json`; an inventory assertion therefore has no source lock/map to compare and fails. The test first requires source-lock/map presence, exact method/path multiset equality with `api_surface.json`, valid six-class values, DELETE coverage, and vocabulary correctness.

### Green — Batch 4

Rematerialize the ten Batch 4 locks/maps from the regenerated provider inventories. Preserve only provider-matching `covered_by` bindings; classify streams as `etl`, typed write bindings as enabled `direct_write`, other bounded operation bindings as direct read/write, and documented binary endpoints as binary read/write. Assess ETL against the connector-neutral source contract; a missing descriptor/conformance claim is pending. Record reverse-ETL eligibility beneath every typed direct-write action, with the verified generic typed-destination foundation gap and never a fabricated action binding. Commit the regenerated Batch 4 surfaces and source evidence after assertions and connector generation gates pass.

Mailchimp source exception resolved: its public Swagger root was available to the ordinary client but its 181 `$ref` path documents returned Akamai 503 after serialized retry. Chrome successfully retrieved all 181 documents, recording a per-document UTF-8 byte count and SHA-256. Their current unique physical request inventory is 295, not the inherited 298-row bundle count. The materialized surface has 323 rows because its 28 historical executor bindings are retained as explicit local compatibility rows; the provider total remains 295.

### Green — Batch 5

Repeat the same source-lock and full-inventory materialization for Batch 5, including GraphQL/native-hook inventory facts without inventing GraphQL documents or tenant schemas. TikTok Marketing and eBay Fulfillment are browser-source failures in this environment (TikTok: `ERR_SSL_PROTOCOL_ERROR`; eBay: official error page/403), so their locks explicitly record `skipped: no-public-api-description` with exact browser evidence and unknown totals. Commit the regenerated Batch 5 surfaces and source evidence after its assertions and connector generation gates pass.

### Refactor / verification

Review generated JSON for deterministic ordering and no source data invented beyond the retrieved provider inventory. Per connector report the old API-surface count, regenerated count, and source basis. Run the required generated-file, boundary, focused test, and source-map assertions; document any check the local per-command limit cannot complete.

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
