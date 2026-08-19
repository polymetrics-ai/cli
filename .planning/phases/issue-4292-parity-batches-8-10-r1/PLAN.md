# Issue #4292 — complete six-class parity map for batches 8–10

## Reconciliation supersession — 2026-08-20

This section supersedes every conflicting delivery, base, scope, and
reverse-ETL statement below. It records the captain's authoritative
reconciliation directive and is the current execution contract.

## Task Delivery Header

- Issue: Refs #4292 — chore(connectors): map parity batches 8, 9 and 10
- Base branch: `fm/cli-reverse-etl-destination-r1`
- Merges into: `fm/cli-reverse-etl-destination-r1` → `main`
- Delivery: Existing PR #4301 is retargeted to the foundation branch and has
  the complete seven-surface ledger, connector-local declarations, focused
  tests, and required local gates recorded before its final push. The PR base
  is read back through the GitHub API after every retarget.
- Working branch: `fm/cli-map-batch8910-r1`
- Task: Reconcile the thirty listed connector bundles across `binary_read`,
  `binary_write`, `direct_read`, `direct_write`, `etl`, `reverse_etl`, and
  executable CLI-command surfaces. Keep every provider-sourced operation
  reachable through its faithfully representable connector-owned contract;
  do not hide privileged, destructive, or unusual operations. Generate the
  required 30-row machine-readable ledger and human summary. Reverse-ETL
  declarations are connector-owned and may only select the exact generic typed
  destination executor, typed action, input fields, acknowledgement, delivery
  facts, and conformance evidence that the merged foundation admits.
- Verification: Red/green source-map tests; targeted generator, bundle, and
  transport declaration checks per increment; `connectorgen validate`,
  `surface-sync --check`, generated checks, `connector-boundary`, scoped Go
  tests, and the repository's individual `make verify` gates. Before the final
  push, fetch and merge the latest foundation head, prove it is an ancestor,
  then exercise the installed App/CLI route without credentials.

### Reconciliation constraints

- Work is confined to the assigned `internal/connectors/defs/<connector>/`
  directories and this phase's evidence. No generic HTTP writer, engine edit,
  or App/CLI workaround is allowed.
- #4304's current head exposes the generic declarative typed destination to
  the transport composition layer but has not yet landed persisted App/CLI
  dispatch integration. Connector declarations may be authored and statically
  validated, but this branch must not claim application-level reverse ETL is
  deployable until the updated foundation head is merged and the installed App
  route is exercised.
- Safety, scopes, destructive confirmation, and certification limits govern
  execution; they never remove a faithfully representable command from the
  provider-sourced reachability inventory.
- Any operation that lacks a provider-sourced typed request contract remains
  explicitly `declaration-pending`; `unsupported` is reserved for a concrete
  refusing engine file/line and minimal hook. No request, response,
  pagination, or body schema is inferred.
- Commits cover at most five connector bundles plus their shared phase evidence
  and each message carries `Refs #4292`.

### Reconciliation TDD slices

1. **Red — current verifier:** retain the failed CI and local
   `TestLeverHiringAPISurfaceOperationLedger` evidence before changing the
   generator. The failure demonstrated that source regeneration lost both its
   named dependency metadata and Lever's provider-formatted EEO route.
2. **Green — source-map repair:** restore the named connector-local dependency,
   filter the rendered prose-only `/profile_forms` false positive, preserve the
   documented no-leading-slash EEO route, regenerate only Lever's source map,
   and pass the focused ledger test plus Batch 9 integrity check.
3. **Red/green — seven-surface declarations:** a deterministic ledger checker
   must fail if a connector declaration names an absent stream/action, leaves a
   destination mapping without exact `input_fields`, claims an application
   dispatch route before its foundation evidence, or drops a provider operation
   from the reachability count. Generate declarations only from existing typed
   bundle contracts and connector schemas; pass static bundle/transport loading
   and focused no-I/O transport preflight checks per five-connector increment.
4. **Foundation handoff:** fetch and merge the latest #4304 head before final
   verification. Prove that SHA is an ancestor and run the installed App/CLI
   path. Until that passes, record application deployment as pending rather
   than adding a local fallback.

## Task Delivery Header

- Issue: Refs #4292 — chore(connectors): map parity batches 8, 9 and 10
- Base branch: main
- Merges into: main
- Delivery: A direct PR against `main`, containing three committed,
  source-locked ten-connector declaration-map increments, with local
  validation green and no credentialed provider access. Publication is held
  until the source-lock completeness correction below is green.
