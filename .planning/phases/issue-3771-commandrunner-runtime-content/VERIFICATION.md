# Verification — #3771 command-runner runtime content

## Planned checks

- [x] `gofmt -w` on changed Go files
- [x] focused `internal/connectors/commandrunner` tests — #3782 and #3790 selections passed
- [x] focused `internal/app` tests — `TestPlanConnectorCommandPersistsCompleteDeclaredContent` passed
- [x] focused `internal/cli` tests — policy test and regenerated `TestGoldenTranscripts` passed
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
