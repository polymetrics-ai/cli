# Verification — Phase 410 OpenTelemetry tracing

## Final security hardening checklist (PR #459 head `216b5e076d82302621574a9fb4fa71c8acb79204`)

- [x] Red tests captured before production edits for SDK/provider/resource ambient OTel env bypass.
- [x] `OTEL_RESOURCE_ATTRIBUTES=api_key=<synthetic>` and `OTEL_SERVICE_NAME=<synthetic>` are sanitized/unset during resource/provider construction and never appear in file exporter spans.
- [x] Same resource attr marker never appears in OTLP HTTP payloads.
- [x] Invalid `OTEL_TRACES_SAMPLER(_ARG)` cannot write raw SDK parse errors to process stderr; project warning names env vars only and omits values.
- [x] Unsupported SDK env names around `sdktrace.NewTracerProvider` are warned by name only and temporarily unset while building explicit safe resource/provider.
- [x] `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` remains in config test env cleanup/list.
- [x] Focused telemetry/config/cli tests, full gates (`go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`), docs generation if artifacts changed, `git diff --check`, and `git diff -- go.mod go.sum` complete.

## Final focused review-fix checklist (PR #459 head `75433cefa9a00671b06c6c3e83bcde1e4730211c`)

- [x] Red tests captured before production edits for ambient OTel env bypass and docs/help wording residuals.
- [x] Unsafe `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` is validated/rejected before exporter construction, emits only redacted project `warning: telemetry:`, produces no raw process stderr, and does not contact the unvalidated collector.
- [x] Unsupported `OTEL_EXPORTER_OTLP_HEADERS` / traces headers and other unsafe OTLP env are warned/neutralized before `otlptracehttp.New`, no raw process stderr, no secret/header reaches collector.
- [x] OTLP exporter always passes a validated endpoint/default and empty supported headers to the SDK constructor.
- [x] ADR 0004 superseding note matches hardened trusted-env behavior.
- [x] Root help/goldens list `none`, `off`, `file`, and `otlp`.
- [x] Config docs/website say both OTLP exporter and endpoint must come from trusted env/flag; config-file endpoint alone is ignored.
- [x] Focused tests/docs generation/goldens/website data plus `gofmt`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify` if feasible, `git diff --check`, and `git diff -- go.mod go.sum` complete.

## Review-fix verification checklist (PR #459)

- [x] Red tests captured before production edits for accepted findings.
- [x] Exported span tests assert error metadata is allowlisted, registry-redacted, and contains no SDK `exception.*` attrs/events.
- [x] Failed command, HTTP, and flow spans omit synthetic registered marker and response-body-like details; `capture=minimal` suppresses message-like error attrs.
- [x] Config-sourced OTLP exporter/endpoint rejected or disabled by default with sanitized warning; env/CLI opt-in accepted.
- [x] Telemetry file exporter rejects absolute paths, `..` escapes, symlinked dirs, and symlinked files; created dirs/files use restrictive permissions.
- [x] Event attrs are attached to span events and remain allowlisted; retry/attempt/status metadata not overwritten/lost.
- [x] OTLP init/export/shutdown failures preserve exit code and stdout JSON; stderr uses redacted `warning: telemetry:` only.
- [x] OTLP endpoint validation rejects userinfo/query/fragment and non-http(s) without leaking endpoint secrets.
- [x] Root/config help, docs/cli, website data, and goldens mention HTTP method/status/attempt/retry attrs and supported exporters `none,off,file,otlp` consistently.
- [x] Exporter/log/telemetry warning smokes show redacted stderr and uncorrupted stdout JSON; Redis runtime smoke not applicable to tracing review fix.
- [x] `git diff --check` and `git diff -- go.mod go.sum` clean/expected (no new module/version; review-fix event attrs promote existing `go.opentelemetry.io/otel/trace v1.44.0` from indirect to direct).

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

- [x] Initial `go.mod` direct OTel lines matched ADR 0004 Stage 12 trace modules (`otel`, `sdk`, `stdouttrace`, `otlptracehttp` at v1.44.0).
- [x] Review-fix event attrs require `trace.WithAttributes`; `go mod tidy` promotes existing `go.opentelemetry.io/otel/trace v1.44.0` from indirect to direct with no version or checksum change.
- [x] No `otelhttp`, metrics SDK direct import, otel log bridge, Temporal OTel contrib, or grpc exporter added.
- [x] MVS consequence recorded: OTel v1.44.0 updates existing `golang.org/x/*`, `google.golang.org/grpc`, `grpc-gateway`, and `go.opentelemetry.io/*` indirects; no unapproved top-level non-OTel module was intentionally added.
- [x] `go mod tidy`/`make tidy-check` clean after commit (`make verify` passed).

## Runtime/credential boundaries

- Runtime services not started unless explicitly requested.
- No credentialed connector checks.
- Reverse ETL plan → preview → approval → execute semantics untouched; no reverse ETL execution.

## Review-fix focused gates to run

```bash
go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'Telemetry|OTLP|Endpoint|Event|RecordError|FileExporter' -count=1
go test ./internal/connectors/connsdk -run Telemetry -count=1
go test ./internal/cli -run 'Telemetry|Golden|Config|Agentic' -count=1
go test ./internal/app -run Telemetry -count=1
go test ./internal/flow -run Telemetry -count=1
```

## Final SDK-env hardening focused gates to run

```bash
go test ./internal/cli -run 'TestTelemetrySanitizesSDKResourceEnvForFileExporter|TestTelemetrySanitizesSDKResourceEnvForOTLPExporter|TestTelemetrySanitizesInvalidSamplerEnvBeforeProvider' -count=1
go test ./internal/telemetry ./internal/config ./internal/cli -run 'Telemetry|TestLoadTelemetry|Config' -count=1
```

## Review-fix final gates to rerun

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check
git diff -- go.mod go.sum
```

Additional smoke/parity:

- file/off/secret telemetry smoke with synthetic marker.
- OTLP failure/endpoint smoke with no credentialed services.
- `./pm --help`, `./pm help config`, `./pm etl`, `./pm flow`, `./pm connectors`, invalid `./pm connectors bogus --json`.
- docs generation/diff and `npm --prefix website run gen:docs` when help/docs change.

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
| `scripts/gsd doctor` (review-fix rerun) | pass | Adapter checks all `ok`; 69 commands. |
| `scripts/gsd prompt plan-phase 410 --skip-research` (review-fix rerun) | pass | 142-line prompt generated. |
| `scripts/gsd prompt programming-loop init --phase 410 --dry-run` (review-fix rerun) | fail/fallback | `unknown GSD command: programming-loop`; manual GSD fallback remains active. |
| `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'Telemetry\|OTLP\|Endpoint\|Event\|RecordError\|FileExporter' -count=1` (review-fix red) | fail | Red evidence captured in `TDD-LEDGER.md` row 11 before production edits. |
| `scripts/gsd doctor` (review-fix resume rerun) | pass | Adapter checks all `ok`; 69 commands. |
| `scripts/gsd prompt plan-phase 410 --skip-research` (review-fix resume rerun) | pass | Prompt regenerated to `/tmp/gsd-plan-phase-410-reviewfix-rerun.txt`. |
| `scripts/gsd prompt programming-loop init --phase 410 --dry-run` (review-fix resume rerun) | fail/fallback | `unknown GSD command: programming-loop`; manual GSD fallback remains active. |
| `go test ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'TestTelemetryFailedCommandSpanDoesNotExportRawError\|TestRequesterDoFailedHTTPSpanHasSafeErrorAndEventAttrs\|TestEngineRunFailedStepTelemetryRedactsError' -count=1` (stable error red) | fail | Expected red: missing `internal_error`, forbidden `connsdk.HTTPError`, forbidden `errors.errorString`/`fmt.wrapError`. |
| `go test ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'TestTelemetryFailedCommandSpanDoesNotExportRawError\|TestRequesterDoFailedHTTPSpanHasSafeErrorAndEventAttrs\|TestEngineRunFailedStepTelemetryRedactsError' -count=1` | pass | Stable class/code/status error metadata green. |
| `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'Telemetry\|OTLP\|Endpoint\|Event\|RecordError\|FileExporter' -count=1` | pass | Review-fix focused telemetry/config/CLI/connsdk/flow gate passed. |
| `go test ./internal/telemetry -count=1`; `go test ./internal/connectors/connsdk -run Telemetry -count=1`; `go test ./internal/cli -run 'Telemetry\|Golden\|Config\|Agentic' -count=1`; `go test ./internal/app -run Telemetry -count=1`; `go test ./internal/flow -run Telemetry -count=1`; `go test ./internal/config -count=1` | pass | Focused package gates passed after docs/golden generation. |
| File/off/secret telemetry smoke | pass | Synthetic marker smoke: off mode no telemetry dir; file mode command/certify spans; stdout JSON parsed; forbidden telemetry grep clean. |
| OTLP endpoint smoke | pass | Invalid endpoint preserved Version stdout JSON, emitted redacted `warning: telemetry:` with no synthetic marker/userinfo/query leak. |
| Help/docs/website generation | pass | `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts`; `./pm docs generate --dir docs/cli --connectors-dir $TMP/connectors`; `npm --prefix website run gen:docs`; golden/docs test passed. |
| `go vet ./...`; `go test ./...`; `go build ./cmd/pm` | pass | Full Go vet/test/build passed before `make verify` rerun. |
| `make verify` (review-fix first run) | fail | `tidy-check` promoted existing `go.opentelemetry.io/otel/trace v1.44.0` from indirect to direct because event attrs use `trace.WithAttributes`; rerun after accepting tidy diff required. |
| `make verify` (review-fix after commit) | pass | fmt, tidy-check, vet, 20m tests, build, docs validate, smoke, lint, and connectorgen validate passed. |
| `git diff --check`; `git diff -- go.mod go.sum` | pass | No output after final planning update. |
| `go test ./internal/cli -run 'TestTelemetryRejectsUnsafeAmbientOTLPTracesEndpointBeforeExporter\|TestTelemetryNeutralizesUnsupportedAmbientOTLPHeaders' -count=1` (final residual red) | fail | Red evidence captured in `TDD-LEDGER.md` row 20 before production edits. |
| `go test ./internal/cli -run 'TestTelemetryRejectsUnsafeAmbientOTLPTracesEndpointBeforeExporter\|TestTelemetryNeutralizesUnsupportedAmbientOTLPHeaders' -count=1` | pass | Unsafe traces endpoint rejected before exporter construction; unsupported OTLP headers neutralized with no raw process stderr and no ambient Authorization header. |
| `go test ./internal/telemetry ./internal/config ./internal/cli -run 'Telemetry\|TestLoadTelemetry\|Golden\|Config' -count=1` | pass | Ambient OTLP env hardening, config traces endpoint alias, goldens, and docs drift checks passed. |
| `go test ./internal/telemetry -count=1`; `go test ./internal/connectors/connsdk -run Telemetry -count=1`; `go test ./internal/app -run Telemetry -count=1`; `go test ./internal/flow -run Telemetry -count=1` | pass | Focused telemetry consumers passed. |
| `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1`; `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir $TMP/connectors`; `npm --prefix website run gen:docs` | pass | Root help goldens, docs/cli config page, website source/generated data updated. |
| `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm` (final residual) | pass | Full Go gates passed; `internal/cli` ~195s and `internal/connectors/certify` ~346s in this run. |
| `make verify` (final residual) | pass | fmt, tidy-check, vet, 20m tests, build, docs validate, smoke, lint, and connectorgen validate passed. |
| Runtime help/docs parity final | pass | `./pm --help`; `./pm help config`; `./pm etl`; `./pm flow`; `./pm connectors`; invalid `./pm connectors bogus --json` exit 2; generated docs diff clean; `npm --prefix website run gen:docs` passed. |
| `git diff --check`; `git diff -- go.mod go.sum` (final residual) | pass | No output. |
| `go test ./internal/cli ./internal/config -run 'TestTelemetrySanitizesSDKResourceEnvForFileExporter\|TestTelemetrySanitizesSDKResourceEnvForOTLPExporter\|TestTelemetrySanitizesInvalidSamplerEnvBeforeProvider\|TestAllBoundEnvVarsIncludesOTLPTracesEndpoint' -count=1` (SDK-env red) | fail | Expected red: missing SDK/resource warnings, invalid sampler wrote raw process stderr, config env list missing traces endpoint. |
| `go test ./internal/cli ./internal/config -run 'TestTelemetrySanitizesSDKResourceEnvForFileExporter\|TestTelemetrySanitizesSDKResourceEnvForOTLPExporter\|TestTelemetrySanitizesInvalidSamplerEnvBeforeProvider\|TestAllBoundEnvVarsIncludesOTLPTracesEndpoint' -count=1` | pass | SDK resource/service/sampler env sanitized around provider/resource construction; file/OTLP synthetic markers absent; process stderr empty. |
| `go test ./internal/telemetry ./internal/config ./internal/cli -run 'Telemetry\|TestLoadTelemetry\|Config' -count=1` | pass | Focused telemetry/config/CLI regression set passed. |
| `gofmt -w cmd internal`; `go vet ./...`; `go build ./cmd/pm`; `go test ./...` (SDK-env final) | pass | Final full `go test ./...` passed after earlier cold-cache harness timeouts and package segment rerun. |
| `make verify` (SDK-env final) | pass | fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, smoke, lint, and connectorgen validate passed. |
| Temp docs generation + diff | pass | `go run ./cmd/pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"`; `diff -ru docs/cli "$TMP_DOCS/cli"` clean. |
| `git diff --check`; `git diff -- go.mod go.sum` (SDK-env final) | pass | No output. |
