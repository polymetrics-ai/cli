# PLAN — Issue 403 progress event bus

Sub-issue: https://github.com/polymetrics-ai/cli/issues/403
Parent: https://github.com/polymetrics-ai/cli/issues/397 / PR #438
Branch: `feat/403-progress-event-bus`
Original base parent head: `e5ee4075`
Current parent head: `f12d573b6415aed2c47cb3fd346c564d3b752a60` (includes parent #421 and parent ledger checkpoint)
Production race head: `f16207974cc25f6df111fcd2a99c6acec41f3c44` (coordinator external PR-head source after rebase onto current parent; strict full-race passed)
Post-race artifact policy: commits after `f16207974cc25f6df111fcd2a99c6acec41f3c44` may update issue-local artifacts/PR body only; no production self-SHA chase.
Mode: bounded mutating worker in isolated cwd.

## Required reading complete

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/claude-review-loop.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`
- `docs/plans/universal-programming-loop-prd.md`, `docs/prompts/universal-programming-loop-prompts.md`
- `docs/plans/cli-architecture-v2-improvement-plan.md` Stage 5 / Pillar B
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md` Stage 5
- `docs/design/tui-ux-design.md` event/NDJSON/import-law sections
- `docs/adr/0003-interactive-tui-layer.md`
- Issue #403 and parent issue #397 bodies via `gh issue view`.

## GSD adapter evidence

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt plan-phase 403 --skip-research` generated the official `/gsd-plan-phase 403 --skip-research` prompt.
- `scripts/gsd prompt programming-loop init --phase 403 --dry-run` failed: `scripts/gsd: unknown GSD command: programming-loop`.
- Adapter gap fallback: loaded `.pi/prompts/pm-gsd-loop.md` and run the GSD universal programming loop inline/manual for this worker. Record as `local_critical_path`, not spawned.

## Skills loaded

Routing source: `.agents/agentic-delivery/references/required-skills-routing.md`.

- `gsd-core`
- `caveman` for compact handoff only
- `golang-how-to`
- `golang-testing`
- `golang-context`
- `golang-concurrency`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-lint`
- `golang-code-style`
- `golang-documentation`
- `golang-naming`
- `golang-troubleshooting`

## Scope

Allowed writes:

- new dependency-free `internal/events/**`
- named instrumentation in `internal/flow/engine.go`
- ETL flush instrumentation in `internal/app/app.go` and `internal/app/local_warehouse.go`
- certify pool instrumentation in `internal/connectors/certify/batch.go`
- worker polling instrumentation in `internal/worker/submit.go`
- focused tests beside these packages
- issue-local `.planning/phases/403-progress-event-bus/**`

Explicit non-scope:

- no `internal/cli/**`, docs/website, `go.mod`, `go.sum`
- no logging/vault/Temporal logger setup (#404)
- no telemetry/OTel/connsdk tracing (#410)
- no catalog (#406)
- no `--progress` CLI flag (#405)
- no service/credential/external writes.

## Slice plan

### Slice 1 — events package API + sinks

Red tests first under `internal/events`:

1. context default returns Nop and typed emitter can be carried via context;
2. NDJSON sanitizes terminal controls and redacts secret-like values using `internal/safety`;
3. Chan preserves lifecycle events under backpressure, coalesces progress, and accounts dropped progress;
4. Multi emits to all sinks without races;
5. Throttle forwards lifecycle immediately and coalesces/throttles progress;
6. concurrent Emit is `go test -race` clean.

Implementation:

- `Event` typed struct with kind/scope/run/step/status/message/counters/timestamp/attrs.
- `Emitter` interface: `Emit(context.Context, Event)`.
- `WithEmitter`, `FromContext`, `Nop` default.
- Sinks: Nop, NDJSON(writer), Chan(capacity), Throttle(interval, sink), Multi(...).
- Stdlib + `internal/safety` only in `internal/events`.

### Slice 2 — flow instrumentation

Red test in `internal/flow`: collector emitter sees deterministic flow/step sequence for success, skip, and failure.

Implementation:

- emit flow start/done/error and step start/done/error/skipped beside existing ledger calls.
- do not change ledger/checkpoint semantics or reverse ETL approval gate.

### Slice 3 — ETL flush instrumentation

Red tests in `internal/app`: connector destination and local warehouse ETL produce deterministic run start, batch flush, and done events with counts.

Implementation:

- emit ETL run lifecycle in `RunETL`/`failRun`/`completeRun`.
- emit batch progress after successful destination/local warehouse flush.
- no per-record event hot loop.

### Slice 4 — certify batch instrumentation

Red test in `internal/connectors/certify`: worker pool emits batch start, per-connector queued/running/done/error/skipped/resumed, and batch done in deterministic connector order where concurrency allows; lifecycle not silently dropped through Chan.

Implementation:

- emit queued/skipped/resumed before workers.
- emit running/done/error from workers.
- preserve context cancellation and existing worker-pool behavior.

### Slice 5 — worker polling instrumentation

Red test in `internal/worker`: submitter polling emits workflow submitted/polling/done while preserving cancellation and requiring no external Temporal service.

Implementation:

- introduce small unexported submitter dependency seam for tests around Temporal client/run.
- poll `DescribeWorkflowExecution` while waiting for `run.Get` in a goroutine.
- select on `ctx.Done()` and result channel; no goroutine leaks.
- do not alter Temporal logger setup.

## Review-fix plan — 2026-07-17

PR #451 review findings accepted:

1. HIGH Chan backpressure/close semantics: keep the internal queue strictly within configured
   capacity; evict/coalesce progress first for lifecycle insertion; if the queue is full of lifecycle
   events and the consumer stalls, use a bounded wait with context cancellation/close handling and
   explicit lifecycle drop accounting; define/test deterministic close-drop accounting; keep Close
   finite; prevent Multi fanout from being indefinitely stalled by Chan.
2. MEDIUM Throttle stale progress after terminal: flush the latest pending progress before terminal
   lifecycle events so completed/failed/skipped remain last, with tests that do not codify terminal
   then progress ordering.
3. Race evidence trace: production fixes invalidate the previous strict race evidence at
   `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61` and current head `e5404809fc66296f6d02e243b09b431dade921fb`.
   At review-fix time `RUN-STATE.json` recorded `verificationPassed=false`; this is superseded by
   the coordinator strict full-race pass at production head `f16207974cc25f6df111fcd2a99c6acec41f3c44` below.
4. Residual event sequence gaps: add focused terminal/error/skip/cancel tests only where needed to
   support issue #403 acceptance without broad fixture churn.

Review-fix TDD slices:

1. Add failing `internal/events` tests for Chan strict capacity/accounted lifecycle timeout,
   deterministic close-drop accounting, Multi+Chan finite fanout, and Throttle terminal ordering.
2. Add focused existing-instrumentation sequence tests for skipped flow steps, ETL failure terminal
   event, certify skip/error terminal paths, and worker successful terminal path if gaps are present.
3. Implement minimal events package fixes, keeping dependency-free stdlib + `internal/safety` only.
4. Re-run the requested focused race/non-race gates and update this phase's TDD/verification
   artifacts and PR body with accepted finding dispositions and then-pending strict full-race status.

Review-fix result:

- Chan/Throttle red test captured before production fix; green focused/race events tests passed.
- Focused flow/app/certify/worker sequence tests passed under `-race`.
- `go vet ./...`, focused non-race package tests, `go build ./cmd/pm`, `make verify`, diff-check,
  dependency inspection, and go.mod/go.sum check passed.
- Strict full-race was pending parent orchestrator rerun at review-fix handoff; superseded by final PR-head pass at `f16207974cc25f6df111fcd2a99c6acec41f3c44`.

## Second targeted review-fix plan — 2026-07-17

Starting head: `c9813a788d2bc0ccc29e79920ce6e5e8084e8a8e` on PR #451.

Accepted findings:

1. MEDIUM `Chan.Close` in-flight accounting: current close can return after closing `done` but
   before the runner accounts an event already removed from the queue and blocked on `out`, so
   immediate `DropStats()` and `Events()` closure are nondeterministic. Add a runner stopped/closed
   acknowledgment. `Close` must stay finite with stalled/no consumer; runner must observe `done`,
   account any in-flight drop exactly once, close `out`, then acknowledge.
2. LOW `Multi` contract gap: `Multi` is synchronous and cannot make arbitrary custom or writer
   sinks finite. Narrow comments, tests, artifacts, and PR claims to `Multi` with bounded `Chan`,
   and document that blocking sinks must honor context/cancellation or otherwise be finite. Do not
   add goroutine-per-sink fanout.

Second review-fix TDD slices:

1. Add failing `internal/events` regression that forces an in-flight lifecycle event blocked on
   `out` with a queued progress event and no consumer; assert `Close` returns finite, `Events()` is
   closed immediately after `Close`, and `DropStats()` is exactly one lifecycle plus one progress
   drop.
2. Implement minimal `Chan` runner acknowledgment/stopped channel; ensure `Close` waits for the
   runner after closing `done`, without holding the mutex, so in-flight drop accounting and `out`
   closure are deterministic.
3. Narrow `Multi` docs/test names/artifacts/PR body to synchronous fanout plus bounded-`Chan`
   finite behavior; no asynchronous fanout.
4. Run requested local gates; strict full-race was pending at this step because production changed and is superseded by the final PR-head pass below.

Second review-fix result:

- Red regression captured: `go test ./internal/events/... -run TestChanCloseWaitsForInFlightEventAccounting -count=1` failed with `DropStats() after Close = {Progress:1 Lifecycle:0}, want {Progress:1 Lifecycle:1}`.
- `Chan` now has a runner `stopped` acknowledgment; `Close` waits for `out` closure and in-flight drop accounting. Regression green asserts exact `{Progress:1 Lifecycle:1}` for stalled/no-consumer in-flight close.
- `Multi` contract narrowed: synchronous fanout; bounded `Chan` finite, arbitrary custom/writer sinks must be finite or context-aware.
- Requested local gates and `make verify` passed; strict full race was pending until the coordinator's final PR-head rerun below.

## Final PR-head finalization plan — 2026-07-17

Task: finalize PR #451 / issue #403 after coordinator rebased `feat/403-progress-event-bus` onto current parent `f12d573b6415aed2c47cb3fd346c564d3b752a60` and passed strict race on production head `f16207974cc25f6df111fcd2a99c6acec41f3c44`.

External strict race evidence source: coordinator PR-head run, not this worker's self-SHA chase.

```text
go test -race ./... -count=1 -timeout 120m
PASS
internal/cli 1842.794s
internal/connectors/certify 3802.054s
internal/events 2.665s
internal/flow 2.590s
internal/worker 1.611s
real 3809.60
user 6223.47
sys 77.80
```

Finalization slice:

1. Run `gofmt -w cmd internal` and stop if it changes `cmd/**` or `internal/**`.
2. Run requested local gates: `go vet ./...`, focused package tests, `go build ./cmd/pm`, `make verify`, diff-check, and go.mod/go.sum check.
3. Set `verificationPassed=true` only after the external strict race plus all requested local gates pass.
4. Commit/push issue-local artifact updates only; production head for race remains `f16207974cc25f6df111fcd2a99c6acec41f3c44`.

Finalization result:

- `gofmt -w cmd internal` passed and left no `cmd/**` or `internal/**` diff.
- `go vet ./...` passed with no output.
- `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` passed: `events 0.604s`; `flow 0.793s`; `app 17.306s`; `certify 335.483s`; `worker 0.534s`.
- `go build ./cmd/pm` passed with no output.
- `make verify` passed: `go test -timeout 20m ./...` passed (`internal/cli 163.819s`, `internal/connectors/certify 336.906s`); `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` passed with no output.
- `git diff -- go.mod go.sum` passed with no output.
- `RUN-STATE.json` may now record `verificationPassed=true`; any later commit is artifacts-only.

## Verification plan

Focused after each slice:

```bash
gofmt -w cmd internal
go test -race ./internal/events/... -count=1
go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1
```

Final issue gate:

```bash
gofmt -w cmd internal
go test -race ./internal/events/... -count=1
go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1
go test -race ./...
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Dependency inspection:

```bash
go list -deps ./internal/events | grep -v '^polymetrics.ai/internal/events\|^polymetrics.ai/internal/safety\|^polymetrics.ai\|^[[:alnum:]_./-]*$'
```

CLI parity: N/A because no CLI command/flag/help/docs/website surface changes; `--progress` belongs to #405.

## Commit checkpoints

1. plan artifacts checkpoint.
2. red tests checkpoint if useful.
3. green events package.
4. green instrumentation.
5. final verification / PR body update.

## Spawn decision

`local_critical_path`: user invoked this bounded mutating worker directly in isolated cwd; worker has no `subagent` tool by contract. No recursive delegation.
