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

Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, scoped implementation, and focused tests pushed; finalization evidence update in progress.

Verification result: strict full race passed by coordinator external PR-head source at `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`; remaining non-race final gates passed locally. Prior focused `internal/connectors/certify` 10m/30m timeouts are recorded as suite-duration evidence, not event regression.

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

Verification result: passed. `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`,
`go build ./cmd/pm`, `make verify`, `git diff --check origin/feat/cli-architecture-v2...HEAD`,
and `git diff -- go.mod go.sum` all passed; `verificationPassed=true`.

## Adapter gap

`programming-loop` is not in `.gsd/commands.json`; `scripts/gsd prompt programming-loop init --phase 403 --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`. Loaded `.pi/prompts/pm-gsd-loop.md` and running inline/manual universal loop.
