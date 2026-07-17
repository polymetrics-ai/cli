# PLAN — Issue #404 redacted per-run slog foundation

## Contract

- Sub-issue: #404 `feat(obs): add redacted per-run slog foundation` under parent #397.
- Branch: `feat/404-redacted-slog`; base: `feat/cli-architecture-v2` at `20475ddf`.
- Scope: stdlib slog/redaction foundation, per-run JSONL routing/retention, vault.Get registry chokepoint, Temporal structured logger bridge, focused tests, issue-local planning artifacts, directly necessary non-CLI docs only.
- Out of scope: TTY/global flags/Cobra namespace changes, perf/telemetry/OTel/deps, connector bundles, parent state, CLI docs/help churn.
- Execution decision: `local_critical_path` — mutating worker already isolated in assigned cwd; no subagent tool available to this worker and no recursive delegation permitted.

## Required reading loaded

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/claude-review-loop.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` (read; CLI-visible docs/help changes not planned)
- #404 and #397 issue bodies/acceptance criteria via `gh issue view`
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`
- `docs/plans/cli-architecture-v2-improvement-plan.md` Stage 6/Pillar C
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md` Stage 6
- `docs/adr/0003-interactive-tui-layer.md` sibling layering
- `docs/adr/0004-opentelemetry-observability.md`
- Current seams: `internal/safety/safety.go`, `internal/vault/vault.go`, `internal/app/app.go`, `internal/events/events.go`, `internal/worker/{submit,serve}.go`, `internal/runtimecheck/runtimecheck.go`, `internal/temporalprobe/temporalprobe.go`, `internal/cli/{cli,cobra_router,parse,errors,worker_cli,runtime_helpers}.go`, CLI golden/contract tests.

## Skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-testing` — table tests, focused race gates, no order dependency.
- `golang-security` — secret-safe logging, filesystem traversal/symlink defense, no credentials.
- `golang-safety` — nil-safe handlers, resource close, defensive copies.
- `golang-observability` — stdlib `slog`, structured logs, stderr/stdout discipline.
- `golang-context` — run-id and logger context propagation.
- `golang-concurrency` — registry/handler/file cache locking, race tests.
- `golang-error-handling` — wrap errors, avoid logger-induced exit-code changes.
- `golang-design-patterns` — small constructors/options, explicit lifecycle close.
- `golang-structs-interfaces` — small handler/interfaces, concrete returns.
- `golang-documentation` — package docs/godoc for new logging package.
- `golang-cli` — stdout/stderr contract and CLI test seams.
- `caveman` — final handoff compression only.

## GSD adapter evidence

