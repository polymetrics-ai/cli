# Phase 410 — Opt-in OpenTelemetry tracing

Issue: #410 `feat(obs): add opt-in OpenTelemetry tracing`  
Parent: #397 / parent PR #438 (`feat/cli-architecture-v2`, draft, human-gated)  
Branch: `feat/410-otel-tracing`  
Worker dir: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-410-otel-tracing`

## GSD adapter and execution mode

- `scripts/gsd doctor` passed on 2026-07-18.
- `scripts/gsd prompt plan-phase 410 --skip-research` generated a 142-line prompt in `/tmp/gsd-plan-phase-410.txt`.
- `scripts/gsd prompt programming-loop init --phase 410 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD/TDD fallback is active for the programming-loop step only.
- Universal loop decision for planning cycle: `local_critical_path` — this Pi worker has no subagent tool and owns one isolated issue worktree.

## Required reading completed

- `AGENTS.md`.
- Issue #410 body and acceptance criteria.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`.
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- `.agents/agentic-delivery/references/required-skills-routing.md`.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.
- `docs/plans/cli-architecture-v2-improvement-plan.md` Pillar C / Stage 12.
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md` Stage 12.
- `docs/adr/0004-opentelemetry-observability.md`; ADR 0002/0003 checked for dependency and sibling-layer constraints.
- Existing phase artifacts did not exist before this worker run; this directory is created for #410.

## Required skills loaded

- `gsd-core`.
- `golang-how-to`.
- `golang-testing` (rules 1, 3, 5, 9: named/table tests, independent, behavior not implementation, fast unit tests).
- `golang-security` (trust-boundary questions #1-#3; no bodies/headers/query/argv/secret values; path and env inputs treated as untrusted).
- `golang-safety` (rules 2, 4, 6, 10: safe assertions, map initialization, defensive copies, useful defaults).
- `golang-observability` (rules 7-12: OTel setup, meaningful spans, context propagation, no high-cardinality unsafe attrs).
- `golang-context` (rules 1-8: propagate caller context, no mid-path Background, bounded shutdown context).
- `golang-concurrency` (rules 1, 7: shutdown/ctx awareness; no goroutine leaks in exporters/shutdown).
- `golang-error-handling` (rules 2, 7, 10, 14: wrapped errors, single handling, structured warnings on stderr, exit-code neutrality).
- `golang-cli` (stdout/stderr discipline, exit code stability, flag/config/help parity).
- `golang-documentation` (CLI docs/help/website parity; concise safety docs).

## Dependency constraints

Approved by ADR 0004 Stage 12 only:

- `go.opentelemetry.io/otel@v1.44.0`
- `go.opentelemetry.io/otel/sdk@v1.44.0`
- `go.opentelemetry.io/otel/exporters/stdout/stdouttrace@v1.44.0`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.44.0`

No `otelhttp`, no grpc exporter promotion, no metrics SDK, no otel log bridge, no Temporal OTel contrib in this issue. Any version/module deviation is a human gate.

## Scope / write plan

Allowed production scope:

- New `internal/telemetry/**` package.
- Narrow config/env/help/docs additions for telemetry enablement.
- Command seam instrumentation in `internal/cli/**`.
- Operation instrumentation in `internal/app/**`, `internal/flow/**`, `internal/connectors/certify/**` or certify CLI wrapper as needed.
- HTTP chokepoint instrumentation in `internal/connectors/connsdk/http.go`.
- Tests under touched packages plus docs/website parity updates.

Excluded:

