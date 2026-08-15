# PLAN — Issue 3973: transactional database write sessions

## GSD setup and fallback

- Passed `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check`.
- Resolved sources and generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.
- This issue foundation is outside a numbered `.planning/ROADMAP.md` phase. The canonical single-worker contract prohibits role spawning, so the generated prompts are executed inline; this plan, TDD ledger, verification checklist, and review record are the manual fallback evidence.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`.
- No CLI/docs parity skill applies: the slice exposes no command, flag, help, manual, website, or connector-definition change.

## Goal

Provide a typed driver-neutral apply boundary that accepts only a sealed target
control record and a preview-bound approval. It opens exactly one session, runs
bounded batches through that session, rolls it back as a whole when execution
does not safely commit, returns a durable target receipt only after confirmed
commit, and never turns an indeterminate commit into either a retry or a
rollback claim.

## TDD slices

1. **Red — sealed plan/approval:** Add a fake driver test for a valid preview
   and an otherwise identical target/schema/mode/key/count/destructive-effect
   mutation. It must expect an admission error and zero driver/session/batch
   mutation before the implementation exists.
2. **Green — typed plan and approval admission:** Add validated, immutable
   plan/approval types keyed by `ManagedTargetControlRecord`, the closed
   `synccontract.Mode`, canonical key fields, record count, and destructive
   effects. Consumption is one-shot and happens before `BeginWriteSession`.
3. **Red/green — pinned bounded session:** Exercise a six-record workset with
   a two-record limit and prove one session ID, three bounded batch calls,
   one commit, one durable receipt, and no legacy `Connector.Write` call.
4. **Red/green — failure/cancellation/receipt gate:** Make the fake fail a
   later batch and cancel between batches; assert exactly one rollback and
   zero checkpoint authority. Make commit indeterminate; assert the explicit
   unknown result, no retry/rollback, and zero checkpoint authority.
5. **Red/green — mode contracts:** Require atomic publish support for
   `full_overwrite`; run append/upsert/dedupe only through the selected
   session strategy and reject incompatible mode/strategy before mutation.
6. **Integration/regression:** Adapt the confirmed receipt to the existing
   downstream-acknowledgement gate without touching its semantics. Preserve
   PostgreSQL’s descriptor-only `write=false` fence and `Connector.Write`
   unsupported test.

## Guardrails

- Touch only `internal/connectors/database`, its tests, narrow app/contract
  integration only if needed for the receipt gate, and issue planning evidence.
- Do not add a dependency, database DDL/SQL, raw relation/command input,
  generic database-write command, PostgreSQL writer, connector registration,
  capability promotion, source checkpoint bypass, CLI/doc/website surface, or
  changes to `CommittedTransactionStage`.
- Treat provider/driver errors as opaque boundary failures; neither error text
  nor records/credentials may leak into the shared contract.
- Every refusal must prove zero begin, zero batch, zero commit, zero receipt,
  and zero checkpoint side effects as applicable.

## Checkpoints

1. Commit plan/TDD/evidence artifacts.
2. Commit preserved red-test output.
3. Commit green session implementation and focused tests.
4. Commit only review or GSD-gap fixes with focused green proof.