- Working branch: fm/cli-map-batch8910-r1
- Task: Add a public-source lock, source/API-surface crosswalk, complete
  six-class disposition ledger, and per-connector summary for brex, zoho-books,
  testrail, amplitude, posthog, metabase, dbt-cloud, looker, mode, dremio,
  coda, clickup, calendly, greenhouse, lever, ashby, workable, recruitee,
  hibob, factorial, datadog, pagerduty, auth0, okta, firehydrant,
  adobe-commerce, commercetools, recharge, docuseal, and eventbrite. Every
  documented DELETE must be represented or explicitly disabled without
  inventing a request, response, pagination, or body schema.
- Verification: Run the planned red/green artifact-integrity assertions;
  `go run ./cmd/connectorgen validate` for every changed bundle;
  `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`;
  the changed-package tests and generated-file checks; and
  `go run ./cmd/connectorgen boundary . --json` in a detached capture as the
  brief requires. Record each exact result in `VERIFICATION.md`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every connector has a reproducible public-source lock | live | The lock records URL, SHA-256, byte count, capture time, method counts, and every source operation; its retained source bytes hash to the recorded value. |
| Every documented operation has one map row and parity class | live | A JSON assertion compares unique source IDs from each lock, crosswalk, and ledger; all counts and six-class totals agree exactly. |
| Missing declaration work is not misreported as an engine gap | live | Each disabled row names `declaration-pending` unless it carries concrete engine file/line evidence; direct-write rows carry the locked generic-destination gap only in their reverse-ETL eligibility attribute. |
| Generated bundle metadata remains valid and synchronized | live | `connectorgen validate` accepts each changed bundle and `surface-sync --check` reports no drift. |
| The provider surface is complete enough to make a map claim | live | Every source lock has `counts.total` and per-kind counts from a complete provider OpenAPI/Swagger/Discovery document or complete rendered reference; every ledger reports `operations_found` plus `coverage_confidence` and its basis, never a self-referential percentage. |

## Scope guard

This is the multi-connector declaration-map outcome explicitly assigned by
#4292, following the batch-one declaration-map precedent. It changes the listed
`internal/connectors/defs/<connector>/sources/` directories and regenerates
their `api_surface.json` inventories from the pinned provider operations, plus
this phase evidence. It does not implement a connector, alter shared
runtime/tooling, or make a certification claim. No source document, operation
field, CLI command, or transport binding is guessed or synthesized.

## Foundation classification

| Class | Mapping rule |
| --- | --- |
| Direct read | Existing runnable `rest_read` binding is enabled; a missing local contract/command is `declaration-pending`. |
| Direct write | An existing typed write action, including documented DELETE, is enabled; missing connector-local work is `declaration-pending`. Its `reverse_etl_eligibility` attribute is independently assessed and does not alter this primary class. |
| ETL | Enabled only for existing stream/definition-owned declarative source transport evidence; otherwise retain the exact local declaration state. |
| Reverse ETL | Not a primary endpoint class. Every `direct_write` row records a `reverse_etl_eligibility` attribute: currently `foundation-gap` `generic-typed-destination-executor`, evidence `internal/app/issue_label_warehouse_transport.go:85-95`, with the locked minimal change from `CONTEXT.md`. |
| Binary read/write | Enabled only for a bounded source-backed binary contract; otherwise disabled as `declaration-pending` or a source-backed schema/media limitation. |

## Reverse-ETL freeze preparation — 2026-08-19

Issue #4303 is now the estate-wide merge gate: it must add the
connector-neutral typed destination factory before any connector can truthfully
declare reverse ETL. This issue does **not** predeclare a destination or invent
a `transport_binding` while that capability is absent.

1. Generate a readiness audit and machine-readable typed-action inventory from
   all 30 source-first ledgers. For every mapped provider write, distinguish a
   row already bound to a named typed `writes.json` action from a row that still
   needs connector-local typed-action authoring. Preserve exact source IDs,
   routes, locations, and action IDs for the action-backed rows. Treat that
   distinction as preparation evidence, not a destination declaration or a
   product-safety decision.
2. Retain the current `generic-typed-destination-executor` eligibility gap on
   every direct-write row. After #4303 lands, only action-backed rows that also
   receive connector-owned evidence, explicit source-field bindings,
   acknowledgement, and per-sync-mode apply strategies may become destinations.
