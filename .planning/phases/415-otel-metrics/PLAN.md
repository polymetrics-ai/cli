# Phase 415 — OpenTelemetry metrics

Issue: #415 `feat(obs): add OpenTelemetry metrics`  
Parent: #397 / parent branch `feat/cli-architecture-v2`  
Worker branch: `feat/415-otel-metrics`  
Worker dir: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-415-otel-metrics`  
Base parent head: `56a7ecb08f755184af7b55318c3285582d5adfb7`

## GSD adapter and execution mode

- `scripts/gsd doctor` passed on 2026-07-18; adapter reported 69 commands.
- `scripts/gsd prompt plan-phase 415-otel-metrics --skip-research` generated `/tmp/gsd-plan-415.txt`.
- `scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`.
- Manual programming-loop fallback active via `.pi/prompts/pm-gsd-loop.md`; GSD/TDD order remains mandatory.
- Universal-loop execution decision for planning cycle: `local_critical_path` — this Pi worker owns one isolated issue worktree/branch and has no subagent tool.

## Required reading completed

- `AGENTS.md`.
- Issue #415 body and acceptance criteria via `gh issue view 415 --json ...`.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`.
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`.
- `.agents/agentic-delivery/workflows/claude-review-loop.md`.
- `.agents/agentic-delivery/references/required-skills-routing.md`.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`.
- `docs/plans/universal-programming-loop-prd.md`, `docs/prompts/universal-programming-loop-prompts.md`.
- `docs/plans/cli-architecture-v2-improvement-plan.md` Stage 17 / Pillar C.
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md` Stage 17.
- `docs/adr/0004-opentelemetry-observability.md`.
- Runtime canonical docs: `docs/architecture/runtime-dependencies.md`, `docs/runtime/SETUP.md`, `docs/cli/runtime.md`, `docs/cli/rlm.md`, `docs/cli/perf.md`, `docs/cli/agent.md`, `website/content/docs/architecture.mdx`, `website/content/docs/cli-reference.mdx`, `website/package.json`.
- Prior phase #410 artifacts: `.planning/phases/410-otel-tracing/{PLAN,TDD-LEDGER,VERIFICATION,SUMMARY}.md`.
- Existing #415 phase artifacts did not exist before this worker run; this directory is created for #415.

## Required skills loaded

- `gsd-core` — repo-local GSD adapter workflow.
- `golang-how-to` — Go skill routing.
- `golang-testing` — rules 1, 3, 5, 9: named tests, independent tests, observable behavior, fast unit tests.
- `golang-performance` — measure/avoid hot-loop allocations; no optimization without benchmark evidence.
- `golang-benchmark` — benchmark methodology and `-benchmem` evidence.
- `golang-observability` — rules 4, 5, 7-12: metrics/tracing, low cardinality, context correlation.
- `golang-context` — rules 1-8: propagate caller context; bounded shutdown with cancel.
- `golang-concurrency` — rules 1, 7: goroutine lifecycle and ctx cancellation for Temporal poller/exporters.
- `golang-security` — trust-boundary questions #1-#3; no bodies/headers/query/argv/credentials in metric attrs or warnings.
- `golang-error-handling` — rules 2, 7, 10, 14: wrap errors, single handling, neutral warnings.
- `golang-safety` — rules 2, 4, 6, 10: safe assertions, map init, defensive defaults.
- `golang-lint` — `go vet`, lint discipline, no broad suppressions.
- `golang-dependency-management` — ADR-only exact dependency budget and go.mod/go.sum review.
- `golang-cli` — stdout/stderr/exit-code/config docs parity.
- `golang-documentation` — CLI docs and website parity.
- `caveman` — final compact handoff only.

Note: `.pi/skills/go-implementation/SKILL.md` is absent in this checkout (`.pi/skills` contains only `gsd-core`); loaded repo/global Go skills above and will record absence in handoff. Website TS/UI skills are not loaded because planned website changes are MDX content only, no TS or UI component changes.

## Dependency constraints

ADR 0004 Stage 17 authorizes OTel metrics modules by signal/version line only: `sdk/metric`, `exporters`, and Temporal contrib on the OTel v1.44.0 / Temporal contrib v0.7.0 train. Allowed additions for this issue only if needed and pinned exactly:

- `go.opentelemetry.io/otel/sdk/metric@v1.44.0`.
- `go.opentelemetry.io/otel/exporters/stdout/stdoutmetric@v1.44.0`.
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp@v1.44.0`.
- `go.temporal.io/sdk/contrib/opentelemetry@v0.7.0`.

Any other new direct module or version deviation is a human gate. Existing tracing modules at v1.44.0 remain in scope. No `otelhttp`, no OTel log bridge, no Prometheus exporter, no grpc exporter promotion.

## Scope / write plan

Allowed production scope:

- `internal/telemetry/**`: metrics provider/exporters, local run counters, env sanitization, file/OTLP metric export, test helpers.
- `internal/app/**`: ETL batch-flush counter integration and allocation benchmark/test guard only.
- `internal/cli/**`: telemetry config/warnings, file-exported metric reconciliation contract test, embedded help/docs parity.
- `internal/worker/**`: Temporal tracing interceptor + metrics handler gated on telemetry enabled.
- Minimal docs parity: `docs/cli/config.md`, website CLI reference, generated website docs data if generator requires.
- `go.mod`/`go.sum` only for exact ADR-approved metrics/contrib modules.

