# Verification — Phase 415 OpenTelemetry metrics

## Independent-review correction gates

Correction worker: session `153cfaabe3df4733a85717da46513786`; model `openai-codex/gpt-5.6-sol`; thinking `high`; starting HEAD `c6138292cfcc7205f7968a54b57a65f933a3c1fa`.

- [x] Exact focused RED evidence captured before correction production edits (`/tmp/pm-415-correction-red.txt`).
- [x] PRD §15.2 contract enumerates every required metric name and counter/histogram kind.
- [x] Batch created/retried/skipped methods stay local until batch `Flush`; representative ETL/action batch values reconcile without per-record OTel calls.
- [x] HTTP calls/retries/rate-limit waits, connector-operation duration, and bytes read/written emit at existing bounded seams with secret-safe low-cardinality attributes.
- [x] Stage duration emits at existing ETL/flow stage completion seams with bounded stage labels.
- [x] OTLP metrics export periodically before worker shutdown; disabled mode starts no exporter.
- [x] File metrics remain one cumulative shutdown snapshot and reconcile with final ETL envelope counts.
- [x] Generic OTLP endpoint path appends `/v1/metrics`; trace-signal endpoint rewrites to metrics; metric-specific endpoint remains exact.
- [x] Existing Temporal tracing/metrics gating remains green without runtime services.
- [x] Focused `-race` telemetry/app/connsdk/flow/worker tests pass.
- [x] `BenchmarkEmit` disabled/enabled paths remain 0 allocs/op; no material hot-loop regression.
- [ ] `git diff --check 56a7ecb08f755184af7b55318c3285582d5adfb7..HEAD` passes after correction commit.
- [x] `gofmt -w cmd internal`, relevant/full `go vet`, `go mod verify`, and `go mod tidy -diff` pass with no dependency changes.
- [x] Full package tests, build, and `make verify` pass.
- [ ] Existing branch pushed; no new PR and no external review request.

CLI help/manual/website parity: not applicable to this correction because it changes internal metrics completeness and exporter behavior only; no command, flag, config key, output envelope, help topic, or docs promise changes.

## Required gates

- [x] Red tests captured before original production edits.
- [x] Review-fix red tests captured before production edits for PR #461 findings.
- [x] Deduped warehouse ETL file metrics reconcile `pm.records.loaded` with final envelope counts after overwrite/incremental/delete materialization.
- [x] Temporal worker constructors use telemetry-gated worker options; disabled path remains empty/default and enabled path includes contrib worker interceptor.
- [x] `BenchmarkEmit` has enabled and disabled sub-benchmarks with enabled setup outside the hot loop and allocation reporting.
- [x] Metrics disabled/default mode constructs no metric SDK and creates no `.polymetrics/telemetry` directory.
- [x] File exporter metrics reconcile with final `ETLRun` envelope counts.
- [x] Local counters accumulate in hot paths and flush by batch; no per-record OTel instrument calls.
- [x] `BenchmarkEmit` with allocations recorded: `go test -bench BenchmarkEmit -benchmem ./internal/app`.
- [x] Temporal tracing interceptor + metrics handler from contrib package are gated on `telemetry.Enabled(ctx)`.
- [x] OTLP/network metrics endpoint/exporter are trusted env/flag only; config-file OTLP values ignored/warned; unsupported `OTEL_EXPORTER_OTLP_*` and SDK env sanitized by env name only.
- [x] No metric attrs/warnings include request bodies, headers, query strings, raw argv, credentials, or synthetic secret markers.
- [x] Warnings are stderr-only and exit-code-neutral.
- [x] `gofmt -w cmd internal`.
- [x] `go vet ./...`.
- [x] Focused package tests: `go test ./internal/telemetry ./internal/app ./internal/cli ./internal/worker -run 'Metric|Telemetry|BenchmarkEmit|Temporal' -count=1` (adjust regex to actual tests).
- [x] `go test ./...` or honest extended timeout if default hits existing slow tests.
- [x] `go build ./cmd/pm`.
- [x] `make verify` when feasible (post-commit rerun passed).
- [x] `git diff --check`.
- [x] Dependency diff review: `git diff -- go.mod go.sum`; only ADR 0004 metrics/contrib modules and OTel metric API promotion at v1.44.0 plus OTel metric/x v0.66.0 indirect checksum.

## CLI help/docs/website parity checklist

Applies because telemetry config/help/docs change.

