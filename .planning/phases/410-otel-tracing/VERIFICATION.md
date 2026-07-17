# Verification — Phase 410 OpenTelemetry tracing

## Required gates

- [ ] File exporter smoke: enabled tracing writes JSONL spans with expected names.
- [ ] Off-mode check: disabled/default tracing creates no SDK and no `.polymetrics/telemetry` directory.
- [ ] Secret grep: telemetry exports contain no artificial test token, query string, body, header, argv, or credential value.
- [ ] Envelope-only stdout: JSON CLI commands still emit one final envelope; telemetry warnings go to stderr.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify`.

## Focused test gates

- [ ] `go test ./internal/telemetry -count=1`.
- [ ] `go test ./internal/connectors/connsdk -count=1`.
- [ ] `go test ./internal/cli -run 'Telemetry|Golden|Agentic|Config' -count=1`.
- [ ] `go test ./internal/app -run Telemetry -count=1`.
- [ ] `go test ./internal/flow -run Telemetry -count=1`.
- [ ] `go test ./internal/connectors/certify -run Telemetry -count=1` or CLI certify wrapper equivalent.

## CLI help/docs/website parity checklist

Applies because config/env/help docs change.

- [ ] Runtime help: `./pm --help` includes telemetry opt-in docs.
- [ ] Runtime help: `./pm help config` includes telemetry config keys and safety constraints.
- [ ] Command help: `./pm config --help` is not a real command; topic-only check is applicable.
- [ ] Bare namespaces unaffected: spot-check `./pm etl`, `./pm flow`, `./pm connectors` still contextual help exit 0 or pre-existing behavior.
- [ ] Invalid actions still usage errors.
- [ ] `docs/cli/config.md` generated/updated from embedded docs.
- [ ] Website docs under `website/content/docs/cli-reference.mdx` updated.
- [ ] Generated website data `website/lib/docs.generated.ts` updated if generator changes source.
- [ ] Generated docs check: `./pm docs generate --dir "$TMP/cli" --connectors-dir "$TMP/connectors"` then `diff -ru docs/cli "$TMP/cli"`.
- [ ] Completion metadata: not applicable; no completion surface added.

## Dependency verification

- [ ] `go.mod` direct OTel lines match ADR 0004 Stage 12 only.
- [ ] No `otelhttp`, metrics SDK, otel log bridge, Temporal OTel contrib, or grpc exporter added.
- [ ] `go mod tidy`/`make tidy-check` clean.

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
