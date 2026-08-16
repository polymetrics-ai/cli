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

## Executed before live I/O

| Command | Result |
| --- | --- |
| `go test -timeout 20m -run '^(TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead|TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess|TestOrchestratorAdmitsEmptyResultOnlyFromExplicitSourceMarker|TestOrchestratorCommitsOnlyAfterDurableAcknowledgement)$' ./internal/synctransport` | PASS — 0.522s |
| `go test -timeout 20m -run '^TestIssueLabelTransport' ./internal/app` | PASS — 15.147s |
| `go test -timeout 20m -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$' ./internal/cli` | PASS — fresh binary lifecycle test (cached confirmation: 1.533s) |

## Live-proof decision gate

Blocked before any PM credential injection or provider write. The authoritative
runbook says the recorded fine-grained token is revoked (`HTTP 401`) and must
not be re-tested. The live GitHub App installation is reported working, but no
controlled repository is available to it. A narrowly named run-owned repository
creation attempt through `gh-axi` was refused by GitHub with
`karthik-sivadas does not have the correct permissions to execute
CreateRepository`; no repository, issue, label, or PM project was created.

Required decision: provide an existing run-owned repository to which the
GitHub App is installed, or grant the current GitHub identity organization
repository-creation permission (`CreateRepository`) in `Polymetrics-Cert`.
The proof additionally needs the App's existing `Issues: write` grant on that
repository. This lane must not use the revoked token or a third-party repository.
