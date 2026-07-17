# VERIFICATION — Issue #404

## Required gates

Review-fix command set (requested by coordinator):

```bash
gofmt -w cmd internal
go test -race ./internal/logging/... ./internal/vault/... ./internal/app/... ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... ./internal/connectors/connsdk/... -count=1
go test ./internal/cli/... -run 'TestRedactedRunLogsSmoke|Logging|Error|JSON' -count=1
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Extended full CLI race remains pending for coordinator and must not be run by this worker. `verificationPassed=false` until that coordinator gate is complete.

| Gate | Command | Status | Notes |
|---|---|---|---|
| GSD doctor | `scripts/gsd doctor` | PASS | Adapter healthy. |
| Plan prompt | `scripts/gsd prompt plan-phase 404 --skip-research` | PASS | Prompt generated. |
| Programming loop dry-run | `scripts/gsd prompt programming-loop init --phase 404 --dry-run` | FALLBACK | Command missing from adapter; manual GSD loop recorded. |
| Format | `gofmt -w cmd internal` | PASS | Also run by `make verify`. |
| Focused race | `go test -race ./internal/logging/... ./internal/vault/... ./internal/worker/... ./internal/runtimecheck/... ./internal/cli/... -count=1` | FAIL/BLOCKED | Exact command timed out in `./internal/cli` at Go default 10m; `./internal/cli -timeout=20m` also timed out. Non-CLI focused race packages passed. No Temporal service used. |
| Vet | `go vet ./...` | PASS | Exited 0. |
| Tests | `go test ./...` | PASS | Exited 0; `internal/cli` 174.986s, `internal/connectors/certify` 346.816s. |
| Build | `go build ./cmd/pm` | PASS | Exited 0. |
| Verify | `make verify` | PASS | Exited 0; includes `go test -timeout 20m ./...`, docs validate, smoke, lint, connectorgen validate. |
| Diff check | `git diff --check origin/feat/cli-architecture-v2...HEAD` | PASS | Exited 0. |
| Dependency check | `git diff -- go.mod go.sum` | PASS | Empty after `make verify`/`go mod tidy`. |
| Review-fix focused red tests | see TDD ledger T8-T10 | PASS | Red captured before production edits; green focused packages pass. |
| Review-fix format | `gofmt -w cmd internal` | PASS | Exited 0. |
| Review-fix focused race | `go test -race ./internal/logging/... ./internal/vault/... ./internal/app/... ./internal/worker/... ./internal/runtimecheck/... ./internal/temporalprobe/... ./internal/connectors/connsdk/... -count=1` | PASS | Exited 0; no services. |
| Review-fix CLI focused | `go test ./internal/cli/... -run 'TestRedactedRunLogsSmoke|Logging|Error|JSON' -count=1` | PASS | Exited 0. |
| Review-fix vet | `go vet ./...` | PASS | Exited 0. |
| Review-fix tests | `go test ./...` | PASS | Exited 0 after connsdk body classification fix. |
| Review-fix build | `go build ./cmd/pm` | PASS | Exited 0. |
| Review-fix verify | `make verify` | PASS | Exited 0. |
| Review-fix diff check | `git diff --check origin/feat/cli-architecture-v2...HEAD` | PASS | Exited 0. |
| Review-fix dependency check | `git diff -- go.mod go.sum` | PASS | Empty. |
| Extended full CLI race | not run | PENDING/COORDINATOR | Explicitly deferred by coordinator; `verificationPassed=false`. |

## Issue-specific acceptance checks

| Check | Status | Evidence target |
|---|---|---|
| `RedactingHandler` fixed-key redaction | PASS (focused) | `internal/logging` unit tests. |
| Connector SecretFields key redaction | PASS (focused) | CLI logger setup + logging unit extra-key test. |
| Registered-value redaction from vault.Get only | PASS (focused) | `internal/vault` focused test. |
| Registry never logs/stores/returns raw values | PASS (focused) | Fingerprint-only registry; no raw-value accessors. |
| Message/attrs/nested groups/errors/URLs/query values sanitized | PASS (focused) | `internal/logging` unit tests. |
| No bodies/headers/argv/secrets emitted | PASS (focused) | Sensitive-key set + smoke scanner. |
| Context-routed `.polymetrics/logs/<validated-run-id>.jsonl` | PASS (focused) | `internal/logging` routing tests + CLI smoke. |
| Traversal/control chars/symlink escape blocked | PASS (focused) | `internal/logging` routing tests. |
| 0700 log dir / 0600 log file | PASS (focused) | `internal/logging` routing tests. |
| Bounded retention | PASS (focused) | `internal/logging` retention test. |
| Deterministic close/no leaks | PASS (focused) | handler `Close` test; race pending. |
| Warn+ fanout only to provided stderr | PASS (focused) | logging unit test; stdout JSON smoke. |
| Temporal structured logger bridge | PASS (focused) | `tlog.NewStructuredLogger(pmlogging.FromContext(ctx))`; no-service tests. |
| End-to-end secret-absence hook proves fail and real log clean | PASS (focused) | `internal/cli` smoke test. |
| Events/ledger/logging sibling layering preserved | PASS (code review) | ETL logs added beside events; no logs derived from events. |

## CLI parity status

CLI-visible command/flag/help/docs behavior changes: **not applicable**. This issue wires internal diagnostics only; no new user command, flag, help topic, docs page, generated manual, or website page is planned. Runtime stdout/stderr contract still verified through focused CLI smoke and existing golden/agentic contract tests.

## Runtime services

No Temporal/PostgreSQL/Dragonfly/Podman services required or started. Runtime-backed checks remain optional and are not part of this issue gate.
