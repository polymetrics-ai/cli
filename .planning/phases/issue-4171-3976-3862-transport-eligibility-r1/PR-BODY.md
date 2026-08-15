Closes #4171, #3976, #3862

Parent branch: `integration/4015-mvp-flat-r1`  
Parent PR: #4100 (draft, targets `main`)

## Intent

Close the three source-side transport gaps without widening the closed registry: GitHub admits only
its explicitly declared executable streams, PostgreSQL reaches the shared resumable polling executor,
and the production transport spine remains definition/reference/evidence-led.

## What changed

- Replaced GitHub's one-stream transport declaration with a concrete, no-wildcard allowlist matching
  every executable definition stream, including `commits`.
- Added `SourceStreamIneligibleError` at registry preflight. An absent stream is refused before an
  executor lookup, source request, destination plan, warehouse stage, apply, or checkpoint write.
- Made the declarative source stream-neutral: it honors `max_pages` (`omitted=1`, positive cap,
  `0`/`all`/`unlimited` exhaust), emits bounded batches, and uses source-owned candidate checkpoints
  for resume/replay.
- Promoted PostgreSQL `polling_watermark` only with its declared native binder. It performs typed
  catalog discovery, a strict cursor-plus-full-primary-key keyset query, `PollingPreflight`, and
  `engine.PollingSourceExecutor`; it adds no generic SQL or connector-name routing.
- Regenerated the connector catalog/docs and recorded the GSD TDD, verification, and inline-review
  evidence.

## Production call chains

GitHub commits to PostgreSQL:

`cmd/pm` → `cli.Run` → `app.RunETL` → `dispatchETLMode` → `shouldRunTransport` /
`runTransportETL` → `synctransport.Registry.Preflight` → `synctransport.Orchestrator.Run` →
`declarativeStreamSourceExecutor` → warehouse `Stage`/`Reopen` → PostgreSQL managed target →
read-back → durable checkpoint CAS.

PostgreSQL polling to PostgreSQL:

`cmd/pm` → `cli.Run` → `app.RunETL` → `dispatchETLMode` → `runTransportETL` →
`SnapshotTransportSource.ReadTransport` → `readPollingTransport` → `bindPollingSource` →
`engine.PollingPreflight` → `engine.NewPollingSourceExecutor` → native PostgreSQL keyset runner →
warehouse `Stage`/`Reopen` → PostgreSQL managed target → read-back → durable checkpoint CAS.

`app.Open` tests construct the registered definitions and closed registry used by these routes; they do
not hand-register a test-only transport executor.

## TDD and verification

Red:

- `.planning/phases/issue-4171-3976-3862-transport-eligibility-r1/traces/red-stream-admission.txt`
- `.planning/phases/issue-4171-3976-3862-transport-eligibility-r1/traces/red-postgres-polling-production.txt`

Green:

