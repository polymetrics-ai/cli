# SUMMARY — Issue #404

Status: implementation pushed; PR #455 open; non-race/full verify green; issue race command blocked by existing slow `internal/cli` race suite timeout.

PR: https://github.com/polymetrics-ai/cli/pull/455
Head: `f22757ce47b95585c5148fc1b518a013c21e609b`

## Delivered

- Added `internal/logging` redacting slog foundation:
  - outer `RedactingHandler`;
  - fixed + dynamic sensitive-key redaction;
  - fingerprint-only registered-value registry;
  - message/attr/group/error/URL/query sanitization;
  - context run-id logger helpers;
  - per-run JSONL handler with run ID validation, `os.Root` symlink/traversal defense, 0700 dirs, 0600 files, retention pruning, deterministic close;
  - warn+ fanout to provided stderr.
- Fed registered-value redaction at `vault.Get` only.
- Added ETL run start/complete/fail info logs beside events; no logs derived from events.
- Replaced Temporal `noopLogger` seams with pinned `tlog.NewStructuredLogger(pmlogging.FromContext(ctx))`.
- Added focused tests and hermetic CLI log smoke with scanner-failure proof and real-log clean proof.

## Verification

PASS:

- `go test ./internal/logging/... ./internal/vault/... ./internal/cli/... -run 'TestRedactingHandler|TestRegisterSensitiveKey|TestRunFileHandler|TestRunLogger|TestTemporalStructuredLogger|TestVaultGetRegistersValuesForRedaction|TestRedactedRunLogsSmoke' -count=1`
- `go test -race ./internal/logging/... ./internal/vault/... ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... -count=1`
- `go test ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... -count=1`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`

BLOCKED:

- `go test -race ./internal/logging/... ./internal/vault/... ./internal/worker/... ./internal/runtimecheck/... ./internal/cli/... -count=1` timed out because `./internal/cli` exceeds the Go test timeout while repeatedly loading connector bundles.
- `go test -race ./internal/cli/... -count=1 -timeout=20m` also timed out, later in docs generation.

## Review route

Claude disabled / Copilot quota exhausted per assignment. Human/parent fallback pending; no automated review request made.
