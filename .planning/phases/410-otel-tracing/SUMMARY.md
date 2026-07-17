# Summary — Phase 410 OpenTelemetry tracing

Status: review-fix planned for PR #459; verification pending rerun.

## Current state

- Parent PR #438 exists and is draft/human-gated.
- Worker branch `feat/410-otel-tracing` starts at parent head `c753071b9d6ed795cfdd80fd95f3e1c3e04792e9`.
- GSD doctor passed; plan-phase prompt generated; programming-loop prompt unavailable, manual GSD/TDD fallback active.
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

## Review-fix queue (PR #459)

Accepted review findings target error attr safety, config-sourced OTLP guard, file exporter path/symlink hardening, event attr placement, OTLP failure neutrality, endpoint validation, help/docs exporter wording, and warning redaction/stdout discipline.

Current execution decision: `local_critical_path`. This worker stays on `feat/410-otel-tracing`; coordinator sidecars/human fallback handle review coverage. No Claude/Copilot request from this worker.

## Next

1. Add review-fix red tests and capture exact failing output.
2. Implement minimal fixes.
3. Run focused gates, smoke/parity checks, full gates, update PR #459 body, commit and push.
