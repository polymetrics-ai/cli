# PLAN — Issue 3981: durable target delivery ledger

## GSD setup and fallback

- Passed `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check`.
- Resolved sources and generated prompts for `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`.
- This issue foundation is not a numbered `.planning/ROADMAP.md` phase. The
  canonical single-worker contract prohibits role spawning, so the generated
  prompts are executed inline; this plan, TDD ledger, verification checklist,
  and review record are the manual fallback evidence.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, and `golang-database`.
  No CLI/docs parity skill is applicable because no CLI-visible surface changes.

## Goal

Provide a durable, driver-neutral target-delivery ledger addressed only by an
already asserted source owner, an observed target database identity, and one
immutable managed-target relation (`StreamID`). A provenance artifact rename
and a fresh ledger process must retrieve the same delivery record; sibling
relations must never share or overwrite a record.

## TDD slices

1. **Red — rename/restart lookup:** Add a test that records an observable
   delivery record through the intended API, rebuilds the ref from a renamed
   source artifact and a fresh ledger instance, then expects the same record.
   It must fail before the ledger exists.
2. **Green — sealed identity and durable port:** Add minimal validated ledger
   key/record types and a typed persistence port. The public ledger must derive
   its key only from `TargetOwner`, `TargetDatabaseIdentity`, and
   `ManagedTargetRef`; no table/display name is accepted.
3. **Red/green — independent relations:** Add two refs sharing one owner and
   namespace but carrying different StreamIDs. Assert two persisted records,
   independent lookup, and that updating one record does not mutate the other.
4. **Regression:** Assert invalid or mismatched owner/ref identity fails before
   fake-store mutation; ensure no API accepts `CommittedTransactionStage` or a
   source checkpoint as delivery authority.
5. **Verification/review:** Run focused package/app and race tests plus scoped
   repository gates. Execute `verify-work` and `code-review` prompts inline;
   use the GSD gap loop only if a real behavior gap remains.

## Guardrails

- Touch only the shared `internal/connectors/database` ledger foundation, its
  tests, and issue planning evidence.
- Do not edit the delivered provisioning kernel, any PostgreSQL driver, DDL,
  SQL, write session, transaction stage, mode application, source checkpoint,
  CLI, connector bundle, docs website, or generated code.
- The fake store is a test seam only. It supplies observable durable state to
  prove restart semantics without claiming native target persistence.
- Every rejected request must assert zero fake-store writes.

## Checkpoints

1. Commit plan/TDD/evidence artifacts.
2. Commit the preserved red-test output.
3. Commit green ledger implementation and focused tests.
4. Commit only review or GSD-gap fixes with focused green proof.
