# TDD LEDGER — Issue 403 progress event bus

## Loaded skills

- `gsd-core` — repo-local GSD adapter workflow.
- `caveman` — compact handoff only.
- `golang-how-to` — Go skill router.
- `golang-testing` — red/green tests, race gates.
- `golang-context` — context-carried emitter and worker cancellation.
- `golang-concurrency` — Chan sink, Multi/Throttle races, worker poller goroutines.
- `golang-security` — NDJSON sanitization/redaction; no secrets in events.
- `golang-safety` — nil/default emitters, defensive copies, zero-value behavior.
- `golang-design-patterns` — small dependency-free sinks and lifecycle boundaries.
- `golang-structs-interfaces` — small `Emitter` interface, typed `Event` struct.
- `golang-error-handling` — wrapped errors, no swallowed setup failures.
- `golang-lint` — `go vet`/quality-gate checks and race-safety review.
- `golang-code-style` — minimal, clear concurrency edits.
- `golang-documentation` — corrected sink contract comments and PR/artifact claims.
- `golang-naming` — precise test/comment naming for bounded Chan vs synchronous Multi.
- `golang-troubleshooting` — root-cause analysis for in-flight close nondeterminism.

Stack implementation skill note: `.pi/skills/go-implementation/SKILL.md` was requested by worker
instructions but is absent in this checkout (`ENOENT`); loaded `gsd-core` plus the required global
Go skills from `.agents/agentic-delivery/references/required-skills-routing.md` instead.

## GSD command evidence

```bash
scripts/gsd doctor
```

Result: pass.

```bash
scripts/gsd prompt plan-phase 403 --skip-research
```

Result: generated official `/gsd-plan-phase 403 --skip-research` prompt.

```bash
scripts/gsd prompt programming-loop init --phase 403 --dry-run
```

Result: fail, adapter gap: `scripts/gsd: unknown GSD command: programming-loop`.

Fallback: `.pi/prompts/pm-gsd-loop.md` loaded and executed inline/manual; decision `local_critical_path`.

## Red / Green ledger