- `scripts/gsd doctor` => OK.
- `scripts/gsd prompt plan-phase 404 --skip-research` => prompt generated for `/gsd-plan-phase 404 --skip-research`.
- `scripts/gsd prompt programming-loop init --phase 404 --dry-run` => failed: `scripts/gsd: unknown GSD command: programming-loop`.
- Fallback: manual GSD universal loop from `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; record red/green/refactor evidence in this phase.

## Slice plan

### Review-fix cycle — accepted blockers from PR #455 adversarial/security pass

Execution decision: `local_critical_path` — existing isolated worker cwd, no recursive subagents. Findings treated as untrusted review input; coordinator-owned accepted blockers only.

TDD slices before production edits:

1. Redaction hardening: add red tests for inline/empty slog groups, sensitive group state, deferred bound-attr redaction, typed `url.URL`/`*url.URL` encoded-value scrubbing, and bounded invocation registry behavior.
2. Safe error boundary primitive: add red tests for context-scoped registry redaction across CLI JSON/stderr, app run state, events, logs, and connsdk HTTP errors without surfacing remote bodies.
3. Run-log filesystem hardening: add red tests for `.polymetrics` symlink rejection, `logs` symlink rejection, chmod/permission fail-closed behavior where hermetic, and retention preserving unrelated JSONL while pruning only handler-owned valid run logs.
4. Temporal probe lifecycle: add red tests proving `Probe` uses finite context-aware dial/health flow without an orphan timeout goroutine or logger retention.
5. Run correlation: add red tests proving `recordRuntimeETL` and Temporal structured logger routing bind validated run IDs into file routing even if the SDK adapter drops Handle context.
6. Plain diagnostic hygiene: add focused tests for single-line stderr diagnostics while keeping JSON escaped and error envelope taxonomy unchanged.

Implementation boundaries: `internal/logging`, `internal/safety`, `internal/events`, `internal/app`, `internal/cli`, `internal/runtimecheck`, `internal/temporalprobe`, `internal/worker`, `internal/connectors/connsdk`, and this phase's artifacts only. No parent artifact edits, no deps, no OTel/perf/TTY work, no services/real credentials.

Review-fix verification target is the user-specified command set. `verificationPassed` remains false until coordinator runs the deferred extended full CLI race; this worker will not run that extended race.

### Slice 0 — planning checkpoint

- Create issue-local phase artifacts: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `RUN-STATE.json`, `PROMPTS.md`.
- Commit/push planning checkpoint after diff check when feasible.

### Slice 1 — red tests: logging primitives

- Add failing tests for:
  - fixed sensitive-key redaction;
  - connector SecretFields/extra key redaction;
  - registered-value redaction without exposing registry values;
  - message/error/URL query/nested group sanitization;
  - warn+ fanout to provided stderr only;
  - run ID validation, traversal/control-char rejection, 0700/0600 perms, retention pruning, deterministic close.
- Expected red: package `internal/logging` missing.

### Slice 2 — green logging package

- Add `internal/logging` with:
  - `RedactingHandler` outer wrapper;
  - concurrency-safe fingerprint registry (no raw value storage/export);
  - context logger/run-id helpers;
  - run-file JSONL handler using `os.Root` to prevent traversal/symlink escape;
  - bounded retention and `Close` lifecycle;
  - level-filter and multi-handler fanout.
- Gate: `go test -race ./internal/logging/... -count=1`.

### Slice 3 — vault chokepoint and app run logs

- Add red test proving `vault.Get` registers retrieved values for redaction.
- Register values only in `vault.Get`.
- Add ETL run start/complete/fail info logs with `logging.WithRunID`; no per-record logs, no events-derived logs.
- Gate: `go test -race ./internal/vault/... ./internal/app/... -count=1`.

### Slice 4 — CLI wiring and smoke

- Wire CLI context logger with connector manifest SecretFields and provided stderr fanout.
- Add hermetic log smoke under temp project:
  - init project;
  - add synthetic canary credential from env;
  - run sample→warehouse ETL JSON;
  - assert exactly one JSON envelope shape on stdout;
  - assert log file non-empty;
  - assert canary absence;
  - assert check hook catches a deliberately dirty temp file.
- Gate: `go test -race ./internal/cli/... -count=1`.

### Slice 5 — Temporal bridge

- Replace applicable Temporal `noopLogger` seams with `tlog.NewStructuredLogger(*slog.Logger)` over context logger:
  - `internal/worker/submit.go`
  - `internal/worker/serve.go`
  - `internal/runtimecheck/runtimecheck.go`
  - `internal/temporalprobe/temporalprobe.go`
- Preserve no-service unit tests and no logger-induced exit-code changes.
- Gate: `go test -race ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... -count=1`.

### Slice 6 — full verification / PR

- Run issue verification commands.
- Confirm `git diff -- go.mod go.sum` empty.
- Open/update non-draft stacked PR to `feat/cli-architecture-v2` with `Refs #404` and `Refs #397`.
- Review route: Claude disabled/Copilot quota exhausted per assignment; record human/parent fallback pending, no requests.

## Safety notes

- No real credentials, no credentialed connector checks, no runtime/Temporal services.
- Synthetic canary remains test-only; do not echo it in handoff/output summaries.
- Logs never derive from events; events/ledger/logging remain siblings correlated by run ID.
- Logger failures must not alter command exit codes or stdout JSON envelope count.
