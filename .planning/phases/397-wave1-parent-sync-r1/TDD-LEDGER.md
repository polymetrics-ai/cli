# Issue #397 Wave 1 TDD Ledger

Status: active
Manual GSD fallback: PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE (`programming-loop` is absent from the 69-command adapter registry).

## Planned RED / regression evidence

| Conflict / risk | RED or regression contract | GREEN target | Status |
|---|---|---|---|
| `internal/cli/cli.go` | Combined Gong dynamic dispatch through the Cobra root plus native/help/golden/certify/reverse contracts | Gong routes and Architecture v2 routing/config/events survive together | pending |
| `internal/connectors/connectors.go` | Gong operation/direct-read/write discovery plus existing connector contracts | Main Gong capabilities remain available without dropping parent integration behavior | pending |
| `internal/connectors/connsdk/http.go` | JSON arrays, multipart, query handling, retries and parent telemetry/redaction tests | Payload behavior and instrumentation coexist | pending |
| `go.mod` / `go.sum` | module verify/tidy and build expose dropped requirements or inconsistent sums | union graph verifies and tidy diff is empty | pending |
| auto-merges | focused CLI/app/connsdk suites detect semantic loss outside conflict markers | all focused suites pass | pending |
| machine/reverse safety | golden transcript and reverse native tests | stdout/stderr/JSON/exit contracts and plan → preview → approval → execute remain unchanged | pending |

## Evidence log

- 2026-07-23: `git merge-tree --write-tree origin/feat/cli-architecture-v2 origin/main` reported conflicts in `go.mod`, `go.sum`, `internal/cli/cli.go`, `internal/connectors/connectors.go`, and `internal/connectors/connsdk/http.go`; related `internal/app/app.go` and `internal/connectors/connsdk/http_test.go` auto-merged.
- No production edits or merge resolution occurred before this ledger and the plan/checklist were created.