- [x] Runtime help: `./pm --help` mentions default-off traces+metrics and safe exporters.
- [x] Runtime help: `./pm help config` lists telemetry keys/env aliases and metrics endpoint rules.
- [x] Bare namespaces spot-check: `./pm etl`, `./pm flow`, `./pm connectors` remain contextual help / pre-existing behavior.
- [x] Invalid actions still usage errors: `./pm connectors bogus --json` exits 2.
- [x] `docs/cli/config.md` generated/updated from embedded docs.
- [x] Website docs under `website/content/docs/cli-reference.mdx` updated.
- [x] Generated website data updated if generator requires (`website/lib/docs.generated.ts`).
- [x] Completion metadata: not applicable unless command/flag completion changes.

## Focused commands to run

```bash
go test ./internal/telemetry -run 'Metric|Telemetry' -count=1
go test ./internal/app -run 'Metric|Telemetry' -count=1
go test ./internal/cli -run 'Metric|Telemetry|Golden|Config' -count=1
go test ./internal/worker -run 'Metric|Telemetry|Temporal' -count=1
go test -bench BenchmarkEmit -benchmem ./internal/app
```

## Review-fix focused commands

```bash
go test ./internal/app -run 'TestRunETLDedupedMetrics|TestRunETLMetricsAccumulateAndFlushByBatch|TestRunCounterHotPathAllocations' -count=1
go test ./internal/worker -run 'TestTemporal.*Options|TestTemporalMetricsOnErrorWarnsWithoutPanic' -count=1
go test -bench BenchmarkEmit -benchmem ./internal/app
```

