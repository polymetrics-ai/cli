# Verification — Phase 415 OpenTelemetry metrics

## Required gates

- [ ] Red tests captured before production edits.
- [ ] Metrics disabled/default mode constructs no metric SDK and creates no `.polymetrics/telemetry` directory.
- [ ] File exporter metrics reconcile with final `ETLRun` envelope counts.
- [ ] Local counters accumulate in hot paths and flush by batch; no per-record OTel instrument calls.
- [ ] `BenchmarkEmit` with allocations recorded: `go test -bench BenchmarkEmit -benchmem ./internal/app`.
- [ ] Temporal tracing interceptor + metrics handler from contrib package are gated on `telemetry.Enabled(ctx)`.
- [ ] OTLP/network metrics endpoint/exporter are trusted env/flag only; config-file OTLP values ignored/warned; unsupported `OTEL_EXPORTER_OTLP_*` and SDK env sanitized by env name only.
- [ ] No metric attrs/warnings include request bodies, headers, query strings, raw argv, credentials, or synthetic secret markers.
- [ ] Warnings are stderr-only and exit-code-neutral.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] Focused package tests: `go test ./internal/telemetry ./internal/app ./internal/cli ./internal/worker -run 'Metric|Telemetry|BenchmarkEmit|Temporal' -count=1` (adjust regex to actual tests).
- [ ] `go test ./...` or honest extended timeout if default hits existing slow tests.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify` when feasible.
- [ ] `git diff --check`.
- [ ] Dependency diff review: `git diff -- go.mod go.sum`; only exact ADR 0004 metrics/contrib modules allowed.

## CLI help/docs/website parity checklist

Applies because telemetry config/help/docs change.

- [ ] Runtime help: `./pm --help` mentions default-off traces+metrics and safe exporters.
- [ ] Runtime help: `./pm help config` lists telemetry keys/env aliases and metrics endpoint rules.
- [ ] Bare namespaces spot-check: `./pm etl`, `./pm flow`, `./pm connectors` remain contextual help / pre-existing behavior.
- [ ] Invalid actions still usage errors: `./pm connectors bogus --json` exits 2.
- [ ] `docs/cli/config.md` generated/updated from embedded docs.
- [ ] Website docs under `website/content/docs/cli-reference.mdx` updated.
- [ ] Generated website data updated if generator requires.
- [ ] Completion metadata: not applicable unless command/flag completion changes.

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
