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

Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, scoped implementation, and focused tests pushed.

Verification result: blocked. Focused per-slice tests pass, but `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` timed out in `internal/connectors/certify`; supplemental `-timeout 30m` also timed out.

## Adapter gap

`programming-loop` is not in `.gsd/commands.json`; `scripts/gsd prompt programming-loop init --phase 403 --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`. Loaded `.pi/prompts/pm-gsd-loop.md` and running inline/manual universal loop.