- `go test -timeout 20m -count=1 ./cmd/connectorgen ./internal/synctransport ./internal/connectors/engine ./internal/connectors/native/postgres ./internal/app`
- `go test -timeout 20m -count=1 ./internal/cli`
- `go test -timeout 20m -count=1 ./internal/connectors/certify`
- `go test -timeout 20m -race -count=1 ./internal/synctransport ./internal/connectors/engine ./internal/connectors/native/postgres ./internal/app`
- `go test -timeout 20m -tags=databaseintegration -count=1 ./internal/connectors/native/postgres ./internal/cli`
  (compiled and skipped by its opt-in guards; no shared container retry was attempted)
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connectorgen-certification-matrix`,
  `make github-parity-artifacts-check`, `make connector-boundary`, `make connector-canon-check`,
  `make release-workflow-check`, `go run ./cmd/agentcontractgen check`, and
  `scripts/verify-gsd-workflow` all passed.

| Test class | Evidence | Produced/refused value |
| --- | --- | --- |
| Happy path | `TestOpenComposedGitHubCommitsSourceEmitsEveryUnlimitedPageInBoundedBatches` | Production-composed `commits` reads 103 fixture records from two provider pages and emits five batches of at most 25. |
| Happy path | `TestOpenComposedGitHubCommitsHonorsTransportMaxPages` | Exact one-page default, bounded positive cap, and unlimited counts are asserted. |
| Bad path | `TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess` | Both an unknown stream and case-equivalent `ISSUES` return `SourceStreamIneligibleError` with zero source/stage/plan/apply/checkpoint effects. |
| Edge path | `TestPostgresPollingPlanBindsCompleteCompositeKeyAndStrictResume` | Complete `(cursor, primary-key tuple)` resume, nullable/missing key refusal, stale token rebootstrap, and out-of-order refusal are asserted. |
| Production construction | `TestOpenRegistersDefinitionOwnedProductionTransports` and `TestOpenComposesPostgresImplementedSharedPollingRoute` | `app.Open` resolves the exact GitHub/PostgreSQL references and shared polling declaration. |

### Edge-case coverage

| Edge | Coverage and state |
| --- | --- |
| cancellation mid-run | Orchestrator/app cancellation tests assert no interrupted checkpoint; binary route is integration-gated. |
| process death partway | Production-composed GitHub source replays the same unacknowledged candidate; PostgreSQL live equivalent is integration-gated. |
| empty input | Explicit empty-source marker has zero stage/apply/checkpoint effects; binary Postgres equivalent is integration-gated. |
| single row | Native PostgreSQL live test asserts one delivered row; pending opt-in runtime. |
| large input | 103-record bounded fixture is green; real 99,345-row certification remains pending. |
| duplicate delivery | Managed target binary replay assertion is integration-gated. |
| out-of-order delivery | Native keyset traversal returns a typed non-advancing refusal before downstream effects. |
| schema drift | Native and binary typed rebootstrap/zero-effect assertions are integration-gated. |
| auth refusal | Binary target refusal asserts unchanged target/checkpoint; pending rotated credential/runtime. |
| concurrent runs on one target | `internal/app/transport_dispatch_test.go` state-CAS conflict coverage remains green. |
| resume after interruption | GitHub candidate resume is green; duplicate-cursor PostgreSQL resume is integration-gated. |
| replay of an acknowledged item | Orchestrator acknowledgement sequencing is green; managed target binary replay is integration-gated. |
| undeclared and case-equivalent stream | Typed `SourceStreamIneligibleError`, zero effects, green locally. |

## Live certification status

`TestPMBinaryExecutesAuthenticatedGitHubCommitsWarehousePostgres` is the decisive binary path. It
configures `max_pages=unlimited`, requires at least 99,345 extracted `rails/rails` commits, then
independently compares the warehouse and PostgreSQL row counts to the extracted count.

It is **pending**, not claimed: the task's shared container runtime was unavailable and expressly
must not be retried, while the GitHub certification credential is pending rotation. No credential was
read, printed, stored, committed, or included here, and no live row count is reported.

## Lifecycle, artifacts, and parity

- Inline/manual GSD fallback recorded because this non-numbered direct-PR issue club requires one
  autonomous worker and the canonical lifecycle contract forbids role spawning. Resolved prompts:
  `discuss-phase --auto`, `plan-phase --tdd --auto`, `execute-phase --interactive`, `verify-work`,
  and `code-review`.
- Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-context`, `golang-concurrency`, and `golang-database`.
- Derived artifacts regenerated once with connectorgen, certification-matrix, GitHub GraphQL parity,
  website data, and connector docs generators; all drift checks passed before this push.
- No CLI command/flag/output changed. `pm help connectors`, `pm connectors`, and
  `pm connectors --help` succeeded; generated connector catalog/manual/skill docs reflect the
  definition changes.
- Scope deliberately excludes #4125, #4158, and #4169. No dependencies or generic write/query
  surface was introduced.

## Automated review routing

Primary route: `claude_auto` on PR creation. Status: pending automatic GitHub Action review for the
head commit range. Fallback: none unless that route fails, skips, or is rate-limited; parent PR #4100
is available for the non-default-base review fallback. No manual Claude or Copilot request was made.

## Post-PR certification guard correction

A clean comparison found the exact source of the reported failure before any edit: merge base
`ef3c71caf` passed both `TestCertificationDeclaredTransportPair*` guards, while this branch at
`73280ed81` failed because the guards named the source executor replaced by this PR. The tests now
assert the active declaration-owned `declarative_stream_source` on the resolved and deliberately
unregistered paths. `go test -timeout 20m -count=1 ./internal/connectors/certify` passes; no
assertion was removed or relaxed.
