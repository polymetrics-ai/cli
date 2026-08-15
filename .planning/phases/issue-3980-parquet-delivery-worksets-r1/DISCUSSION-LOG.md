# Discussion log — Issue 3980

## Mode

`discuss-phase 3980 --auto` was executed inline from the task brief and live
issue because this autonomous lane must not wait for interactive answers.

## Resolved questions

| Question | Decision | Basis |
| --- | --- | --- |
| What keys a destination workset? | `ManagedTargetDeliveryLedgerKey` derived from the asserted control record. | #3981 is the durable identity authority and explicitly excludes mutable source artifact/display/table names. |
| Which package owns the type? | `internal/connectors/database`. | The type consumes the database control record and ledger key; putting it in `warehouse` would introduce an import cycle because the database foundation already depends on warehouse. |
| What does phase F5 persist? | Immutable complete projection, delta, tombstone, candidate baseline, and manifest artifacts. | The issue calls for an immutable Parquet workset; a source path or caller-owned row slice is not immutable evidence. |
| Does this lane deliver, receipt, or promote baseline? | No. It derives an unpromoted candidate baseline only. | Target DML and durable receipt binding are #3983/#3973 responsibilities; advancing here would bypass the transactional receipt gate. |
| How are deletes identified? | Only `synccontract.Tombstone` inputs supplied explicitly by a source. | The accepted design and issue both forbid treating snapshot absence as deletion. |
| How is determinism proven? | Derive twice from real Parquet inputs; compare identity bytes/hashes; mutate and replace the source afterward; reopen the original workset. | This detects both mutable-reference bugs and nondeterministic hashing/order. |

## Non-goals

- PostgreSQL connection opening, SQL/DDL generation, target database writes, or driver registration.
- Generic `WarehouseWorkset` changes in `internal/synctransport`; it remains an unrelated generic dispatch prerequisite and lacks the immutable managed-target identity needed here.
- Baseline promotion, durable delivery receipt, source checkpoint advancement, CDC/bootstrap, or capability promotion.
