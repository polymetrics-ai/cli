# Context — Issue 3973: transactional database write sessions

## Task Delivery Header

- Issue: Closes #3973 — Postgres Parity: bind database apply to transactional write sessions
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 → main
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green; after opening, verify the API-reported base is exactly `integration/4015-mvp-flat-r1`.
- Working branch: fm/cli-3973-write-sessions-r1
- Task: Build the shared, driver-neutral transactional database write-session layer: a sealed apply plan, approval admission/consumption, one bounded-batch session, rollback/unknown-commit semantics, a durable receipt before checkpoint authority, and mode-safe execution. No PostgreSQL driver/DDL/SQL or connector capability promotion is in scope.
- Verification: `go test -timeout 20m ./internal/connectors/database/... ./internal/app/...`, focused race tests, `go vet ./...`, `go build ./cmd/pm`, individual `make verify` gates, PR CI, and API read-back of its base branch.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| An approval is consumed before mutation and becomes invalid when target, schema, mode, keys, count, or destructive effects differ. | fake | No native database driver is in scope. A session-recording fake will assert zero `ApplyBatch` calls for every refused request and exactly one approval consumption before the first accepted batch. |
| Cancellation or a batch failure rolls back the entire pinned session and cannot advance a checkpoint. | fake | The fake has observable `Rollback`, batch, commit, and checkpoint counters. Tests will assert one rollback, no receipt, and zero checkpoint calls. |
| A durable receipt is required before checkpointing; an unknown commit outcome is neither retried nor labelled rolled back. | fake | The fake distinguishes `Commit`, `Rollback`, and `CommitOutcomeUnknown`; tests will assert no second commit/batch/rollback and no checkpoint on the unknown outcome. |
| `full_overwrite` is atomic and each supported mode is bound to its canonical closed strategy through one pinned session. | fake | A fake session records the selected mode/strategy and generation-publish hook. Tests assert the hook runs once only after bounded batches, while a non-atomic driver is refused before any mutation. |
| One pinned session carries bounded batches; the legacy per-record `Connector.Write` path cannot make the test pass. | fake | The fake driver exposes one session ID, batch sizes, and a legacy-write trap. Tests assert one begin, maximum batch size, and zero legacy writes. |
| The generic app/transport checkpoint path remains gated by downstream durable acknowledgement. | live | Existing app/transport test coverage is run unchanged and the new database receipt is converted to the same durable acknowledgement only after confirmed commit. |

## Discussion record

- The authoritative issue and firstmate brief resolve the material design choices: one pinned session, whole-session rollback, bounded batches, no raw SQL/target input, no PostgreSQL work, and no re-use of `CommittedTransactionStage` as a database transaction.
- `ManagedTargetControlRecord` supplies the already asserted owner, immutable target identity, target database identity, and schema. The new plan binds these values rather than accepting a relation or connection string.
- The existing closed `synccontract.Mode` and `connectors.DestinationApplyStrategy` vocabulary is consumed. This foundation does not add a new mode enum or generic strategy dispatcher.
- The project’s canonical single-worker delivery contract forbids spawning GSD roles in this environment. `discuss-phase` and `plan-phase --tdd` are therefore executed inline, with this context, plan, ledger, and verification checklist as the manual fallback record.
