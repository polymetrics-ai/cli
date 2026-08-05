# Verification — #3771 command-runner runtime content

## Planned checks

- [ ] `gofmt -w` on changed Go files
- [ ] focused `internal/connectors/commandrunner` tests
- [ ] focused `internal/app` tests
- [ ] focused `internal/cli` tests and command help checks
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] final diff/ownership inspection
- [ ] inline GSD verification and code-review fallback recorded

The full `go test ./...` and aggregate `make verify` are intentionally left to CI under this
repository's timeout guidance.
