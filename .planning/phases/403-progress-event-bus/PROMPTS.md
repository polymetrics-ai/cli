# PROMPTS — Issue 403 progress event bus

## Kickoff snapshot

Task: Execute polymetrics-ai/cli#403 as bounded mutating worker for parent #397.

Branch: `feat/403-progress-event-bus`
Parent PR: #438
Base: `feat/cli-architecture-v2` at `e5ee4075`
Write scope: `internal/events/**`, named instrumentation in `internal/flow/engine.go`, `internal/app/app.go`, `internal/app/local_warehouse.go`, `internal/connectors/certify/batch.go`, `internal/worker/submit.go`, focused tests, and this phase directory.

## GSD prompt commands

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 403 --skip-research
scripts/gsd prompt programming-loop init --phase 403 --dry-run
```

Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, scoped implementation, focused tests, and PR #451 review-fix updates.

Verification result: review-fix focused gates and `make verify` passed locally; strict full race is pending parent orchestrator rerun on final production head. Prior strict full-race evidence at `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61` is invalidated by accepted production fixes.

## Finalization snapshot — 2026-07-17

Task: finalize issue #403 after coordinator completed strict race gate on current branch head
`2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`.

External gate source: coordinator PR-head evidence, not this worker's self-generated final SHA.

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

Planned local final gates:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Downstream artifact: final gate update complete; commit/push and non-draft stacked PR to current
`feat/cli-architecture-v2` with `Refs #403` and `Refs #397` pending.

Verification result: superseded by PR #451 review-fix. Original non-race gates passed, but `verificationPassed=true` is invalidated by production changes; see review-fix snapshot.

## Review-fix snapshot — 2026-07-17

Task: fix accepted PR #451 findings for issue #403 at starting head `e5404809fc66296f6d02e243b09b431dade921fb`.

Accepted findings: Chan backpressure/close accounting, Throttle terminal ordering, invalidated full-race evidence, focused terminal/error/skip/cancel sequence gaps.

Downstream artifact: review-fix red tests, minimal events fix, focused sequence tests, updated phase artifacts and PR body.

Verification result: local focused gates and `make verify` passed. Strict full race (`go test -race ./... -count=1 -timeout 120m`) was not rerun by worker and remains pending with parent orchestrator for final production head; `RUN-STATE.json` records `verificationPassed=false`.

## Second targeted review-fix snapshot — 2026-07-17

Task: fix accepted PR #451 findings for issue #403 at starting head `c9813a788d2bc0ccc29e79920ce6e5e8084e8a8e`.

Accepted findings: `Chan.Close` must wait for runner acknowledgment so in-flight blocked events are accounted before `Close` returns and `Events()` is closed deterministically; `Multi` contract claims must be narrowed to synchronous fanout with bounded `Chan` sinks only.

Downstream artifact: red in-flight close regression, minimal `Chan` stopped acknowledgment, corrected Multi comments/tests/artifacts/PR body, requested local gates.

Verification result: requested focused gates and `make verify` passed locally; strict full race remains parent-orchestrator pending because production changed.

## Adapter gap

`programming-loop` is not in `.gsd/commands.json`; `scripts/gsd prompt programming-loop init --phase 403 --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`. Loaded `.pi/prompts/pm-gsd-loop.md` and running inline/manual universal loop.
