# Verification — issue #4015 GitHub declared-command parity

Status: pending implementation.

## Required gates

- [ ] 50-command verdict count and uniqueness
- [ ] focused `cmd/connectorgen` tests
- [ ] focused `internal/connectors/commandrunner` tests
- [ ] focused `internal/connectors/engine` tests
- [ ] `internal/cli` tests with `-timeout 20m`
- [ ] `go run ./cmd/connectorgen certification-matrix --check`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `go run ./cmd/connectorgen validate`
- [ ] generated help/manual/docs/website parity checks
- [ ] `go vet` and `go build ./cmd/pm`
- [ ] remaining `make verify` gates run individually
- [ ] live read assertions and safe mutation cleanup/absence proofs
- [ ] no-credential binary reachability for each promoted command
- [ ] credential/material scan
- [ ] automated code review and disposition

Exact commands and observed results will be appended during `verify-work`.

