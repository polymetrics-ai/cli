# Discussion Log — Issue #3975: committed-transaction staging and durable receipts

> Audit trail only. Decisions are captured in `CONTEXT.md`; this log preserves
> the alternatives considered by the required GSD discussion step.

**Date:** 2026-08-11
**Mode:** `scripts/gsd prompt discuss-phase 3975 --auto` with documented
manual inline fallback because #3975 is an issue-specific rather than numbered
roadmap phase.
**Areas discussed:** durability boundary, recovery/cleanup, failure limits,
test evidence.

## Durable-stage meaning

| Option | Description | Selected |
|---|---|:---:|
| Treat a fsynced stage as acknowledgement | Would advance source progress when only private storage exists. | |
| Receipt after committed downstream materialization | Stage is recoverable only; a durable whole-transaction receipt gates eligibility. | ✓ |

**Auto-selected decision:** receipt after committed downstream materialization,
because the accepted CDC design and fleet learning prohibit acknowledgement of
an in-progress or merely locally staged transaction.

## Recovery disposition

| Option | Description | Selected |
|---|---|:---:|
| Publish incomplete stages after restart | Would expose uncommitted source data. | |
| Purge incomplete/orphan stages; retain sealed no-receipt work for explicit resume | Preserves transaction boundaries and at-least-once replay semantics. | ✓ |

**Auto-selected decision:** purge pre-commit/orphan files, retain sealed
committed work without a receipt, and clean sealed residue only after validated
receipt persistence.

## Quota and failure outcome

| Option | Description | Selected |
|---|---|:---:|
| Grow memory/disk or fall back to a cursor | Hides a capacity failure and changes change-capture semantics. | |
| Typed, fail-closed finite limits | Returns `TransactionStageLimitExceeded`, publishes and acknowledges nothing. | ✓ |

**Auto-selected decision:** typed finite byte/record/time limits with no
polling/cursor fallback.

## Test evidence

| Option | Description | Selected |
|---|---|:---:|
| Assert only returned errors | Misses silent premature publication and leftover crash files. | |
| Assert artifacts, receipt durability, ordered visibility, eligibility, and cleanup | Tests the facts the stage claims. | ✓ |

**Auto-selected decision:** use isolated directories plus injected disk,
fsync, rename, receipt, cancellation, and crash-boundary failures; first RED is
`TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt`.

## Deferred Ideas

- No PostgreSQL decoder, source feedback, polling fallback, target DML,
  capability flip, or live database test is folded into this shared foundation.
