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
- `vercel-react-best-practices` — loaded for website generated-data parity after CI required `website/lib/docs.generated.ts`; no React code changed.
- `vercel-composition-patterns` — loaded for website TS awareness; no component API/composition changes.
- `caveman` — final compact handoff only.

Missing repo stack skill note: `.pi/skills/go-implementation/SKILL.md` and `.pi/skills/ts-website/SKILL.md` are absent; `.pi/skills` contains only `gsd-core`.

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

## Review-fix red-test plan — PR #461

Starting head: `8748a03ba60042bdc29bd9cce1acf7c3d0b286a3`. User-directed review route: no Claude/Copilot. Current run loaded the same Go/review-fix skills: `gsd-core`, `caveman`, `golang-how-to`, `golang-testing`, `golang-performance`, `golang-benchmark`, `golang-observability`, `golang-context`, `golang-concurrency`, `golang-security`, `golang-error-handling`, `golang-safety`, `golang-lint`, `golang-cli`, `golang-dependency-management`, `golang-documentation`. `.pi/skills/go-implementation/SKILL.md` remains absent (`.pi/skills` contains only `gsd-core`).

Tests to add before production fixes:

- `internal/app` file-metrics reconciliation tests for deduped warehouse ETL: overwrite duplicate collapse, incremental append-dedup final count, and tombstone delete final count. Expected red at starting head: `pm.records.loaded` metric sum follows raw loaded records and exceeds final envelope/materialized counts.
- `internal/worker` gated worker-options tests: disabled worker options are empty; enabled worker options include contrib `WorkerInterceptor` while preserving no default-on behavior.
- `internal/app` benchmark split into disabled/enabled sub-benchmarks with enabled telemetry setup outside the hot loop and `b.ReportAllocs()` inside both sub-benchmarks.

## Independent-review correction RED plan

Session `153cfaabe3df4733a85717da46513786`; model `openai-codex/gpt-5.6-sol`; thinking `high`; starting HEAD `c6138292cfcc7205f7968a54b57a65f933a3c1fa`.

Accepted bounded findings from `/tmp/pm-397-review-415.log`:

- Contract test must require counters `pm.records.{read,transformed,loaded,failed}`, `pm.batches.{created,retried,skipped}`, `pm.api.{calls,retries,rate_limit_waits}`, and `pm.bytes.{read,written}`, plus histograms `pm.connector.operation.duration`, `pm.api.rate_limit_wait.duration`, and `pm.stage.duration`. Existing `pm.batches.flushed` remains a reconciliation metric.
- Batch methods remain local-only until `Flush`; HTTP and stage metrics emit only at existing operation/attempt/retry/completion seams. Test attributes must be bounded and must exclude synthetic secret markers, query strings, bodies, and headers.
- OTLP mode must export before `Shutdown` at a short configured test interval; disabled mode must create neither SDK nor collector traffic; file mode must retain one cumulative shutdown snapshot.
- Generic `OTEL_EXPORTER_OTLP_ENDPOINT=<collector>/prefix` must post metrics to `/prefix/v1/metrics`; `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=<collector>/custom` must post exactly to `/custom`.

Planned focused RED command (run after tests are added and before production edits):

```bash
go test ./internal/telemetry ./internal/connectors/connsdk ./internal/app ./internal/flow -run 'TestPRDMetricContract|TestOTLPMetricsExportBeforeShutdown|TestOTLPMetricEndpointPathSemantics|TestRequesterDoEmitsPRDMetrics|TestRunETLBatchMetricFamilies|TestEngineEmitsStageDurationMetric' -count=1
```

