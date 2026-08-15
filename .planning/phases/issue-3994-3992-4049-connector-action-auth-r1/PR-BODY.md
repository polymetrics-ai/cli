Closes #3994

Closes #4049

Refs #3992

## Outcome

- Adds payload-bound prepared-execution identity and durable replay refusal to the connector action path.
- Makes flow creation positively resolve existing ETL connections and already-approved reverse-ETL jobs before any flow write.
- Makes schedule create/install positively resolve an existing valid flow before any schedule/backend write.
- Keeps the installed command exactly `pm --root <root> flow run <name> --json`; no approval token or schedule authority is rendered or stored.
- Revalidates standing job approval on every firing and parks the existing durable fire state on drift or ambiguous failure.
- Exposes `*connsdk.RateBudgetRefusalError` / `shared_coordinator_unavailable` before provider transport when `require_shared` cannot coordinate.

## Production call chains

### #3994 and #3992

Creation: `cmd/pm` -> `cli.Run` -> `runFlow`/`flowCreate` -> `resolveManifestJobs` ->
`App.GetConnection` or `App.GetReversePlan` -> `App.ValidateAuthorizedFlowAction` -> atomic flow
manifest write.

Schedule creation/install: `cmd/pm` -> `cli.Run` -> `runScheduleCreate`/`runScheduleInstall` ->
`validateStoredScheduledFlow` -> `resolveManifestJobs` -> `schedule.Save`/backend `Install`.

Installed firing: `cmd/pm` -> `cli.Run` -> `flowRun` -> `schedule.FindByFlow` ->
`schedule.BeginFire` -> `resolveManifestJobs` -> `flow.Engine.Run` ->
`connectorFlowActionRunner.ExecuteStep` -> `App.ExecuteAuthorizedFlowAction` ->
`PrepareAuthorizedFlowAction` -> `ExecutePreparedFlowAction` -> connector `ValidateWrite`/`Write`/
`Read` -> receipt/checkpoint -> `FireLease.Complete`; any unsafe error goes to `FireLease.Park`.

`TestPMBinaryExecutesInstalledApprovedJobFlow` builds a fresh `cmd/pm`, creates and approves the job,
creates/installs the schedule, executes the exact installed argv, observes destination read-back and
terminal schedule state, and verifies the same `pex_` identity reached the flow and fire receipt.

### #4049

`cmd/pm` -> `cli.Run` -> connector command -> `commandrunner` ->
`engine.Runtime.RequesterFor` -> `rateLimitResolver` ->
`RateBudgetRefusalError(shared_coordinator_unavailable)` before `Requester.DoJSON`/transport.

`TestPMBinaryRefusesRequiredSharedRateBudgetBeforeSend` builds a fresh binary and observes the stable
JSON policy code with an instrumented transport count of zero.

## Required edge-case evidence

| Criterion | Typed/terminal result | Zero-side-effect assertion |
| --- | --- | --- |
| Cancellation mid-operation | `context.Canceled`; scheduled fire parks | zero pre-dispatch sends/receipts/checkpoint; prepared lease is released only when no dispatch was possible |
| Process death partway | persisted `running` fire / prepared marker refuses replay | no automatic replay or checkpoint advance after reopen |
| Already-granted/already-fired replay | `PreparedExecutionReplayError`, `ErrFireInProgress`, or `ErrFireParked` | concurrent/replayed path performs zero additional writes |
| Expired or revoked approval | typed authorization expiry/revocation wrapped by `JobReferenceError`; fire parks | zero provider requests and zero checkpoint advance |
| Refused approval | `JobReferenceError{Reason: unapproved}` | no flow file and no provider write |
| Shared coordinator unavailable | `RateBudgetRefusalError{Code: shared_coordinator_unavailable}` | zero HTTP sends and no rate checkpoint/admission |
| Concurrent firings | one prepared identity/fire lease wins; loser gets typed replay/in-progress error | exactly one connector write, no losing checkpoint |
| Partial write / cleanup failure | ambiguous delivery or cleanup reason parks/halt state | no receipt/checkpoint and no automatic replay; cleanup occurs only after durable terminal state |

## Approval and secret boundary

Approval remains on the ETL/reverse-ETL job at connection + schema + preview granularity. Flows and
schedules inherit that approval; neither creates a second grant. Tests scan the rendered crontab and
project state for the one-time approval token and carrier field names. No token is written to
crontab, argv, flow/schedule/fire state, logs, or JSON.

No GitHub certification credential or database-integration runtime variable was available in this
worktree, so no credential was requested or printed. The deterministic fresh-binary production
proofs are included here; #4166 separately owns credentialed live-certification coverage. #4125 and
#4158 remain untouched.

## Verification

- Focused `internal/app`, `internal/flow`, `internal/schedule`, `internal/cli`, connsdk, engine, and GitHub hook tests pass with `-timeout 20m`.
- Selected race test `TestAuthorizedFlowActionConcurrentPreparedExecutionHasOneWinner` passes.
- Fresh-binary installed firing and shared-coordinator refusal tests pass.
- CLI help/manual, generated docs, golden transcripts, and website docs are updated.
- Final one-pass generation completed; `surface-sync --check` scanned 552 bundles with zero drift; changed Go files are gofmt-clean; `git diff --check` and final `git status` are clean.

## Required skills used

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, and the repository's
GSD discuss/plan/execute/verify/code-review, CLI parity, issue-delivery, and runtime-integration
references.

## Delivery record

- GSD lifecycle: discuss -> TDD plan -> execute -> verify-work -> code-review.
- Execution mode: documented inline/manual fallback because the canonical contract requires one worker and forbids spawned roles for this job.
- Review: inline review passed with no unresolved actionable findings; Claude automatic review remains the primary PR review gate.
