# Verification — Phase 415 OpenTelemetry metrics

## Required gates

- [x] Red tests captured before production edits.
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
- [x] Generated website data updated if generator requires (not required; MDX-only text change, docs data generator not part of this CLI reference update).
- [x] Completion metadata: not applicable unless command/flag completion changes.

## Focused commands to run

```bash
go test ./internal/telemetry -run 'Metric|Telemetry' -count=1
go test ./internal/app -run 'Metric|Telemetry' -count=1
go test ./internal/cli -run 'Metric|Telemetry|Golden|Config' -count=1
go test ./internal/worker -run 'Metric|Telemetry|Temporal' -count=1
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
