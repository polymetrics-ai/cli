# TDD LEDGER — Issue #404

## Loaded skills

`gsd-core`, `golang-how-to`, `golang-testing`, `golang-security`, `golang-safety`, `golang-observability`, `golang-context`, `golang-concurrency`, `golang-error-handling`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-cli`, `caveman`.

Routing notes from `.agents/agentic-delivery/references/required-skills-routing.md`:

- Go work starts with `golang-how-to`.
- Runtime/Temporal work also loads context/concurrency/security/safety/testing/documentation.
- CLI stdout/stderr seam work loads `golang-cli`; no CLI-visible docs/help change planned.

## GSD command evidence

| Step | Command | Result |
|---|---|---|
| Doctor | `scripts/gsd doctor` | PASS |
| Plan prompt | `scripts/gsd prompt plan-phase 404 --skip-research` | PASS (prompt generated) |
| Programming loop dry-run | `scripts/gsd prompt programming-loop init --phase 404 --dry-run` | FAIL — `scripts/gsd: unknown GSD command: programming-loop` |

Fallback: manual GSD universal programming loop. Execution decision for this worker cycle: `local_critical_path`.

## Red/green/refactor entries

| ID | Slice | Test/validation | Red evidence | Green evidence | Refactor/gate |
|---|---|---|---|---|---|
| T0 | Planning | Phase artifacts created before production edits | n/a | This ledger + PLAN/VERIFICATION created | pending commit |
| T1 | Logging primitives | `go test ./internal/logging/... -run 'TestRedactingHandler|TestRunFileHandler|TestRunLogger|TestTemporalLogger' -count=1` | RED — build failed: `undefined: NewValueRegistry`, `undefined: NewRedactingHandler`, `undefined: RedactionOptions`, `undefined: NewLogger`, `undefined: LoggerOptions`, `undefined: WithRunID`, `undefined: NewRunFileHandler`, `undefined: RunFileOptions` | pending | pending |
| T2 | Vault registry | `go test ./internal/vault/... -run TestVaultGetRegistersValuesForRedaction -count=1` | RED — build failed: `polymetrics.ai/internal/logging: no non-test Go files` | pending | pending |
| T3 | CLI log smoke | `go test ./internal/cli/... -run TestRedactedRunLogsSmoke -count=1` | RED — `expected at least one run log` | pending | pending |
| T4 | Temporal bridge | `go test ./internal/logging/... -run TestTemporalLoggerUsesContextRedactingLogger -count=1` | RED covered by T1 missing `TemporalLogger`/logging package API | pending | pending |
| T5 | Focused race gate | `go test -race ./internal/logging/... ./internal/vault/... ./internal/worker/... ./internal/runtimecheck/... ./internal/cli/... -count=1` | pending | pending | pending |
| T6 | Full gate | `go vet ./... && go test ./... && go build ./cmd/pm && make verify` | pending | pending | pending |

## Canary handling rule

Tests may use a clearly synthetic non-secret canary fixture to prove redaction. Test failure messages, phase summaries, PR body, and handoff must not print the fixture value.