| Slice | Test / validation | Red evidence | Green evidence | Refactor evidence |
|---|---|---|---|---|
| 1 events package | `go test ./internal/events/... -count=1` | fail (build): undefined `FromContext`, `Event`, `KindStarted`, `ScopeFlow`, `NewCollector`, `WithEmitter`, `Emit` | pass: `ok   polymetrics.ai/internal/events 0.340s` | `gofmt -w internal/events` |
| 1 race | `go test -race ./internal/events/... -count=1` | pending until package builds | pass: `ok   polymetrics.ai/internal/events 1.179s` | no refactor beyond gofmt |
| 2 flow sequence | `go test -race ./internal/flow/... -run TestEngineEmits -count=1` | fail: collector sequence length 0, want flow start/step start/step completed/flow completed; failure path also length 0 | pass: `ok   polymetrics.ai/internal/flow 1.437s` | `gofmt -w internal/flow` |
| 3 app ETL sequence | `go test -race ./internal/app/... -run 'TestRunETLEmits|TestRunWarehouseETLEmits' -count=1` | fail: collector sequence length 0, want ETL start/batch progress/completed for connector + warehouse flush paths | pass: `ok   polymetrics.ai/internal/app 18.027s` | `gofmt -w internal/app` |
| 4 certify sequence | `go test -race ./internal/connectors/certify/... -run TestRunBatchEmits -count=1` | fail: collector sequence length 0, want certify batch/connector lifecycle | pass: `ok   polymetrics.ai/internal/connectors/certify 1.632s` | `gofmt -w internal/connectors/certify` |
| 5 worker poller | `go test -race ./internal/worker/... -run TestSubmitterEmits -count=1` | fail (build): undefined `workflowPollInterval`, `submitterForWorkflowClient`, `workflowRun` | pass: `ok   polymetrics.ai/internal/worker 1.351s` | `gofmt -w internal/worker` |
| final focused | `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | pending | fail: exact command hit Go test default `panic: test timed out after 10m0s` in `internal/connectors/certify` after flow/app passed; supplemental `-timeout 30m` also timed out in existing certify source-stage tests | superseded by strict full-race external gate below |
| prior strict race (invalidated) | `go test -race ./... -count=1 -timeout 120m` | external PR-head source before review-fix production changes | invalidated by accepted PR #451 fixes; do not claim coverage for head `e5404809fc66296f6d02e243b09b431dade921fb` or later | parent orchestrator owns final strict full-race rerun after final production head |
| review-fix events red | `go test ./internal/events/... -run 'TestChan|TestThrottle' -count=1` | fail (build): `internal/events/events_test.go:108:19: sink.DropStats undefined`; `:132:19`; `:146:16`; package build failed | pass: `ok   polymetrics.ai/internal/events 0.385s`; race pass: `ok   polymetrics.ai/internal/events 1.388s` | `gofmt -w internal/events` |
| review-fix sequence tests | focused `go test` commands under `internal/flow`, `internal/app`, `internal/connectors/certify`, `internal/worker` | characterization/focused gap tests added for flow skip, ETL failure terminal, certify skip/failure, worker success; worker first build failed due test double type (`cannot use *successfulWorkflowRun as *fakeWorkflowRun`) and was fixed before production sequence changes | pass: flow `0.303s`, app `2.991s`, certify `0.532s`, worker `0.297s`; race pass: flow `1.291s`, app `29.128s`, certify `1.639s`, worker `1.331s` | no production sequence changes needed |
| review-fix focused gates | requested focused race/non-race gates from PR #451 review-fix task | events red above | pass: `go vet ./...` no output; package test gate passed (`events 0.488s`, `flow 0.506s`, `app 18.817s`, `certify 340.499s`, `worker 0.439s`); `go build ./cmd/pm` no output; `make verify` passed with `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`; diff checks clean; go.mod/go.sum diff empty | strict full-race intentionally not rerun by worker; parent orchestrator pending |

## Review-fix accepted findings

- HIGH Chan backpressure/close semantics accepted: queue must stay within capacity; lifecycle delivery under full lifecycle queues must bounded-wait with context/close handling and explicit drop accounting; Close uses explicit accounted close-drop semantics; Multi must not block indefinitely on Chan.
- MEDIUM Throttle terminal ordering accepted: latest pending progress must flush before completed/failed/skipped terminal lifecycle events; terminal remains last.
- Race evidence accepted: prior strict full-race pass is invalidated by production fixes. `verificationPassed=false` until parent orchestrator reruns strict full race on final production head.
- Residual sequence gaps accepted for focused tests only; no broad fixture churn.

## Second targeted review-fix accepted findings

- MEDIUM Chan Close in-flight acknowledgment accepted: add red regression for a lifecycle event
  removed from the queue and blocked on `Events()` with no consumer; `Close` must wait for runner
  shutdown so immediate `DropStats()` and `Events()` closure are deterministic.
- LOW Multi contract correction accepted: `Multi` remains synchronous; only bounded `Chan` sinks have
  finite backpressure/close semantics. Custom sinks must be finite or honor context cancellation.

Red command before production edits:

```bash
go test ./internal/events/... -run TestChanCloseWaitsForInFlightEventAccounting -count=1
```

Red evidence at starting head `c9813a788d2bc0ccc29e79920ce6e5e8084e8a8e`:

```text
--- FAIL: TestChanCloseWaitsForInFlightEventAccounting (0.00s)
    events_test.go:194: DropStats() after Close = {Progress:1 Lifecycle:0}, want {Progress:1 Lifecycle:1}
FAIL
FAIL	polymetrics.ai/internal/events	0.382s
FAIL
```

Green evidence:

```bash
gofmt -w internal/events
go test ./internal/events/... -run TestChanCloseWaitsForInFlightEventAccounting -count=1
```

Result: `ok   polymetrics.ai/internal/events   0.430s`.

Focused race evidence: `go test -race ./internal/events/... -count=1` passed with
`ok   polymetrics.ai/internal/events   1.279s` on the final local gate rerun.

Close semantics after fix: `Close` sets `closed`, accounts queued events, closes `done`, signals the
runner, waits for `stopped`; the runner accounts any event removed from the queue and blocked on
`out`, closes `Events()`, then closes `stopped`. The regression asserts exact drops
`{Progress:1 Lifecycle:1}` for one queued progress event plus one in-flight lifecycle event.

Multi contract correction: `Multi` remains synchronous. The finite fanout test is explicitly scoped
as `TestMultiWithBoundedChanDoesNotBlockIndefinitelyWhenChanLifecycleQueueStalls`; arbitrary custom
or writer sinks must be finite or observe context cancellation.

## Red test capture rule

Before production edits, add focused failing tests only. Capture exact command and failure output here before implementing each slice.
