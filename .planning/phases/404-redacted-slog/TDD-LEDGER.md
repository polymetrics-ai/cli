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

Review-fix note: synthetic non-secret markers may appear in test fixtures only; handoff/PR/artifact summaries must not print marker values.

| ID | Slice | Test/validation | Red evidence | Green evidence | Refactor/gate |
|---|---|---|---|---|---|
| T0 | Planning | Phase artifacts created before production edits | n/a | This ledger + PLAN/VERIFICATION created | pending commit |
| T1 | Logging primitives | `go test ./internal/logging/... -run 'TestRedactingHandler|TestRunFileHandler|TestRunLogger|TestTemporalStructuredLogger' -count=1` | RED — build failed: `undefined: NewValueRegistry`, `undefined: NewRedactingHandler`, `undefined: RedactionOptions`, `undefined: NewLogger`, `undefined: LoggerOptions`, `undefined: WithRunID`, `undefined: NewRunFileHandler`, `undefined: RunFileOptions` | GREEN — `ok polymetrics.ai/internal/logging 0.441s` | pending focused race |
| T2 | Vault registry | `go test ./internal/vault/... -run TestVaultGetRegistersValuesForRedaction -count=1` | RED — build failed: `polymetrics.ai/internal/logging: no non-test Go files` | GREEN — `ok polymetrics.ai/internal/vault 0.747s` | pending focused race |
| T3 | CLI log smoke | `go test ./internal/cli/... -run TestRedactedRunLogsSmoke -count=1` | RED — `expected at least one run log` | GREEN — `ok polymetrics.ai/internal/cli 3.052s` | pending focused race |
| T4 | Temporal bridge | `go test ./internal/logging/... -run TestTemporalStructuredLoggerUsesContextRedactingLogger -count=1`; `go test ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... -count=1` | RED covered by T1 missing logging package API and pre-existing `noopLogger` seams | GREEN — worker/runtimecheck/temporalprobe `ok` with no service required | pending focused race |
| T5 | Focused race gate | `go test -race ./internal/logging/... ./internal/vault/... ./internal/worker/... ./internal/runtimecheck/... ./internal/cli/... -count=1` | n/a after green | PARTIAL — logging/vault/worker/runtimecheck/temporalprobe race passed; exact issue command timed out because `./internal/cli` exceeds Go test timeout while repeatedly loading connector bundles (`TestCertifyCLISingleConnectorSavesReport`, then `TestDocsGenerateAndValidateConnectorDocs`) | BLOCKER: issue race command not green |
| T6 | Full gate | `go vet ./... && go test ./... && go build ./cmd/pm && make verify` | n/a after green | GREEN — `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` all exited 0 | `make verify` includes `go test -timeout 20m ./...`, docs validate, smoke incl reverse preview/run, lint, connectorgen validate |
| T7 | Review-fix planning | Accepted PR #455 blockers converted into TDD slices before production edits | n/a | PLAN/VERIFICATION/RUN-STATE/PROMPTS updated for review-fix cycle | production edits pending |
| T8 | Review-fix redaction hardening | `go test ./internal/logging/... ./internal/app/... ./internal/cli/... ./internal/connectors/connsdk/... ./internal/temporalprobe/... -run 'TestRedactingHandler|TestValueRegistry|TestRunFileHandler|TestTemporalStructuredLogger|TestRunETLFailureRedacts|TestRedactedErrorOutputSingleLineSmoke|TestHTTPError|TestRequesterReturnsHTTPErrorOn404|TestProbeUses' -count=1` | RED — build/test failures: `undefined: NewValueRegistryWithLimit`, `undefined: pmlogging.WithRegistry`, `undefined: dialTemporal`, CLI stderr/stdout redaction failed, connsdk surfaced response body/userinfo, logging red tests not yet buildable | GREEN — `ok` for logging/app/cli/connsdk/temporalprobe focused packages after implementation | gofmt run on touched packages |
| T9 | Review-fix safe error boundaries | same focused command as T8 | RED — app/CLI/connsdk tests fail before safe context registry + output redaction implementation | GREEN — app state/events/logs and CLI stdout/stderr focused tests clean; connsdk no longer surfaces response body | pending requested gates |
| T10 | Review-fix Temporal/runtime correlation | same focused command as T8 plus later runtime/worker focused tests | RED — temporalprobe bounded dial seam missing and bound run-ID routing test not yet implemented | GREEN — temporalprobe bounded dial tests and Temporal bound run-id routing test pass | requested gates passed except extended CLI race deferred |
| T11 | Review-fix requested verification | Coordinator command set in VERIFICATION.md | n/a after green | GREEN — requested gofmt, focused race, CLI focused, vet, all tests, build, make verify, diff-check, go.mod/go.sum diff all exited 0 | extended full CLI race not run per instruction |

## Canary handling rule

Tests may use a clearly synthetic non-secret canary fixture to prove redaction. Test failure messages, phase summaries, PR body, and handoff must not print the fixture value.
