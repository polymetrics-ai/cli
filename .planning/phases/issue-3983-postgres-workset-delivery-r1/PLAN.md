# Plan — Issue 3983: PostgreSQL workset delivery

## Goal

Bridge immutable keyed Parquet worksets into one managed PostgreSQL write
session. The bridge must preserve #3980's immutable workset identity, #3973's
approval/receipt semantics, and #3982's managed target ownership checks.

## TDD slices

1. **Red — sealed delivery-plan binding.** Add database-package tests that
   compile against the missing workset delivery plan/controller. Derive a real
   Parquet workset, seal a mapped `incremental_upsert` plan, and vary workset,
   owner/OID/schema, and ordered keys. A stale approval must leave fake
   `BeginDatabaseWrite` and `ApplyWriteBatch` counters at zero.
2. **Green — approved workset apply.** Implement the smallest concrete
   immutable controller/plan. It reads only `ReadDelta` and `Tombstones`,
   remaps the sealed tombstone keys through the accepted mapping, and delegates
   preview/session/receipt to `DatabaseWriteExecutor`. Assert fake sessions see
   exact changed records and tombstones, never rows inferred from absence.
3. **Red/green — receipt-bound baseline ledger.** Add a per-destination
   baseline store sealed by the same `ManagedTargetDeliveryLedgerKey`. Its
   record binds workset identity/content, baseline candidate, and receipt.
   Assert prior baseline identity remains unchanged on batch failure,
   receipt-store failure, ledger failure, and `CommitOutcomeUnknown`; successful
   delivery promotes exactly the candidate after receipt/ledger confirmation.
4. **Red/green — PostgreSQL live delivery.** Extend the existing opt-in dbtest
   fixture with a real Parquet workset and managed driver/controller. Query
   actual target rows for insert/update/unchanged/composite-key behavior; prove
   physical absence retains a prior target row and an explicit tombstone deletes
   it. Query the driver ledger/receipt and observed promoted baseline state.
5. **Refactor/review.** Run race-aware targeted tests and the required live
   dbtest. Inspect the final changed paths for ownership, SQL construction,
   cancellation, receipt order, and no-secret guarantees; record dispositions.

## Guardrails

- Only the sealed mapped `incremental_upsert` path is admitted. There is no
  arbitrary table, SQL, direct target connection, source-to-destination route,
  target full overwrite, physical-absence delete, implicit retry, or capability
  promotion.
- Validate every caller-provided mapping/key/owner identity before session
  open. Propagate `context.Context`; use a non-cancelled cleanup context only
  where the existing executor's durable rollback/ledger semantics require it.
- Batches remain bounded by the sealed `DatabaseWritePlan`. Mapping and
  tombstone records are cloned/validated at the boundary.
- A per-destination key retains workspace, connector, connection, target
  database identity, stream ID, namespace, and relation. No display/table name
  can select another destination's baseline.
- Do not modify #4125, #4136, or #4090. If testing exposes one, add a
  `needs-decision:` status entry rather than silently absorbing it.

## Checkpoints

1. Commit phase evidence and failing targeted test output.
2. Commit the minimal shared delivery bridge and green unit tests.
3. Commit PostgreSQL live proof and any small driver-owned baseline persistence
   needed by the shared contract.
4. Rebase, execute targeted/live/local gates, review, push, open the explicit
   base PR, verify the base via the GitHub API, and wait for green CI.