- `.planning/traces/cli-architecture-v2-orchestration-state.yaml` and other shared parent orchestration artifacts.
- Later phases (#415 metrics, #419 log bridge), TUI work, connector defs, broad generated rewrites unrelated to docs parity.
- Credentialed connector checks and runtime service startup.

## Slice plan

### Slice 0 — red tests + dependency gate

1. Add failing tests before production code:
   - `internal/telemetry`: disabled mode no SDK/dir; file exporter smoke; attribute allowlist and secret absence; init/exporter/shutdown warn-and-neutral behavior via test hooks.
   - `internal/connectors/connsdk`: HTTP span records only scheme/host/path/status/attempt attrs and omits query/header/body/token.
   - `internal/cli` or integration-level tests: command span wraps expected commands and keeps stdout envelope-only; telemetry init/shutdown failures do not change exit codes.
   - `internal/app` / `internal/flow` / certify wrapper: ETL, flow step, and certify spans when enabled.
2. Run focused tests and record exact red output in `TDD-LEDGER.md` before production edits.
3. Add only ADR-approved OTel modules/version lines.

### Slice 1 — `internal/telemetry` core

- Implement `Config`, `Init(ctx,cfg,warn)`, disabled short-circuit with no SDK and no directory creation, file exporter under `<root>/.polymetrics/telemetry/<run-id-or-timestamp>.jsonl`, OTLP http/protobuf endpoint, `OTEL_SDK_DISABLED` override, bounded shutdown (3s in CLI, injectable in tests).
- Provide context helpers: `StartSpan`, `SpanFromContext` fallback/nop behavior, `SetAttributes` through an allowlist, and HTTP attribute builders that drop query/userinfo/fragment/body/header/argv.
- Warning callback writes sanitized warnings to stderr; never returns errors that alter command exit.

### Slice 2 — command and operation instrumentation

- Wire config/env keys into `internal/config`: exporter mode default `none`/`off`, endpoint, directory, capture mode.
- In `RunWithOptions`, initialize telemetry after config/logging setup; start `pm.command`; defer bounded shutdown warning.
- Start `pm.etl.run` around app ETL run; `pm.flow.run` / `pm.flow.step` in flow engine; `pm.certify.connector` / `pm.certify.batch` in certify paths; keep action approval semantics untouched.

### Slice 3 — connector HTTP instrumentation

- Instrument `connsdk.Requester.do` as one `pm.connector.http` logical-request span with per-attempt events.
- Attributes allowlist only: `pm.http.scheme`, `pm.http.host`, `pm.http.path`, `pm.http.method`, status/attempt counts/retry booleans; no request/response bodies, headers, query strings, full URLs, argv, or credential values.

### Slice 4 — CLI parity, docs, smoke gates

- Update embedded help (`rootHelp`/`configHelp`), generated `docs/cli/config.md` (and root docs if generator changes), website `website/content/docs/cli-reference.mdx`, and generated website data if needed.
- Verify `pm help config`, `pm --help`, relevant bare namespaces unchanged, docs generate diff, website docs generator.
- Run file exporter smoke, off-mode check, secret grep, envelope-only stdout, `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify` when feasible.

## Acceptance criteria mapping

| AC | Plan evidence |
|---|---|
| Disabled mode constructs no SDK and creates no telemetry directory | `internal/telemetry` disabled test + CLI off-mode smoke |
| Expected command, ETL/flow/certify, connector HTTP spans emitted when enabled | file exporter smoke and focused package tests |
| HTTP attributes only allowlisted scheme/host/path metadata | connsdk span attr test + exported-span walker |
| Exporter/init/shutdown failures warn without changing exit codes | CLI tests with failing exporter/shutdown hooks |
| Secret and attribute allowlist tests pass | test token absence grep + allowlist walker |

## Commit/push checkpoints

1. Planning artifacts checkpoint.
2. Red-test checkpoint if tests can compile without dependencies; otherwise dependency + red tests checkpoint with failing output recorded.
3. Green telemetry core + focused tests.
4. Green instrumentation + docs parity.
5. Full verification + PR checkpoint.

## PR plan

Open stacked PR to base `feat/cli-architecture-v2`, title `feat(obs): add opt-in OpenTelemetry tracing`, body includes `Refs #410` and `Refs #397`, GSD/TDD evidence, dependency evidence, gates, skill list, parity checklist, safety notes, and review route status.
