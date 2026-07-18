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
# continuation after truncated prior worker output
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415-continuation.txt
```

Result:

- `doctor`: pass, 69 commands (initial and continuation).
- `list`: pass on continuation; command registry visible.
- `plan-phase`: generated `/tmp/gsd-plan-415.txt`.
- `programming-loop`: failed with `scripts/gsd: unknown GSD command: programming-loop` initially and on continuation; manual fallback to `.pi/prompts/pm-gsd-loop.md` active and recorded.
- Continuation preserved dirty implementation worktree; no reset/discard/recreate.

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
| 2 | red | Test | `go test ./internal/app ./internal/cli ./internal/worker -run 'Metric|Telemetry|Temporal' -count=1` | Fail | Expected red before production edits. Key output: `internal/worker/telemetry_test.go:17:14: undefined: temporalClientOptions`; `internal/app/metrics_test.go:72:24: undefined: telemetry.NewRunCounters`; CLI reconciliation failed with `metric pm.records.read sum = 0, want 3` and only trace spans present; metrics endpoint hardening test failed with `stderr missing redacted OTLP metrics endpoint warning: "warning: telemetry: telemetry export failed\n"`. |
| 3 | continuation | Planning | `git status --short --branch`; `scripts/gsd doctor`; `scripts/gsd list`; `scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415-continuation.txt` | Pass/fallback | Dirty worktree preserved after prior output truncation. GSD doctor/list pass; programming-loop command still unavailable, manual-GSD fallback remains active. Execution decision: `local_critical_path`. |
| 4 | green | Test | `go test ./internal/telemetry ./internal/app ./internal/cli ./internal/worker -run 'Metric|Telemetry|Temporal|Golden|Config|RunCounter' -count=1` | Pass | Metrics SDK/exporters, batched counters, CLI reconciliation, env hardening, docs goldens, and Temporal telemetry gating green. |
| 5 | green | Benchmark | `go test -bench BenchmarkEmit -benchmem ./internal/app` | Pass | `BenchmarkEmit-12 592256128 2.040 ns/op 0 B/op 0 allocs/op`; enabled/disabled hot-path allocation tests also assert 0 allocs. |
| 6 | verify | Gate | `gofmt -w cmd internal`; `go vet ./...`; `go test -timeout 20m ./...`; `go build ./cmd/pm`; `git diff --check` | Pass | Full package suite green; build and whitespace gates green. |
| 7 | verify | Gate | `make verify` before dependency commit | Fail/expected | Stopped at `tidy-check` because ADR-approved go.mod/go.sum dependency delta was intentionally uncommitted; rerun after commit. |
