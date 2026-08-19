# Issue #4292 — complete six-class parity map for batches 8–10

## Task Delivery Header

- Issue: Refs #4292 — chore(connectors): map parity batches 8, 9 and 10
- Base branch: main
- Merges into: main
- Delivery: A direct PR against `main`, containing three committed,
  source-locked ten-connector declaration-map increments, with local
  validation green and no credentialed provider access.
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
| Missing declaration work is not misreported as an engine gap | live | Each disabled row names `declaration-pending` unless it carries concrete engine file/line evidence; the reverse-ETL rows all use the locked generic-destination gap. |
| Generated bundle metadata remains valid and synchronized | live | `connectorgen validate` accepts each changed bundle and `surface-sync --check` reports no drift. |

## Scope guard

This is the multi-connector declaration-map outcome explicitly assigned by
#4292, following the batch-one declaration-map precedent. It changes only the
listed `internal/connectors/defs/<connector>/sources/` directories plus this
phase evidence. It does not implement a connector, alter shared runtime/tooling,
or make a certification claim. No source document, operation field, CLI command,
or transport binding is guessed or synthesized.

## Foundation classification

| Class | Mapping rule |
| --- | --- |
| Direct read | Existing runnable `rest_read` binding is enabled; a missing local contract/command is `declaration-pending`. |
| Direct write | Existing typed `rest_write` binding, including documented DELETE, is enabled; missing connector-local work is `declaration-pending`. |
| ETL | Enabled only for existing stream/definition-owned declarative source transport evidence; otherwise retain the exact local declaration state. |
| Reverse ETL | `foundation-gap` `generic-typed-destination-executor`, evidence `internal/app/issue_label_warehouse_transport.go:85-95`, with the locked minimal change from `CONTEXT.md`. |
| Binary read/write | Enabled only for a bounded source-backed binary contract; otherwise disabled as `declaration-pending` or a source-backed schema/media limitation. |

## TDD implementation slices

### Red — complete-map assertion before each batch

1. Assert that all ten target disposition ledgers, locks, and crosswalks exist.
   It must fail before production artifacts are added.
2. Assert that every source operation is represented exactly once, that all
   class totals sum to the source total, and that reverse-ETL rows retain the
   locked generic-destination foundation gap. It must fail before the batch
   map is complete.

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

### Refactor — review and publication

1. Review the changed source ledgers for duplicated/missing IDs, incorrect
   class totals, invented mappings, stale transport claims, secret leakage, and
   false foundation gaps.
2. Keep only source artifacts and phase evidence in the diff. Commit after
   batches 8, 9, and 10; rebase onto current `main` immediately before each
   push, never force-push.

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
