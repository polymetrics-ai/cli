# PLAN — Issue 403 progress event bus

Sub-issue: https://github.com/polymetrics-ai/cli/issues/403
Parent: https://github.com/polymetrics-ai/cli/issues/397 / PR #438
Branch: `feat/403-progress-event-bus`
Base parent head: `e5ee4075`
Coordinator-confirmed race head: `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`
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

## Finalization plan — 2026-07-17

Coordinator completed the strict full race gate externally on branch head
`2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`. This worker will not rebase, chase a self-generated
SHA, or make production changes unless a real remaining gate failure requires it. The finalization
slice is documentation/evidence plus the requested non-race final gates, then push and open the
stacked PR to the current `feat/cli-architecture-v2` base.

External strict race evidence to carry forward honestly:

```text
go test -race ./... -count=1 -timeout 120m
PASS
internal/cli 1841.988s
internal/connectors/certify 3892.688s
internal/events 1.317s
real 3898.97
user 6294.91
sys 84.56
```

Baseline/worker suspect certify tests had nearly identical pass times, confirming prior 10m/30m
timeouts were suite duration, not an event-regression signal.

Remaining final gates passed locally after this evidence update: gofmt had no production diff; vet and
build had no output; `go test ./...` passed with `internal/connectors/certify 343.668s`; `make verify`
passed with `0 issues` and `connectorgen validate: 547 connector(s) checked, 0 findings`; diff-check
was clean; `go.mod`/`go.sum` diff was empty. `verificationPassed=true` is now recorded in
`RUN-STATE.json`.

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
