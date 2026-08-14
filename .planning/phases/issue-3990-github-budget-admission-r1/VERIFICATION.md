# Verification: Issue #3990

## Required checks

- [ ] `go test -timeout 20m ./internal/connectors/engine/...`
- [ ] `go test -timeout 20m ./internal/coordination/...`
- [ ] `go test -timeout 20m ./internal/connectors/certify/...`
- [ ] Focused multi-process shared-budget proof with the provider-double second-send count equal to zero.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `git diff --check`
- [ ] `go vet ./...`
- [ ] no-mistakes pipeline produces an open PR and green CI.

## Live evidence rule

Every failing-admission test asserts its physical provider-double request count is zero. Every success case asserts the exact request/event/ledger state claimed; no test treats an error-free return as proof.

## Out-of-scope finding protocol

Append `needs-decision:` to this record for any unrelated review finding rather than editing it in this branch. Known exclusion: #4125, `internal/coordination/shared_rate_limits.go` duration overflow.
