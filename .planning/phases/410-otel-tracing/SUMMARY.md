# Summary — Phase 410 OpenTelemetry tracing

Status: review-fix verified locally for PR #459; push/PR body update pending.

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

## Verification final

Focused tests, docs/golden/website generation, file/off/secret smoke, OTLP endpoint smoke, `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` passed after the review-fix commit.