3. Track Zoom separately as captain-directed critical-path preparation: its
   204-action destination cohort is outside the #4292 batch source inventory
   and is recorded as a decision-provided target, not as a source-derived
   operation count. No Zoom destination declaration is made in this PR.

## TDD implementation slices

### Red — complete-map assertion before each batch

1. Assert that all ten target disposition ledgers, locks, and crosswalks exist.
   It must fail before production artifacts are added.
2. Assert that every source operation is represented exactly once, that all
   class totals sum to the source total, that typed write actions are classified
   `direct_write`, and that each direct-write row retains the locked
   generic-destination gap in its reverse-ETL eligibility attribute. It must
   fail before the batch map is complete.

### Green — source-backed mapping

1. Pin each public provider description without credentials and retain the
   source-derived operation inventory in its connector source lock.
2. Crosswalk each documented operation to `api_surface.json` without creating
   an operation contract, command, schema, pagination, or body not in the
   source.
3. Write the complete disposition ledger and summary. Use the batch-one JSON
   shape, preserve an exact source location, and classify every documented
   DELETE.
4. Run the red assertions until they pass, then validate all changed bundles
   and generated metadata.

### Red/green — reverse-ETL readiness freeze

1. Red: an audit that counts a primary `reverse_etl` endpoint, treats a typed
   direct-write action as an already-declared destination, or contains a
   `transport_binding` is false.
2. Green: generate an all-30 table from the ledgers, showing source write count,
   typed-action-backed count, connector-local typed-action-pending count, and
   the uniform #4303 foundation gate. Record unavailable/dynamic source states
   as non-enumerable rather than zero-operation claims.

### Refactor — review and publication

1. Review the changed source ledgers for duplicated/missing IDs, incorrect
   class totals, invented mappings, stale transport claims, secret leakage, and
   false foundation gaps.
2. Keep only source artifacts and phase evidence in the diff. Commit after
   batches 8, 9, and 10; rebase onto current `main` immediately before each
   push, never force-push.

## Source-lock completeness correction — 2026-08-19

The captain's `SOURCE-LOCK-DEFECT.md` invalidated the initial maps' input
model: a documentation landing page and a count derived from
`api_surface.json` cannot prove a provider's documented operation surface.
The initial Batch 8 and 9 commits remain on the branch for auditability but
their generated source maps are not eligible for a PR until replaced.

1. Retrieve each connector's complete provider-published OpenAPI, Swagger,
   Discovery/service-model document. Where none exists, retrieve the complete
   official rendered reference, explicitly record that representation, and
   retain the exact source-document or source-bundle bytes/hashes. A landing
   page or index alone is rejected.
2. Extract the operation inventory from that document, including its source
   location, then map every extracted operation. Existing `api_surface.json`
   bindings may enable a row only when they exactly match a pinned source
   operation; they are never the source inventory.
3. Require `rest.counts.total`, per-method counts, `operations_found`, and a
   `coverage_confidence` of `machine-readable` or
   `complete-rendered-reference`, with its concrete evidence. A genuinely
   unavailable public reference or instance-dependent surface records an
   explicit `null` total and reason instead of fabricating a count. Delete
   `declared_percent`.
4. Reject a source whose count is implausibly small for its documented API
   unless the lock records an explicit provider-scope explanation and a
   non-`machine-readable` confidence. TestRail and Eventbrite retain a source
   skip only when no official public reference can actually be retrieved.

## Batch checkpoints

1. Batch 8: brex, zoho-books, testrail, amplitude, posthog, metabase,
   dbt-cloud, looker, mode, dremio.
2. Batch 9: coda, clickup, calendly, greenhouse, lever, ashby, workable,
   recruitee, hibob, factorial.
3. Batch 10: datadog, pagerduty, auth0, okta, firehydrant, adobe-commerce,
   commercetools, recharge, docuseal, eventbrite.

## Lifecycle record

- Inline/manual GSD fallback: this worker has no compatible interactive Pi
  runtime and the canonical delivery contract forbids role spawning. Resolved
  and executed inline: `scripts/gsd prompt discuss-phase 4292`,
  `plan-phase 4292 --tdd`, `execute-phase 4292`, `verify-work 4292`, and
  `code-review 4292`.
- Skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, and `golang-testing`.
- CLI help/manual/website parity: not applicable; this plan forbids CLI,
  API-surface, command, generated-help, and website changes.
