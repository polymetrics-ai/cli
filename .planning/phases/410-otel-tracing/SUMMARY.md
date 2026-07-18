# Summary — Phase 410 OpenTelemetry tracing

Status: final SDK-level env hardening verified, PR body updated, and branch pushed for PR #459.

## Current state

- Parent PR #438 exists and is draft/human-gated.
- Worker branch `feat/410-otel-tracing` starts at parent head `c753071b9d6ed795cfdd80fd95f3e1c3e04792e9`.
- GSD doctor passed; plan-phase prompt generated and rerun; programming-loop prompt unavailable, manual GSD/TDD fallback active.
- Required skills loaded and recorded.
- Dependency budget restricted to ADR 0004 Stage 12 OTel trace modules at v1.44.0.

## Delivered so far

- Created issue-local GSD artifacts.
- Added `internal/telemetry` with default-off file/OTLP tracing, bounded shutdown, warning-only failures, allowlisted attributes, and `OTEL_SDK_DISABLED` override.
- Added config/env support: `PM_TELEMETRY`, `PM_TELEMETRY_DIR`, `PM_TELEMETRY_CAPTURE`, `OTEL_EXPORTER_OTLP_ENDPOINT` aliases.
- Instrumented `pm.command`, `pm.etl.run`, `pm.flow.run`, `pm.flow.step`, `pm.certify.connector`, `pm.certify.batch`, and `pm.connector.http`.
- Connector HTTP spans record method/scheme/host/path/status/attempt metadata only; tests assert no query/body/header/full URL/token leakage.
- Updated embedded help, `docs/cli/config.md`, website CLI reference, generated website docs data, and golden transcripts.

## Verification so far

- Red tests captured before production edits.
- Focused telemetry/config/CLI/connsdk/app/flow tests passed.
- File/off/secret smoke passed with synthetic marker.
- `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm` passed.
- First `make verify` run failed at `tidy-check` because go.mod/go.sum dependency changes were uncommitted.
- Green implementation slice committed/pushed, then `make verify` passed from clean dependency diff.

## Review-fix delivered (PR #459)

- Error telemetry no longer calls SDK `RecordError`; spans export only allowlisted `pm.error.type`, `pm.error.code`, and `pm.error.status_code` metadata with no `exception.*`, raw messages, registered markers, or response-body-like text.
- Config-file OTLP exporter/endpoint values are rejected/ignored unless network telemetry is explicitly enabled by env/CLI source; env opt-in remains accepted.
- File exporter directories are constrained under `--root`, reject absolute/escaping/symlinked paths and existing/symlink files, and use `0700` dirs plus `0600` `O_EXCL` files.
- Event attributes now use `trace.WithAttributes` on events; HTTP retry/attempt/status attrs remain event-scoped and allowlisted.
- OTLP export/shutdown/init failures warn through redacted `warning: telemetry:` and preserve stdout/exit code.
- Root/config help, `docs/cli/config.md`, golden transcripts, website source, and generated website docs data are updated.

Current execution decision: `local_critical_path`. This worker stayed on `feat/410-otel-tracing`; coordinator sidecars/human fallback handle review coverage. No Claude/Copilot request from this worker.

## Final focused review-fix delivered

- Ambient OTel env bypass fixed: `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` is bound/validated before exporter construction; unsafe endpoint values warn through project stderr, do not reach process stderr, and do not contact unvalidated collectors.
- Unsupported ambient OTLP headers/TLS/compression/protocol/timeout env are warned and neutralized before `otlptracehttp.New`; exporter construction always passes a validated endpoint/default and empty headers.
- Docs residuals fixed: ADR 0004 superseding note, root help/goldens exporter values `none/off/file/otlp`, config docs/website trusted-env endpoint wording, generated website data.
- Red tests captured before production edits; no Claude/Copilot request.

## Final SDK-level env hardening delivered

- SDK/provider/resource ambient OTel env is warned by name only and temporarily unset around explicit safe resource plus `sdktrace.NewTracerProvider` construction.
- Covered env includes `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, `OTEL_TRACES_SAMPLER(_ARG)`, span/attribute/link/event limits, BSP limits, and experimental Go OTel toggles.
- File exporter and OTLP regressions prove synthetic `api_key` resource attrs/service markers do not export; invalid sampler args no longer hit raw process stderr.
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` is in config test env cleanup/list.
- No Claude/Copilot request.

## Verification final

Focused tests, docs/golden/website generation, file/off/secret smoke, OTLP endpoint smoke, `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` passed after the previous review-fix commit. Final focused review-fix verification also passed: focused ambient env tests, docs generation/diff, website data generation, runtime help parity, full Go gates, `make verify`, `git diff --check`, and `git diff -- go.mod go.sum`.

SDK-level env hardening verification also passed: new focused red/green tests, `go test ./internal/telemetry ./internal/config ./internal/cli -run 'Telemetry|TestLoadTelemetry|Config' -count=1`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, temp docs generation/diff, `git diff --check`, and `git diff -- go.mod go.sum`.
