# Verification — PR #4308 status-check result preservation

## Checklist

- [x] Remote PR #4308 author, branch, and exact head `4712dd03b6469764b37e241732e8b07d91622ba7` re-read before branch mutation.
- [x] Completed scout scratch inventory confirmed no intended production change; moved the untracked 403 MB scratch directory to Trash rather than committing or deleting it.
- [x] Required GSD command sources, adapter doctor, and canonical contract check completed.
- [x] Required skills and CLI parity reference loaded; the parity assessment is recorded in `PLAN.md`.
- [x] Record focused red result-boundary failure.
- [x] Record focused green JSON, human, classification, and binary-regression results.
- [x] Rerun source-locked loader negative tests and installed CLI live HEAD/GET proof.
- [x] Run format, changed-package tests, vet, build, lint, and applicable non-suite gates. `golangci-lint run ./internal/cli/...` reports seven pre-existing unrelated findings; `golangci-lint run --new-from-rev=HEAD ./internal/cli/...` and the repository `make lint` gate both pass.
- [x] Review remediation keeps final non-2xx status probes typed through requester, engine, runner, and CLI output while retaining the binary-download error path; focused regression command passed.
- [ ] Complete no-mistakes validation on the pushed current head.
- [ ] Update current PR title/body, push same branch, verify GitHub API base and current-head CI.
- [ ] Write stand-alone external `ship-report.md` with command, commit, live-output, cleanup, and limitation evidence.

## Captured results

- Focused CLI: `go test -count=1 -timeout 20m ./internal/cli -run '^TestWriteConnectorCommandResult'` — pass after the restored generic branch; its behavioral red is in `TDD-LEDGER.md`.
- Loader closure: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestBundleLoadRegistersStatusAndTextExportOperations|TestBundleLoadRejectsInvalidStatusAndTextExportDeclarations)$'` — pass.
- Live installed binary: JSON returned `ConnectorCommandStatusCheck` with connector, command, operation, `HEAD`, path, `200`, and `body_bytes: 0`; human returned the same values in one non-empty deterministic line. Export returned 48,219 bytes and SHA-256 `0845078a290b48e3149ab8639966824110a251db4e06fc144c06ebb534af23be`; it is byte-identical to the raw `vega/vega-datasets@dedfc126e87dfde2df0332744689844314911d5d` file.
- Local gates: `gofmt`, `git diff --check`, `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`, changed-line CLI lint, `make docs-check-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`, and `make smoke-no-build` — pass.
- Review remediation: `go test -count=1 -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -run '^(TestRequesterDoStatusCheckPreservesFinalNon2xxResponse|TestOperationStatusCheckUsesDeclaredHEADWithoutJSONBody|TestOperationStatusCheckPreservesFinalNon2xxStatus|TestRunStatusCheckPreservesFinalNon2xxMetadata|TestWriteConnectorCommandResultPreservesStatusCheckJSON|TestWriteConnectorCommandResultPreservesStatusCheckHumanOutput|TestWriteConnectorCommandResultPreservesBinaryDownloadEnvelope|TestBinaryDownloadPreservesHTTPErrorTextAndLeavesNoFile)$'` — pass.