Excluded:

- Shared parent orchestration ledgers/planning artifacts.
- Connector definition bundles under `internal/connectors/defs/**`.
- Dashboard/UI work for #408, OTel logs for #419, broad flow/app refactors beyond metrics call sites.
- Credentialed connector checks, runtime service startup, reverse ETL execution.

## Slice plan

### Slice 0 — red tests + dependency diff review

1. Add failing tests before production code:
   - `internal/telemetry`: file metrics exporter emits OTel metric data; disabled mode constructs no metric SDK/dir; OTLP metrics endpoint/env sanitization mirrors #410 tracing hardening.
   - `internal/app`: local `RunCounters` accumulates per-record counts and flushes only per batch; disabled/enabled hot-loop counter increments allocate 0 via `testing.AllocsPerRun`; `BenchmarkEmit` reports allocations.
   - `internal/cli`: `PM_TELEMETRY=file pm etl run --json` exports metrics that reconcile with `ETLRun.run.records_*` and `batch_count`; stdout remains one JSON envelope; warnings stay stderr.
   - `internal/worker`: Temporal client/worker options add tracing interceptor and metrics handler only when telemetry is enabled.
2. Run focused red command and record exact output in `TDD-LEDGER.md` before production edits.
3. Add only exact ADR-approved modules if production imports require them; record dependency diff.

### Slice 1 — telemetry metrics SDK/exporters

- Extend `internal/telemetry.Config`/`Handle` with meter provider/manual reader/exporter and metrics endpoint.
- Disabled/default path constructs no tracer or meter SDK and creates no telemetry directory.
- File mode writes metrics JSONL under `.polymetrics/telemetry` with restrictive permissions; keep trace files unchanged.
- OTLP mode uses HTTP/protobuf metric exporter only from trusted env/flag endpoint/default; unsupported `OTEL_EXPORTER_OTLP_*`, metrics-specific OTLP env, resource/provider/exporter/self-observability env are warned by name and sanitized before exporter/provider construction.
- Shutdown collects/exports metrics and traces within existing bounded timeout, warns without changing exit code.

### Slice 2 — batched ETL counters + benchmark guard

- Add `telemetry.RunCounters` with local integer accumulation (`RecordRead`, `RecordTransformed`, `RecordLoaded`, `RecordFailed`, `RecordBatch`) and `Flush(ctx)` that calls OTel instruments once per batch.
- Wire ETL source loops to increment local counters in hot paths and flush from existing batch flush closures; no OTel instrument calls per record.
- Final run completion values remain source of truth; metrics reconciliation test compares file metrics with final envelope counts.
- Add allocation guard and `BenchmarkEmit` under `internal/app`.

### Slice 3 — Temporal integration gated on telemetry enablement

- Add worker helper to derive Temporal `client.Options` / `worker.Options` with contrib tracing interceptor and metrics handler only when `telemetry.Enabled(ctx)`.
- Keep `dialTemporalClient` test seam; do not require runtime services.
- Ensure contrib `OnError` path warns/redacts rather than panics.

### Slice 4 — CLI/docs/website parity + full gates

- Update embedded config/root help text to say telemetry covers traces and metrics; file mode writes both; OTLP metrics endpoint env rules are trusted env/flag only; warnings exit-code neutral.
- Regenerate/update `docs/cli/config.md` and website docs data/source if needed.
- Verify runtime help parity (`./pm --help`, `./pm help config`, spot-check bare namespaces/invalid action), focused tests/benchmarks, full gates, dependency diff, `git diff --check`.

## Acceptance criteria mapping

| AC | Planned evidence |
|---|---|
| Counters accumulate locally and flush by batch | `internal/app`/`internal/telemetry` tests proving local counter increments and batch-only flush counts; allocation guard. |
| File-exported metrics reconcile with final envelope counts | CLI contract test parses `ETLRun` envelope and OTel metrics JSONL sums for records/batches. |
| Temporal tracing/metrics integration is gated on telemetry enablement | `internal/worker` unit test with dial seam: disabled options have no contrib handler/interceptor; enabled options do. |
| Benchmark guard shows no material emission-path allocation regression | `BenchmarkEmit` with `-benchmem` and an allocation assertion for hot-loop counter methods. |

## Commit/push checkpoints

1. Planning artifact checkpoint.
2. Red-test checkpoint with exact failing output recorded.
3. Green metrics SDK/exporter + batched counters checkpoint after focused gates.
4. Temporal integration + docs parity checkpoint after focused gates.
5. Full verification checkpoint and stacked PR to `feat/cli-architecture-v2`.

## PR plan

Open stacked PR to base `feat/cli-architecture-v2`, title `feat(obs): add OpenTelemetry metrics`, body includes `Refs #415`, `Refs #397`, GSD/TDD evidence, skills loaded, dependency justification, safety notes, parity checklist, gates, and review route status.
