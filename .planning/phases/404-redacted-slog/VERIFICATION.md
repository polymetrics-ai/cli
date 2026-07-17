# VERIFICATION — Issue #404

## Required gates

| Gate | Command | Status | Notes |
|---|---|---|---|
| GSD doctor | `scripts/gsd doctor` | PASS | Adapter healthy. |
| Plan prompt | `scripts/gsd prompt plan-phase 404 --skip-research` | PASS | Prompt generated. |
| Programming loop dry-run | `scripts/gsd prompt programming-loop init --phase 404 --dry-run` | FALLBACK | Command missing from adapter; manual GSD loop recorded. |
| Format | `gofmt -w cmd internal` | pending | Run after code edits. |
| Focused race | `go test -race ./internal/logging/... ./internal/vault/... ./internal/worker/... ./internal/runtimecheck/... ./internal/cli/... -count=1` | pending | Includes no Temporal service. |
| Vet | `go vet ./...` | pending | Full repo. |
| Tests | `go test ./...` | pending | Full repo. |
| Build | `go build ./cmd/pm` | pending | CLI build. |
| Verify | `make verify` | pending | Full project gate when feasible. |
| Diff check | `git diff --check origin/feat/cli-architecture-v2...HEAD` | pending | Whitespace. |
| Dependency check | `git diff -- go.mod go.sum` | pending | Must be empty. |

## Issue-specific acceptance checks

| Check | Status | Evidence target |
|---|---|---|
| `RedactingHandler` fixed-key redaction | pending | `internal/logging` unit tests. |
| Connector SecretFields key redaction | pending | CLI logger setup + logging unit test extra key. |
| Registered-value redaction from vault.Get only | pending | `internal/vault` focused test + code grep. |
| Registry never logs/stores/returns raw values | pending | Fingerprint-only registry implementation and tests. |
| Message/attrs/nested groups/errors/URLs/query values sanitized | pending | `internal/logging` unit tests. |
| No bodies/headers/argv/secrets emitted | pending | Sensitive-key set + smoke grep. |
| Context-routed `.polymetrics/logs/<validated-run-id>.jsonl` | pending | `internal/logging` routing tests + CLI smoke. |
| Traversal/control chars/symlink escape blocked | pending | `internal/logging` routing tests. |
| 0700 log dir / 0600 log file | pending | `internal/logging` routing tests. |
| Bounded retention | pending | `internal/logging` retention test. |
| Deterministic close/no leaks | pending | handler `Close` test + race gate. |
| Warn+ fanout only to provided stderr | pending | logging unit test; stdout JSON smoke. |
| Temporal structured logger bridge | pending | static replacement + no-service tests. |
| End-to-end secret-absence hook proves fail and real log clean | pending | `internal/cli` smoke test. |
| Events/ledger/logging sibling layering preserved | pending | code review; no logs derived from events. |

## CLI parity status

CLI-visible command/flag/help/docs behavior changes: **not applicable**. This issue wires internal diagnostics only; no new user command, flag, help topic, docs page, generated manual, or website page is planned. Runtime stdout/stderr contract still verified through focused CLI smoke and existing golden/agentic contract tests.

## Runtime services

No Temporal/PostgreSQL/Dragonfly/Podman services required or started. Runtime-backed checks remain optional and are not part of this issue gate.
