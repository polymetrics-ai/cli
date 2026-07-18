# TDD Ledger — Phase 415 OpenTelemetry metrics

## Skills loaded

- `gsd-core` — repo-local GSD adapter workflow.
- `golang-how-to` — Go skill routing.
- `golang-testing` — named/table tests, independent tests, behavior specs.
- `golang-performance` — hot-loop allocation discipline.
- `golang-benchmark` — benchmark methodology and `-benchmem` evidence.
- `golang-observability` — OTel metrics/tracing, low-cardinality attrs, signal correlation.
- `golang-context` — context propagation and bounded shutdown.
- `golang-concurrency` — goroutine/Temporal poller lifecycle.
- `golang-security` — trusted telemetry env boundaries; no secrets in metrics/warnings.
- `golang-error-handling` — wrapped errors and warning-only telemetry failures.
- `golang-safety` — nil/default/defensive-copy/resource safety.
- `golang-lint` — vet/lint gate discipline.
- `golang-dependency-management` — ADR-only exact dependency additions.
- `golang-cli` — stdout/stderr/exit-code/config docs parity.
- `golang-documentation` — concise CLI docs and website parity.
- `caveman` — final compact handoff only.

Missing repo stack skill note: `.pi/skills/go-implementation/SKILL.md` is absent; `.pi/skills` contains only `gsd-core`.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 415-otel-metrics --skip-research >/tmp/gsd-plan-415.txt
scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415.txt
```

Result:

- `doctor`: pass, 69 commands.
- `plan-phase`: generated `/tmp/gsd-plan-415.txt`.
- `programming-loop`: failed with `scripts/gsd: unknown GSD command: programming-loop`; manual fallback to `.pi/prompts/pm-gsd-loop.md` active and recorded.

## Planned red-test requirements

Required failing evidence before production edits:

- File metrics exporter absent/failing: no OTel metric JSONL for ETL counts yet.
- Batched counter API absent/failing: local counters cannot yet flush OTel metrics by batch and allocation guard cannot pass.
- CLI reconciliation absent/failing: `PM_TELEMETRY=file pm etl run --json` cannot reconcile metrics to envelope counts.
- Temporal contrib integration absent/failing: disabled/enabled worker options do not yet expose gated tracing interceptor + metrics handler behavior.

Use synthetic non-secret markers only. Never use real credentials.

## Ledger

| # | Cycle | Type | Command / evidence | Result | Notes |
|---:|---|---|---|---|---|
| 1 | plan | Planning | Create `.planning/phases/415-otel-metrics/{PLAN,TDD-LEDGER,VERIFICATION,SUMMARY,PROMPTS,RUN-STATE}.md/json` before production edits | Pass | Existing phase artifacts were absent. Execution decision: `local_critical_path`. |
