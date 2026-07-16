# PLAN — Issue 403 progress event bus

Sub-issue: https://github.com/polymetrics-ai/cli/issues/403
Parent: https://github.com/polymetrics-ai/cli/issues/397 / PR #438
Branch: `feat/403-progress-event-bus`
Base parent head: `e5ee4075`
Previous coordinator-confirmed race head: `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61` (invalidated by PR #451 review-fix production changes; strict full-race rerun pending with parent orchestrator)
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
   `RUN-STATE.json` now records `verificationPassed=false`; strict full-race rerun is pending with
   the parent orchestrator after final production head. This worker will not rerun the 65-minute full
   race gate.
4. Residual event sequence gaps: add focused terminal/error/skip/cancel tests only where needed to
   support issue #403 acceptance without broad fixture churn.

Review-fix TDD slices:

1. Add failing `internal/events` tests for Chan strict capacity/accounted lifecycle timeout,
   deterministic close-drop accounting, Multi+Chan finite fanout, and Throttle terminal ordering.
2. Add focused existing-instrumentation sequence tests for skipped flow steps, ETL failure terminal
   event, certify skip/error terminal paths, and worker successful terminal path if gaps are present.
3. Implement minimal events package fixes, keeping dependency-free stdlib + `internal/safety` only.
4. Re-run the requested focused race/non-race gates and update this phase's TDD/verification
   artifacts and PR body with accepted finding dispositions and pending strict full-race status.

Review-fix result:

- Chan/Throttle red test captured before production fix; green focused/race events tests passed.
- Focused flow/app/certify/worker sequence tests passed under `-race`.
- `go vet ./...`, focused non-race package tests, `go build ./cmd/pm`, `make verify`, diff-check,
  dependency inspection, and go.mod/go.sum check passed.
- Strict full-race remains pending parent orchestrator rerun; `verificationPassed=false` by design.

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
