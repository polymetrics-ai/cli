# Verification — API → API GitHub route proof

Status: in progress.

Commands and results are appended here as they are run. Full-suite execution is
left to CI under the repository's per-command-timeout rule; all local checks
are scoped and executed separately with `-timeout 20m` where applicable.

## Planned gates

- `go test -timeout 20m ./internal/synctransport`
- `go test -timeout 20m ./internal/app`
- `go test -timeout 20m ./internal/cli`
- `go test -timeout 20m -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$' ./internal/cli`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make agent-contract-check`
- `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, `make release-workflow-check`,
  `make connector-runtime-preflight`, and `make connector-canon-check`
- `scripts/verify-gsd-workflow 2c48e4deb34128339fccbe5d4b7daad4e13a23e7`
- fresh-binary controlled GitHub run and independent read-back
