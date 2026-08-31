# Issue 4283 — Stripe retained-source mapping and legacy app witnesses (R1)

## Delivery boundary

- Base: `origin/fm/cli-top100-declaration-batch-r1` at
  `39be11d322bc523e6a044a7213ea636d3e240ca9`.
- Delivery: candidate branch only; no parent integration or merge is included.
- Scope: Stripe definition artifacts, Stripe-focused tests, and this trace.
  Source locks, shared runtime/engine, Atlas, CI, and other connectors remain
  untouched.

## Retained-source accounting

The frozen Stripe REST lock retains 589 operations (263 GET, 294 POST, 32
DELETE), or 4,123 exact seven-lane cells. The source lane matrix is strictly
mapping-only:

| Lane | mapped-unproven | missing-foundation | not-applicable |
| --- | ---: | ---: | ---: |
| Direct read | 263 | 0 | 326 |
| Direct write | 326 | 0 | 263 |
| Binary download | 1 | 0 | 588 |
| Binary upload | 1 | 0 | 588 |
| ETL | 128 | 0 | 461 |
| Reverse ETL | 326 | 0 | 263 |
| Sync transport | 0 | 1 | 588 |

Totals are 1,045 mapped-unproven, one missing-foundation, and 3,077
not-applicable cells. There are no `implemented` cells and no matrix
`execution` objects.

The sole missing foundation remains
`stripe.rest.PostWebhookEndpoints` / `stripe.inbound-webhook-receiver.v1`.
Its ledger explicitly records no inbound receiver, selected executor, or
runnable sync transport. Direct and binary cells remain mapped-unproven.

## Legacy behavior witnesses (not source-bound execution evidence)

Existing local Stripe declarations expose five ETL streams—customers, charges,
invoices, subscriptions, and products—and three customer reverse actions—create,
update, and delete. Their local fake-server/DuckDB tests exercise the existing
App behavior:

- every declared stream reads two cursor pages into temporary DuckDB;
- a fixture 502 stops materialization before warehouse rows are committed;
- every declared customer action follows plan → preview → approval → execute;
  delete requires typed destructive confirmation before the fake provider sees a
  request.

These witnesses are deliberately limited to legacy local behavior. The Stripe
declarations do not carry a source descriptor contract, so no public command
contains `source_operation`, no matrix cell is promoted, and the witnesses do
not establish source-bound CLI execution.

The public CLI regression dynamically enumerates all eight current ETL/reverse
commands. With an intentionally invalid `base_url` and no credential, each must
reach the ordinary missing-credential boundary with zero provider I/O; the test
rejects descriptor-specific or source-preflight error output. That proves only
that the legacy declarations are not accidentally routed through a descriptor
preflight path.

## Remediation record

The earlier candidate incorrectly added eight `source_operation` fields and
promoted five ETL plus three reverse-ETL matrix cells. This follow-up removes
exactly those fields, returns all eight source cells to their cited
`mapped_unproven` state, and restores the original 589-operation / seven-lane
denominators. No source lock, runtime, or sync transport was changed.

## Focused validation

Commands are run only with the task-isolated cache
`/private/tmp/cli-4283-stripe-source-materialization-r1/.gocache-stripe`:

```text
go test ./internal/connectors/defs/stripe -run '^(TestStripeSourceLaneMatrix|TestStripeMissingFoundation)' -count=1
go test ./internal/cli -run '^TestStripe(Declared|Public)' -count=1
go test -race ./internal/cli -run '^TestStripe(Declared|Public)' -count=1
go vet ./internal/connectors/defs/stripe ./internal/cli
```

The final JSON and scope checks additionally verify valid JSON, zero descriptor
fields in `cli_surface.json`, zero matrix execution claims, the retained matrix
totals above, the unchanged missing-foundation ledger, and the intended
connector-local diff.
