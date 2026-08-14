# Plan — Issue 4094: PostgreSQL incremental dedupe history

## Goal

Implement the PostgreSQL-only `incremental_dedupe_history` managed target on
the existing mapping, receipt, managed-driver, workset, and delivery-ledger
foundations. Its non-PostgreSQL routes must fail with typed, route-specific
errors before any provider or database I/O.

## TDD slices

1. **Red — route admission.** Add tests that invoke every non-PostgreSQL route
   with operation counters. Require a route-specific typed reason and all
   provider/database counters to remain zero.
2. **Green — exclusive route gate.** Implement the smallest preflight route
   gate at the managed-target boundary, so invalid routes cannot acquire a
   source reader, target driver, write session, or mutate a ledger.
3. **Red — history state.** Add PostgreSQL tests that define the missing
   history target behavior: exactly one current keyed version, closed prior
   validity windows, soft-delete close, deterministic late/replay outcome,
   receipt, and restart recovery.
4. **Green — native target.** Reuse `MappingContractV1`, `DeliveryReceiptV1`,
   sealed workset delivery, and the existing ledger to persist the exact rows.
   Keep data access bounded, context-aware, and parameterized through existing
   native driver APIs.
5. **Live proof and review.** Run the requested tagged dbtest and query actual
   history rows. Execute targeted unit tests, static gates, GSD verify-work,
   and a focused review before opening the explicit-base PR.

## Guardrails

- One target connector only: PostgreSQL. Changed production paths must remain
  within `internal/connectors/database/**` and
  `internal/connectors/native/postgres/**` as required by the existing seams.
- No generic SQL or arbitrary provider/write surface; no new dependencies or
  schema/contract reinventing.
- A rejected route has no observable source, target, session, ledger, or
  database operation. The error must expose its typed reason through the
  existing error-inspection convention.
- All success tests assert row state and receipt/ledger state. They do not use
  an error-free return as success evidence.
- No command/help/manual/website surface is expected to change. Reassess before
  implementation if code discovery contradicts that assumption.

## Checkpoints

1. Commit GSD context/plan/TDD checklist and red test output.
2. Commit route gate plus green rejection tests.
3. Commit history target plus green unit/live proofs.
4. Rebase, run required verification and review, push, open PR explicitly
   against `integration/4015-mvp-flat-r1`, and observe its base via API.
