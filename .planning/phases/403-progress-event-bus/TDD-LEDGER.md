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
| final strict race | `go test -race ./... -count=1 -timeout 120m` | external PR-head source; no new code red phase | pass on head `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`: `PASS`; `internal/cli 1841.988s`; `internal/connectors/certify 3892.688s`; `internal/events 1.317s`; `real 3898.97`; `user 6294.91`; `sys 84.56` | baseline/worker suspect certify tests had nearly identical pass times, so prior 10m/30m timeouts were suite duration, not event regression |
| final non-race gates | `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`; diff checks | no new red phase; finalization/evidence slice only | pass: gofmt no production diff; vet no output; `go test ./...` `real 348.54` with `internal/connectors/certify 343.668s`; `make verify` `real 367.33` with `0 issues` and `547 connector(s) checked, 0 findings`; diff-check clean; go.mod/go.sum diff empty | verification complete; no production changes after strict race head |

## Red test capture rule

Before production edits, add focused failing tests only. Capture exact command and failure output here before implementing each slice.