## Full gates to run

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check
git diff -- go.mod go.sum
```

## Results log

| Command | Result | Evidence |
|---|---|---|
| `scripts/gsd doctor` | pass | Adapter checks `ok`; 69 commands. |
| `scripts/gsd prompt plan-phase 415-otel-metrics --skip-research` | pass | Prompt generated at `/tmp/gsd-plan-415.txt`. |
| `scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run` | fail/fallback | `scripts/gsd: unknown GSD command: programming-loop`; manual fallback to `.pi/prompts/pm-gsd-loop.md`. |
| `go test ./internal/app ./internal/cli ./internal/worker -run 'Metric\|Telemetry\|Temporal' -count=1` | fail/red | Expected red: missing `telemetry.NewRunCounters`, missing `temporalClientOptions`, no file metrics for ETL counts, and no OTLP metrics endpoint warning/sanitization yet. |
| Continuation preflight: `git status --short --branch`; `scripts/gsd doctor`; `scripts/gsd list`; `scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415-continuation.txt` | pass/fallback | Dirty worktree preserved after previous worker output truncation; doctor/list pass; programming-loop remains unavailable with same unknown-command fallback. |
| `go test ./internal/telemetry ./internal/app ./internal/cli ./internal/worker -run 'Metric\|Telemetry\|Temporal\|Golden\|Config\|RunCounter' -count=1` | pass | Focused telemetry/app/CLI/worker tests green after metrics implementation, docs generation, and Temporal OnError warning hook. |
| `gofmt -w cmd internal` | pass | No remaining format diff. |
| `go vet ./...` | pass | No vet findings. |
| `go test -timeout 20m ./...` | pass | Full suite green; slow packages included `internal/cli` and `internal/connectors/certify`. |
| `go test -bench BenchmarkEmit -benchmem ./internal/app` | pass | `BenchmarkEmit-12 592256128 2.040 ns/op 0 B/op 0 allocs/op`. |
| `go build ./cmd/pm` | pass | Binary builds. |
| `./pm --help`; `./pm help config`; `./pm etl`; `./pm flow`; `./pm connectors`; `./pm connectors bogus --json` | pass | Help/docs parity spot-checks green; invalid connector action exits 2 with JSON error. |
| `git diff --check` | pass | No whitespace errors. |
| `git diff -- go.mod go.sum`; `go list -m all | grep -E '^(go\\.opentelemetry\\.io/otel|go\\.temporal\\.io/sdk/contrib/opentelemetry)( |/)'` | pass | Direct additions: `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp@v1.44.0`, `go.opentelemetry.io/otel/exporters/stdout/stdoutmetric@v1.44.0`, `go.opentelemetry.io/otel/sdk/metric@v1.44.0`, `go.temporal.io/sdk/contrib/opentelemetry@v0.7.0`; `go.opentelemetry.io/otel/metric@v1.44.0` promoted from indirect due API import; `go.opentelemetry.io/otel/metric/x@v0.66.0` checksum added transitively. |
| `make verify` before commit | fail/expected | Stopped at `tidy-check` because dependency diff was intentionally uncommitted; rerun after committing dependency delta. |
| `make verify` after `9894e6ef` | pass | Full verify passed: fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, smoke, golangci-lint connector subset (`0 issues`), and `connectorgen validate` (`547 connector(s) checked, 0 findings`). |
| `cd website && pnpm run gen:website-data` | pass | Regenerated `website/lib/docs.generated.ts` after Website checks CI reported docs data drift. |
| `cd website && COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 install --frozen-lockfile --reporter=silent && COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 run typecheck` | pass | Website dependencies installed with CI pnpm version; `tsc --noEmit` passed. Local pnpm 9.15.4 frozen install fails on lockfile override mismatch, so CI-matched corepack pnpm 11.7.0 was used. |
| Final post-generated-data gates | pass | `make verify` passed; `COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 run gen:website-data` produced no tracked generated-data diff; `COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm@11.7.0 run typecheck` passed; `git diff --check` passed. |
| Review-fix preflight and artifacts | pass/fallback | `scripts/gsd doctor` and `scripts/gsd list` passed; `scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415-review-fix.txt` failed with `scripts/gsd: unknown GSD command: programming-loop`; plan/TDD/verification artifacts updated before production edits. |
| Review-fix red tests | fail/red | `go test ./internal/app ./internal/worker -run 'TestRunETLDedupedMetrics\|TestTemporalWorker\|TestRunCounterHotPathAllocations' -count=1 > /tmp/pm-415-review-fix-red.txt 2>&1` failed as expected: missing `temporalWorkerOptions`; `pm.records.loaded` sums `3 want 2`, `1 want 2`, `1 want 2` for deduped overwrite/incremental/delete cases. |
| Review-fix focused tests | pass | `gofmt -w internal/app/local_warehouse.go internal/app/metrics_test.go internal/worker/submit.go internal/worker/serve.go internal/worker/telemetry_test.go && go test ./internal/app ./internal/worker -run 'TestRunETLDedupedMetrics\|TestTemporalWorker\|TestRunCounterHotPathAllocations\|TestTemporalClientOptionsTelemetryGated\|TestTemporalMetricsOnErrorWarnsWithoutPanic' -count=1` passed. |
| Review-fix benchmark | pass | `go test -bench BenchmarkEmit -benchmem ./internal/app`: `BenchmarkEmit/disabled-12 591002476 2.038 ns/op 0 B/op 0 allocs/op`; `BenchmarkEmit/enabled_file-12 589499779 2.036 ns/op 0 B/op 0 allocs/op`. |
| Review-fix broader focused tests | pass | `go test ./internal/telemetry ./internal/app ./internal/cli ./internal/worker -run 'Metric\|Telemetry\|Temporal\|Golden\|Config\|RunCounter\|Deduped' -count=1` passed. |
| Review-fix full Go gates | pass | `gofmt -w cmd internal && go vet ./... && go test -timeout 20m ./... && go build ./cmd/pm` passed. |
| Review-fix make verify | pass | `make verify` passed: fmt, tidy-check, vet, full tests, build, docs validate, smoke, golangci-lint connector subset (`0 issues`), and `connectorgen validate` (`547 connector(s) checked, 0 findings`). |
| Review-fix diff/dependency checks | pass | `git diff --check` passed; `git diff -- go.mod go.sum` produced no output, so review-fix dependency delta is empty. |
| Independent correction preflight | pass/fallback | Exact starting HEAD confirmed; worktree initially clean; `scripts/gsd doctor`/`list` pass; `programming-loop` command unavailable; correction verification pending and therefore not yet passed. |
| Independent correction focused RED | fail/expected | Exact command failed before production edits with missing PRD metric APIs, all-zero HTTP metric values, absent ETL batch-created metric, and absent flow stage-duration metric. Full output: `/tmp/pm-415-correction-red.txt`. |
| Independent correction batch RED | fail/expected | Action batch retry/idempotent-skip test produced no metric JSONL before `internal/flow/action.go` edits. Full output: `/tmp/pm-415-correction-batch-red.txt`. |
| Correction focused GREEN | pass | PRD contract, HTTP/batch/stage behavior, periodic export, disabled exporter, and endpoint path tests passed. |
| Correction focused race | pass | `go test -race` passed for telemetry, connsdk, app, flow, and worker telemetry tests; OTLP live/path/disabled tests also passed `-race -count=10`. |
| Correction reconciliation/Temporal | pass | App+CLI metric/deduped tests passed under race; Temporal client/worker options and OnError gating passed under race without services. |
| Correction benchmark | pass | 5 runs: disabled `1.998–2.024 ns/op`; enabled file `1.996–2.042 ns/op`; all `0 B/op`, `0 allocs/op`. |
| Correction modules/static | pass | Relevant and full vet passed; `go mod verify` passed; `go mod tidy -diff` and dependency diff empty. |
| Correction full Go gates | pass | `gofmt -w cmd internal`, `go vet ./...`, `go test -timeout 20m ./...`, and `go build ./cmd/pm` passed. |
| Correction `make verify` | pass | Full gate passed, including docs validate, smoke, connector lint, and 547-bundle validation. |
