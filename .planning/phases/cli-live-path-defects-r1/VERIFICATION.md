# Verification: live-path defects r1

## Required checks

- [ ] Focused #4119 requester/rate-admission tests, including zero-send refusal
  and formatter typed-error preservation.
- [ ] `go test -timeout 20m ./internal/coordination/...` for #4125 bounds.
- [ ] Focused #4169 real CLI construction-path and classification tests.
- [ ] `go test -timeout 20m ./internal/cli/...`.
- [ ] `go vet ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs` and
  `go run ./cmd/connectorgen surface-sync --check` when the final diff touches
  connector definitions or generator-derived surfaces; otherwise record N/A.
- [ ] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connector-boundary`, and
  `make release-workflow-check` individually, per the scoped verification rule.
- [ ] `git diff --check`.
- [ ] Inline `verify-work` and `code-review` evidence with all actionable
  findings fixed or explicitly dispositioned.
- [ ] Push direct PR and confirm the API-reported base is
  `integration/4015-mvp-flat-r1`.

## CLI parity disposition

- N/A: no command, flag, help topic, output schema, manual, website page, or
  generated help artifact changes. The typed provider-error classification is
  proven via focused CLI/binary tests instead.
