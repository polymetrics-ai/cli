## Intent

Refs #4176

Make certification observe an installed schedule's actual `schedule fire` execution, and prevent full runs from silently losing the ordinary flow/schedule stage set.

## What Changed

- Added a `schedule_fire` certification stage after install. It requires a `ScheduleFire` response, flow status `ok`, and terminal schedule status `succeeded` before removal and byte-for-byte backend restoration.
- A failed install assertion now refuses `schedule fire` before its CLI invocation; the stage and aggregate report fail even when cleanup succeeds.
- Reports the unstarted scheduler-daemon boundary as `capabilities.schedule.result=not_live`; it does not call it pass.
- Added full-versus-ordinary report-stage comparison. The investigation found that the current base already executes flow/schedule stages inside the full per-stream sweep; this test protects that real path against regression.

## Test Contract

- Happy: `TestFullCertificationStageSetIsStrictSuperset`; `TestGlueStagesScheduleFireObservesInstalledFlowAndRestoresBackend`.
- Bad: `TestGlueStagesScheduleFireRefusalFailsBeforeRemovalSuccess` asserts the exact install-gated refusal, zero fire calls, aggregate failure, and cleanup.
- Edge: `TestGlueStagesScheduleFireEmptyBackendIsRestoredAndDaemonIsNotLive` asserts byte-for-byte empty backend restoration and explicit `not_live` reporting.
- Shipped construction path: `TestCertifyCLISingleConnectorPassExitsZero` runs `pm connectors certify sample --full --json` and asserts two `schedule_fire` stages.

## Testing

- `go test -timeout 20m ./internal/connectors/certify -count=1`
- `go test -timeout 20m ./internal/cli -count=1`
- `go vet ./internal/connectors/certify ./internal/cli`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`
- `scripts/verify-gsd-workflow`

## GSD / Review

- Lifecycle prompts resolved and executed inline: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`.
- Inline fallback is recorded because the canonical delivery contract and current runtime require one active worker.
- Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, and `golang-lint`.

## CLI parity

No public command/flag/help/documentation surface changed. Existing `connectors certify --full` and `schedule fire` docs remain applicable; `pm help connectors`, `pm connectors`, and `pm connectors certify --help` were checked.

## Explicit non-live boundary

The certification harness directly fires the installed schedule through product CLI code against an isolated crontab fixture. It does not start a real crontab, systemd, or Temporal daemon, and makes no credentialed provider call. That untested boundary is explicitly `not_live`, not a silent pass.