Expected RED at starting HEAD: missing metric APIs/families, OTLP collection only on shutdown, and generic endpoint posting to `/prefix` instead of `/prefix/v1/metrics`. Exact output will be captured before production changes.

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
| 8 | verify | Gate | `make verify` after `9894e6ef` | Pass | Full verify green, including smoke, connector lint, and `connectorgen validate`. |
| 9 | ci-fix | Generated parity | `cd website && pnpm run gen:website-data`; `cd website && COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 install --frozen-lockfile --reporter=silent && COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 run typecheck` | Pass | Website checks CI required `website/lib/docs.generated.ts`; regenerated docs data and typechecked with CI pnpm 11.7.0 because local pnpm 9.15.4 rejects the lockfile override config. |
| 10 | verify | Gate | `make verify`; `(cd website && COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 run gen:website-data)` with generated-data status check; `(cd website && COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 run typecheck)`; `git diff --check` | Pass | Final post-generated-data gates green; generated-data check has no tracked diff. |
| 11 | review-fix-plan | Planning | `scripts/gsd doctor`; `scripts/gsd list`; `scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415-review-fix.txt`; artifact updates | Pass/fallback | Doctor/list pass; `programming-loop` still unavailable (`scripts/gsd: unknown GSD command: programming-loop`). Review-fix plan/TDD/verification artifacts updated before production edits. Execution decision: `local_critical_path`. |
| 12 | review-fix-red | Test | `go test ./internal/app ./internal/worker -run 'TestRunETLDedupedMetrics\|TestTemporalWorker\|TestRunCounterHotPathAllocations' -count=1 > /tmp/pm-415-review-fix-red.txt 2>&1` | Fail/red | Expected red before production edits. Key exact output: `internal/worker/telemetry_test.go:37:14: undefined: temporalWorkerOptions`; `internal/worker/telemetry_test.go:43:13: undefined: temporalWorkerOptions`; `metrics_test.go:144: metric pm.records.loaded sum = 3, want 2`; `metrics_test.go:144: metric pm.records.loaded sum = 1, want 2` for incremental rematerialize; `metrics_test.go:144: metric pm.records.loaded sum = 1, want 2` for tombstone delete. Full red output stored at `/tmp/pm-415-review-fix-red.txt`. |
| 13 | review-fix-green | Test | `gofmt -w internal/app/local_warehouse.go internal/app/metrics_test.go internal/worker/submit.go internal/worker/serve.go internal/worker/telemetry_test.go && go test ./internal/app ./internal/worker -run 'TestRunETLDedupedMetrics\|TestTemporalWorker\|TestRunCounterHotPathAllocations\|TestTemporalClientOptionsTelemetryGated\|TestTemporalMetricsOnErrorWarnsWithoutPanic' -count=1` | Pass | Deduped loaded metrics now emit final materialized counts; worker telemetry options helper enabled/disabled tests pass. |
| 14 | review-fix-bench | Benchmark | `go test -bench BenchmarkEmit -benchmem ./internal/app` | Pass | `BenchmarkEmit/disabled-12 591002476 2.038 ns/op 0 B/op 0 allocs/op`; `BenchmarkEmit/enabled_file-12 589499779 2.036 ns/op 0 B/op 0 allocs/op`. Enabled telemetry setup happens outside hot loop; benchmark only supports measured hot-path allocation claim. |
| 15 | review-fix-focused | Test | `go test ./internal/telemetry ./internal/app ./internal/cli ./internal/worker -run 'Metric\|Telemetry\|Temporal\|Golden\|Config\|RunCounter\|Deduped' -count=1` | Pass | Focused telemetry/app/CLI/worker regression suite green after review fixes. |
| 16 | review-fix-full | Gate | `gofmt -w cmd internal && go vet ./... && go test -timeout 20m ./... && go build ./cmd/pm` | Pass | Full Go gates green; slow packages included `internal/cli` and `internal/connectors/certify`. |
| 17 | review-fix-verify | Gate | `make verify`; `git diff --check`; `git diff -- go.mod go.sum` | Pass | `make verify` passed including fmt/tidy-check/vet/full tests/build/docs validate/smoke/golangci-lint connector subset/connectorgen validate. `git diff --check` passed. Dependency diff empty for review-fix (no go.mod/go.sum changes). |
| 18 | independent-correction-plan | Planning | Read issue/PR, ADR 0004, PRD §15.2, phase artifacts, GSD fallback policy, required skills, and `/tmp/pm-397-review-415.log`; update correction artifacts | Pass/fallback | Session/model/thinking/starting HEAD recorded before production edits. GSD doctor/list pass; programming-loop command remains unavailable, so recorded manual fallback continues. Decision: `local_critical_path`. |
| 19 | independent-correction-red | Test | `go test ./internal/telemetry ./internal/connectors/connsdk ./internal/app ./internal/flow -run 'TestPRDMetricContract\|TestOTLPMetricsExportBeforeShutdown\|TestOTLPMetricEndpointPathSemantics\|TestRequesterDoEmitsPRDMetrics\|TestRunETLBatchMetricFamilies\|TestEngineEmitsStageDurationMetric' -count=1` | Fail/red | Exact RED captured in `/tmp/pm-415-correction-red.txt` before production edits. Telemetry build failed on missing batch methods, API/retry/rate-limit/latency/stage metric functions; connsdk values were all 0; ETL lacked `pm.batches.created`; flow lacked `pm.stage.duration`. Endpoint/live-export tests also require the missing `MetricExportInterval` field after the initial compile-error limit. |
| 20 | independent-correction-batch-red | Test | `go test ./internal/flow -run '^TestActionBatchRetryAndSkipMetrics$' -count=1` | Fail/red | Exact output in `/tmp/pm-415-correction-batch-red.txt`: no telemetry metric JSONL was produced for the existing action batch create/retry/idempotent-skip seams. Added before editing `internal/flow/action.go`. |
| 21 | independent-correction-green | Test | `go test ./internal/telemetry ./internal/connectors/connsdk ./internal/app ./internal/flow -run 'TestPRDMetricContract\|TestOTLPMetricsExportBeforeShutdown\|TestOTLPMetricEndpointPathSemantics\|TestDisabledMetricsStartsNoExporter\|TestRequesterDoEmitsPRDMetrics\|TestRunETLBatchMetricFamilies\|TestEngineEmitsStageDurationMetric\|TestActionBatchRetryAndSkipMetrics' -count=1` | Pass | Complete PRD metric contract, file snapshot, periodic OTLP, endpoint semantics, HTTP values, batch seams, and stage seam green. |
| 22 | independent-correction-race | Test | `go test -race ./internal/telemetry ./internal/connectors/connsdk ./internal/app ./internal/flow ./internal/worker -run 'Metric\|Telemetry\|Temporal\|RunCounter\|BatchRetry\|StageDuration\|OTLP' -count=1` | Pass | Focused race suite passed across all changed seams and Temporal gating. |
| 23 | independent-correction-benchmark | Benchmark | `go test -bench '^BenchmarkEmit$' -benchmem -count=5 ./internal/app` | Pass | Disabled `1.998–2.024 ns/op`, enabled file `1.996–2.042 ns/op`; both `0 B/op`, `0 allocs/op` in every run. No material regression versus reviewed `2.023–2.081 ns/op`. |
| 24 | independent-correction-focused | Test | OTLP live/path/disabled tests under `-race -count=10`; app+CLI metric reconciliation under `-race`; worker Temporal options/OnError under `-race` | Pass | OTLP tests stable across 10 runs; file/envelope and deduped reconciliation green; Temporal enablement gate remains default-off. |
| 25 | independent-correction-static-modules | Gate | relevant package `go vet`; `go mod verify`; `go mod tidy -diff`; `git diff --exit-code -- go.mod go.sum`; `git diff --check` | Pass | No vet findings, modules verified, tidy diff empty, dependency files unchanged, worktree whitespace clean. |
| 26 | independent-correction-full-go | Gate | `gofmt -w cmd internal`; `go vet ./...`; `go test -timeout 20m ./...`; `go build ./cmd/pm` | Pass | Full Go suite passed, including slow CLI/certify packages; build passed. |
| 27 | independent-correction-verify | Gate | `make verify` | Pass | Full repository gate passed: fmt, tidy-check, vet, full tests, build, docs validate, smoke, connector lint, and 547-bundle validation. |
| 28 | independent-correction-commit-push | Checkpoint | Commit `ceb3a35ce13642a0d8c8ea0010272582202f8afd`; reviewed-range `git diff --check`; `git push origin feat/415-otel-metrics` | Pass | Implementation correction committed and pushed to existing PR branch; local/remote heads matched; no new PR or external review request. |
