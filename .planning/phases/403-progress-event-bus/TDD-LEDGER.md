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
| 1 events package | `go test ./internal/events/... -count=1` | pending | pending | pending |
| 1 race | `go test -race ./internal/events/... -count=1` | pending | pending | pending |
| 2 flow sequence | `go test -race ./internal/flow/... -run TestEngineEmits -count=1` | pending | pending | pending |
| 3 app ETL sequence | `go test -race ./internal/app/... -run 'TestRunETLEmits|TestRunWarehouseETLEmits' -count=1` | pending | pending | pending |
| 4 certify sequence | `go test -race ./internal/connectors/certify/... -run TestRunBatchEmits -count=1` | pending | pending | pending |
| 5 worker poller | `go test -race ./internal/worker/... -run TestSubmitterEmits -count=1` | pending | pending | pending |
| final focused | `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | pending | pending | pending |
| final broad | issue verification commands | pending | pending | pending |

## Red test capture rule

Before production edits, add focused failing tests only. Capture exact command and failure output here before implementing each slice.
