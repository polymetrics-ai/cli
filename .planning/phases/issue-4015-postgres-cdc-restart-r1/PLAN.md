# Plan — PostgreSQL CDC Restart Recovery for 0.2.1

## Delivery and lifecycle

This is the explicit inline/manual GSD fallback recorded in `CONTEXT.md`: the repository's canonical contract requires one worker and forbids role spawning. The required command path is `discuss-phase cli-cdc-resume-fix-r1`, `plan-phase cli-cdc-resume-fix-r1 --tdd --skip-research --auto`, `execute-phase cli-cdc-resume-fix-r1`, `verify-work cli-cdc-resume-fix-r1`, and `code-review cli-cdc-resume-fix-r1`.

Required skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`.

## Scope and ownership

- Target connector: exactly `postgres`.
- Production ownership: `internal/connectors/native/postgres/**` only unless the reproduced call path proves a shared database-sync foundation defect; a shared foundation change would stop this connector lane and require a separate issue.
- Test/evidence ownership: focused PostgreSQL tests, the existing PostgreSQL pipeline database-integration test as needed for restart proof, this phase directory, and the existing PostgreSQL CDC capability artifact if its claim changes.
- Out of scope: `release/0.2.0-mvp`, PR #4250, unrelated connectors, generic SQL, connector-to-connector shortcuts, CLI command/flag/help changes, new dependencies, daemon lifecycle changes, and runtime restarts.
- Changed-path compliance will be checked with `git diff --name-only origin/integration/4015-mvp-flat-r1...HEAD` before delivery.

## Foundation check

| Need | Required proof | Planned evidence |
| --- | --- | --- |
| PostgreSQL CDC executor | The registered executor is streamed `pgoutput` v2, not polling fallback. | Trace `ReadCDC`/bootstrap/restart dispatch and assert the executor/checkpoint mechanism in focused tests. |
| Durable position | Receipt succeeds before checkpoint persistence, which succeeds before PostgreSQL acknowledgement. | Existing ordering tests plus the new restart regression assert the persisted LSN and no advance on injected failures. |
| Resume identity | Checkpoint binds source/system/timeline/slot/publication/relation/schema identity. | Existing drift matrix remains green; new regression uses the same durable envelope on restart. |
| Warehouse/target effect | Post-interruption records reach the managed PostgreSQL target once. | Live binary harness plus an independent target query for exact count and per-key multiplicity. |
| Capability claim | Certification evidence describes the behavior actually proven. | Inspect and, if required, update `postgres_cdc_r1-capability-cdc.json` with exact restart limitations/results. |

## TDD slices

1. **Red/live — reproduce before repair.** Run the existing opt-in PostgreSQL pipeline restart scenario against the already-running local runtime. Preserve the exact checkpoint mechanism/outcome and independent target absence/count evidence. Confirm or refute the supervisor's mechanism-mismatch hypothesis.
2. **Red/focused — lock the correct restart contract.** Add the smallest failing test that performs a first CDC run through durable checkpoint commit, constructs a fresh executor/process boundary, resumes with that checkpoint, emits a later committed transaction, and asserts the later transaction is delivered once from the correct LSN. The red assertion must fail on current production code for the same reason as live.
3. **Green — repair producer/consumer agreement.** Change the checkpoint production/validation/resume path only after tracing the live value. The accepted mechanism must remain logical replication and retain all source/slot/publication/schema identity checks, receipt-before-checkpoint-before-ack ordering, bounded stage, and explicit rebootstrap outcomes for genuinely incompatible state. Do not accept polling checkpoints or weaken the validator.
4. **Green/live — process-death recovery proof.** Interrupt after a durable position is established, commit a uniquely keyed row after interruption, restart with the persisted checkpoint, and independently query the target. Record exact rows before interruption, at interruption, and after restart; assert the new key occurs exactly once and earlier keys were not duplicated or lost.
5. **Truthful evidence and refactor.** Update the CDC capability artifact only if the fixed/proven behavior changes its truth. Keep CLI/help/docs/website parity explicitly not applicable unless the investigation discovers a user-facing contract change. Run focused and structural gates, then inspect the diff for unrelated changes.

## Red/green/refactor evidence

- Red traces are written under `traces/` before the production edit and include the exact failing command plus observable target/checkpoint assertions.
- Green traces use the same focused and live scenarios after the fix.
- Refactor is limited to removing duplication or clarifying invariants exposed by the fix; no speculative checkpoint version or abstraction is added.

## Verification commands

1. Focused PostgreSQL unit/restart regression: `go test -timeout 20m ./internal/connectors/native/postgres/... -run '<restart-test>'`.
2. PostgreSQL package: `go test -timeout 20m ./internal/connectors/native/postgres/...`.
3. Live PostgreSQL pipeline recovery against the already-running configured endpoint: the exact `databaseintegration` command recorded in the red/green traces.
4. CLI regression where touched: `go test -timeout 20m ./internal/cli`.
5. Format/vet/build: `gofmt -w <changed Go files>`, `go vet ./...`, `go build ./cmd/pm`.
6. Structural/generated gates: `scripts/verify-gsd-workflow`, `make tidy-check lint`, `make docs-check smoke-no-build`, `make agent-contract-check connectorgen-validate connectorgen-surface-sync connector-runtime-preflight connector-canon-check connector-boundary release-workflow-check`.
7. Diff hygiene: `git diff --check` and changed-path inspection.
8. Full `go test -timeout 20m ./...` / aggregate `make verify` are delegated to CI per `AGENTS.md`; all named non-full-suite gates run separately locally.

## Commit and push checkpoints

1. Planning header/context/plan/TDD checklist.
2. Red regression and preserved live-red evidence when coherent.
3. Green implementation and focused/live proof.
4. Review fixes and final verification evidence.
5. Push only `fm/cli-cdc-resume-fix-r1`; open a direct PR into `integration/4015-mvp-flat-r1`; verify the API-reported base.

## Plan verification

- Every acceptance criterion has an observable live assertion in `CONTEXT.md`.
- The plan starts with reproduction and a failing test, not a production edit.
- The design decision remains evidence-driven and bounded by the canonical logical-replication contract.
- No task step touches the 0.2.0 branch/PR, restarts the shared runtime, reads secrets, or bypasses warehouse mediation.
