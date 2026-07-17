# Verification — Phase 410 OpenTelemetry tracing

## Required gates

- [x] File exporter smoke: enabled tracing writes JSONL spans with expected names.
- [x] Off-mode check: disabled/default tracing creates no SDK and no `.polymetrics/telemetry` directory.
- [x] Secret grep: telemetry exports contain no artificial test token, query string, body, header, argv, or credential value.
- [x] Envelope-only stdout: JSON CLI commands still emit one final envelope; telemetry warnings go to stderr.
- [x] `gofmt -w cmd internal`.
- [x] `go vet ./...`.
- [x] `go test ./...`.
- [x] `go build ./cmd/pm`.
- [x] `make verify` (passed after green-slice commit).

## Focused test gates

- [x] `go test ./internal/telemetry -count=1` (covered by focused regex run).
- [x] `go test ./internal/connectors/connsdk -count=1` (covered by focused regex run).
- [x] `go test ./internal/cli -run 'Telemetry|Golden|Agentic|Config' -count=1` (`Telemetry|TestLoadTelemetry|Golden|Config` run passed; agentic covered in `go test ./...`).
- [x] `go test ./internal/app -run Telemetry -count=1` (covered by focused regex run).
- [x] `go test ./internal/flow -run Telemetry -count=1` (covered by focused regex run).
- [x] `go test ./internal/connectors/certify -run Telemetry -count=1` or CLI certify wrapper equivalent (`TestTelemetryCertifyConnectorSpan` passed in `internal/cli`).

## CLI help/docs/website parity checklist

Applies because config/env/help docs change.

- [x] Runtime help: `./pm --help` includes telemetry opt-in docs.
- [x] Runtime help: `./pm help config` includes telemetry config keys and safety constraints.
- [x] Command help: `./pm config --help` is not a real command; topic-only check is applicable.
- [x] Bare namespaces unaffected: spot-check `./pm etl`, `./pm flow`, `./pm connectors` still contextual help exit 0 or pre-existing behavior.
- [x] Invalid actions still usage errors (`./pm connectors bogus --json` exit 2).
- [x] `docs/cli/config.md` generated/updated from embedded docs.
- [x] Website docs under `website/content/docs/cli-reference.mdx` updated.
- [x] Generated website data `website/lib/docs.generated.ts` updated.
- [x] Generated docs check: `./pm docs generate --dir "$TMP/cli" --connectors-dir "$TMP/connectors"` then `diff -ru docs/cli "$TMP/cli"`.
- [x] Completion metadata: not applicable; no completion surface added.

## Dependency verification

- [x] `go.mod` direct OTel lines match ADR 0004 Stage 12 trace modules (`otel`, `sdk`, `stdouttrace`, `otlptracehttp` at v1.44.0).
- [x] No `otelhttp`, metrics SDK direct import, otel log bridge, Temporal OTel contrib, or grpc exporter added.
- [x] MVS consequence recorded: OTel v1.44.0 updates existing `golang.org/x/*`, `google.golang.org/grpc`, `grpc-gateway`, and `go.opentelemetry.io/*` indirects; no unapproved top-level non-OTel module was intentionally added.
- [x] `go mod tidy`/`make tidy-check` clean after commit (`make verify` passed).

## Runtime/credential boundaries

- Runtime services not started unless explicitly requested.
- No credentialed connector checks.
- Reverse ETL plan → preview → approval → execute semantics untouched; no reverse ETL execution.

## Results log

| Command | Result | Evidence |
|---|---|---|
| `scripts/gsd doctor` | pass | Adapter checks all `ok`; 69 commands. |
| `scripts/gsd prompt plan-phase 410 --skip-research` | pass | 142-line prompt generated. |
| `scripts/gsd prompt programming-loop init --phase 410 --dry-run` | fail/fallback | `unknown GSD command: programming-loop`; manual GSD fallback active. |
| `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/app ./internal/flow -run 'Telemetry\|TestLoadTelemetry\|Golden\|Config' -count=1` | pass | Focused telemetry/config/CLI/connsdk/app/flow gate. |
| File/off/secret smoke script | pass | `PM_TELEMETRY=file` command/certify spans; default no telemetry dir; synthetic marker/forbidden attr grep clean. |
| `gofmt -w cmd internal` | pass | No output. |
| `go vet ./...` | pass | No output. |
| `go test ./...` | pass | Full test suite passed; slowest packages `internal/connectors/certify` ~352s and `internal/cli` ~208s. |
| `go build ./cmd/pm` | pass | No output. |
| `make verify` | pass | fmt, tidy-check, vet, 20m tests, build, docs validate, smoke, lint, connectorgen validate all passed. |
