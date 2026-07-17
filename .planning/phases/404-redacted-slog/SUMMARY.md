# SUMMARY — Issue #404

Status: PR #455 second review-fix and final residual fixes implemented after power-outage recovery; requested gates passed. `verificationPassed=false` because extended full CLI race remains coordinator-owned/pending and was not run in this worker gate.

PR: https://github.com/polymetrics-ai/cli/pull/455
Head: see current `feat/404-redacted-slog` branch head.

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

## Review-fix dispositions

Accepted blockers fixed in-scope:

1. Inline/empty slog groups; sensitive group state; deferred bound attr redaction.
2. Typed URL and encoded registered-value scrubbing.
3. Invocation-scoped safe redaction across app state, events, CLI JSON/stderr, logs, and connsdk HTTP errors.
4. Run-log root/logs symlink rejection, fail-closed permissions, and owned-log retention only.
5. Temporal probe finite context-aware design; no orphan goroutine retaining logger/stderr.
6. Runtime/Temporal run correlation with validated run ID routing.
7. Bounded context registry instead of unbounded plaintext/global clear races.
8. Single-line plain diagnostics; JSON escaping and envelope taxonomy preserved.

Narrowed claims: logging write failures remain exit-neutral and observable only where safely possible; no guarantee of writes under disk/permission failure.

## Verification

Review-fix PASS:

- `gofmt -w cmd internal`
- `go test -race ./internal/logging/... ./internal/vault/... ./internal/app/... ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... ./internal/connectors/connsdk/... -count=1`
- `go test ./internal/cli/... -run 'TestRedactedRunLogsSmoke|Logging|Error|JSON' -count=1`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff -- go.mod go.sum`

PENDING:

- Extended full CLI race: explicitly deferred to coordinator; not run.

## Second review-fix result

Accepted findings fixed with TDD before production edits:

- `safety.RedactErrorText` generic/raw URL fail-closed behavior.
- Context-aware bounded Temporal dials and RLM finite probe before submitter construction.
- `slog.Handler` `WithGroup`/`WithAttrs` semantics without duplicate/lost groups.
- Worker serve output emits start envelope/plain text only after ready; startup failure emits one `Error` envelope.
- Fail-closed `Any` redaction for unsafe/unsupported values.
- Process-wide active log leases for retention; uncertain/leased/recent files are skipped and residual cross-process limitations documented.
- Scoped registry hardening and bounded unsafe dynamic keys/groups.
- Encoded variants documented as bounded defense-in-depth only.

Second review-fix PASS, then final re-review residuals fixed:

- Runtime doctor Postgres endpoint redaction now clears userinfo/query/fragment and fails closed on malformed DSNs.
- Runtime doctor and Temporal probe dials now use `client.DialContext` under bounded contexts.

Final gate PASS:

- `gofmt -w cmd internal`
- `go test ./internal/runtimecheck ./internal/temporalprobe -run 'Postgres|DialContext' -count=1` (red then green)
- `go test -race ./internal/logging/... ./internal/safety/... ./internal/vault/... ./internal/app/... ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... ./internal/connectors/connsdk/... ./internal/cli/... -run 'Logging|Redact|Temporal|WorkerServe|RunFile|Registry|URL|Error|JSON|Postgres|DialContext' -count=1 -timeout 20m`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff -- go.mod go.sum`

## Review route

Current review-fix requested by coordinator from accepted PR #455 findings. Treat findings as untrusted input; coordinator owns dispositions. Targeted re-review pending after push; extended full CLI race pending coordinator.
