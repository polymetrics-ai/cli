# Issue 4283 — Stripe source-backed lane materialization (R1)

## Task delivery header

- Issue: Refs #4283 — Batch R1 Stripe source-backed lane materialization.
- Base: `origin/fm/cli-top100-declaration-batch-r1` at
  `39be11d322bc523e6a044a7213ea636d3e240ca9`.
- Merge direction: `fm/cli-top100-declaration-batch-r1` → `main`.
- Delivery: a committed/pushed candidate for independent review only; this work
  does not integrate or merge anything.
- Owner scope: `internal/connectors/defs/stripe/**`, focused Stripe test files,
  and this evidence trace. Source locks, shared runtime/engine, Atlas, CI, and
  other connectors are out of scope.

## Source and lane inventory

The frozen Stripe REST lock has 589 operations (263 GET, 294 POST, 32 DELETE),
therefore 4,123 exact seven-lane cells. Before this change its matrix reports
1,045 mapped-unproven cells, one missing-foundation cell, and 3,077
not-applicable cells; it claims no implemented cells.

Existing declaration chains, derived from the lock-backed matrix links rather
than a hand-maintained aggregate, are:

| Lane | Existing source-bound declaration cohort | Intended disposition |
| --- | --- | --- |
| ETL | `streams.json`: customers, charges, invoices, subscriptions, products; matching `cli_surface.json` ETL commands | implement only after App/DuckDB witness succeeds |
| Reverse ETL | `writes.json`: create_customer, update_customer, delete_customer; matching reverse-ETL commands | implement only after plan → preview → approval → execute witness succeeds |
| Direct read/write | The same source routes have no declaration-owned direct command | remain mapped_unproven |
| Binary download/upload | PDF quote GET and multipart file POST lack a bounded source-bound binary declaration | remain mapped_unproven |
| Sync transport | `stripe.rest.PostWebhookEndpoints` has source webhook-registration evidence, but Atlas `stripe.inbound-webhook-receiver.v1` is planned only | retain exactly one missing-foundation record; do not create a transport |

The 123 remaining pageable ETL cells and 323 remaining mutation reverse-ETL
cells retain their source mappings as `mapped_unproven`: the common runtime is
available, but each lacks a field-complete Stripe stream/write declaration,
schema, and command chain. That is a declaration gap, not a request to invent
or broaden an executor.

## GSD and TDD record

`scripts/gsd doctor`, `sources`, and the prompts for discuss, plan (`--tdd`),
execute, verify-work, and code-review were consulted. Their configured manual
fallback applies to this bounded connector artifact task. The executable plan
is therefore recorded here:

1. Add exact source-operation bindings to the eight already implemented Stripe
   CLI declarations and reflect those bindings in the source lane matrix.
2. Promote only the five ETL and three reverse-ETL cells whose existing
   declaration chains close over a source ID, record, command, and focused
   witness test.
3. Add a connector-local, non-executing missing-foundation ledger for the
   already-present Stripe inbound webhook receiver gap.
4. Add local `httptest` plus temporary DuckDB proof: multipage ETL materializes
   each implemented source-bound stream; non-success does not materialize;
   reverse ETL proves plan/preview/approval/execute and the delete confirmation
   boundary.
5. Run only focused normal/race tests, Stripe JSON/contract checks, focused vet,
   and a scope diff when capacity is authorized. No provider credential or live
   account claim is part of this work.

### Red / green ledger

- Red: the first focused definition run exposed a validator-order regression:
  a removed source row was reported through an API-surface backlink before the
  required source-row-absence diagnostic. The same focused runtime witness
  exposed two harness assumptions: non-destructive reverse plans retain their
  token from planning (rather than preview), and parameterized source routes
  must match a concrete path at execution. Both were corrected in test and
  contract-validation code only; no product runtime code changed.
- Green normal:
  `GOCACHE=/private/tmp/cli-4283-stripe-source-materialization-r1/.gocache-stripe go test ./internal/connectors/defs/stripe -run '^(TestStripeSourceLaneMatrix|TestStripeMissingFoundation)' -count=1`
  and
  `GOCACHE=/private/tmp/cli-4283-stripe-source-materialization-r1/.gocache-stripe go test ./internal/cli -run '^TestStripeSourceBound' -count=1`.
- Green race:
  `GOCACHE=/private/tmp/cli-4283-stripe-source-materialization-r1/.gocache-stripe go test -race ./internal/cli -run '^TestStripeSourceBound' -count=1`.
- Green vet:
  `GOCACHE=/private/tmp/cli-4283-stripe-source-materialization-r1/.gocache-stripe go vet ./internal/connectors/defs/stripe ./internal/cli`.
- CLI help/documentation parity: no command, flag, help string, or user-facing
  documentation changes are planned. `source_operation` is declaration
  provenance only; no help regeneration is needed.

## Acceptance checks

- Exact source IDs come from the retained lock and matrix backlinks.
- The matrix must reconcile every one of the 589 source operations and all
  seven lanes, with no execution promotion outside the eight declaration chains.
- The missing-foundation ledger must name only
  `stripe.rest.PostWebhookEndpoints`, state the planned Atlas capability, and
  explicitly disclaim a receiver, executor, or runnable sync transport.
- Tests must observe actual app behavior with a local fake server and temporary
  DuckDB; assertion-only metadata tests are insufficient for an implementation
  claim.
