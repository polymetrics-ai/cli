# Verification: connector certification foundation G1/G2/G6

## Required checks

- [ ] Focused `cmd/connectorgen` classification, sweep, proof, and evidence tests (including race/concurrent reader tests).
- [ ] `go run ./cmd/connectorgen certification-sweep --connector github --check`
- [ ] `go run ./cmd/connectorgen certification-matrix --check`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] Changed package tests and `internal/cli` in separate timeout-bounded commands.
- [ ] `go vet ./...`, `go build ./cmd/pm`, generated/snapshot checks, and repository verification gates.
- [ ] `git diff --check`; automated review and PR API base read-back.

## Status

Pending implementation.
