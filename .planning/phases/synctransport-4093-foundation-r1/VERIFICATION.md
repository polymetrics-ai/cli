# Verification — Refs #4093

## Acceptance evidence

| Acceptance | Evidence | Result |
| --- | --- | --- |
| Strict malformed/unknown loading and clone-safe projection | `TestBundleLoadSyncTransportProjectsIndependentDefinition`, `TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations` | pass |
| Definition-selected evidence, no foreign admission | `TestDefinitionConformanceVerifierAcceptsEvidenceSelectedByEachDefinition`, `TestDefinitionConformanceVerifierRefusesAlteredEvidenceBeforeSourceIO` | pass |
| A second declaration reuses the production source without dispatch edits | `TestDefinitionTransportFactoriesRegisterSharedSourceOnce` | pass; one shared executor preflights a loaded synthetic bundle with its distinct evidence |
| Production declarations retain GitHub/PostgreSQL registration | `TestOpenRegistersDefinitionOwnedProductionTransports` and `TestDefinitionTransportFactoriesSelectDeclaredEvidence` | pass |
| Wrong role and source binding fail before I/O | `TestDestinationTransportDescriptorRefusesChangeCaptureDestinationMode`, `TestTransportFamilyHalfPathConformanceRefusesChangeCaptureDestinationBeforeIO`, `TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration`, `TestRunETLTransportRefusesDeclaredChangeCaptureDestinationBeforeIO` | pass; `change_capture` is source-only into the connection warehouse |
| Kill-after-commit and owned-stage bounds | `TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit`, `TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages` | pass |
| Existing GitHub execution route | `TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle` | pass |
| Live GitHub Gate A: fresh binary through neutral definition composition | `TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels` with the approved local Docker endpoint and an ephemeral `gh auth token` environment value | pass: append, keyed set, and keyed replay each report `completed`, `records_read=1`, `records_loaded=1`; each durable acknowledgement checkpoint and independent post-exit GitHub label read-back match |
| Mechanical authoring guide | `docs/sync-transport-definition.md` reviewed by `make docs-check` | pass |

## Commands

All commands below exited zero unless stated otherwise.

```text
scripts/gsd doctor
scripts/gsd sources discuss-phase
scripts/gsd sources plan-phase
scripts/gsd sources execute-phase
scripts/gsd sources verify-work
scripts/gsd sources code-review
scripts/gsd prompt discuss-phase 4093
scripts/gsd prompt plan-phase 4093 --tdd
scripts/gsd prompt execute-phase 4093
scripts/gsd prompt verify-work 4093
scripts/gsd prompt code-review 4093
go run ./cmd/agentcontractgen check

# RED
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestDefinitionConformanceVerifierAcceptsEvidenceSelectedByEachDefinition$'
# failed to compile: DefinitionFactory has no AcceptedSourceEvidences
go test -count=1 -timeout 20m ./internal/app -run '^TestDefinitionTransportFactoriesSelectDeclaredEvidence$'
# failed to compile: definitionTransportDefinitionFactories is undefined
go test -count=1 -timeout 20m ./internal/app -run '^TestDefinitionTransportFactoriesRegisterSharedSourceOnce$'
# failed: declarative_stream_source is registered more than once
go test -count=1 -timeout 20m ./internal/connectors -run '^TestDestinationTransportDescriptorRefusesChangeCaptureDestinationMode'
# failed: Validate() = <nil>, want destination change_capture role refusal

# GREEN and regression proof
go test -count=1 -timeout 20m ./internal/app -run '^(TestDefinitionTransportFactoriesSelectDeclaredEvidence|TestDefinitionTransportFactoriesRegisterSharedSourceOnce|TestOpenRegistersDefinitionOwnedProductionTransports|TestAppCompositionRoutesLoadedSyntheticDefinitionConnector)$'
go test -count=1 -timeout 20m ./internal/app -run '^(TestOpenRegistersDefinitionOwnedProductionTransports|TestRunETLTransportRefusesDeclaredChangeCaptureDestinationBeforeIO|TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages|TestDefinitionTransportFactoriesSelectDeclaredEvidence|TestDefinitionTransportFactoriesRegisterSharedSourceOnce)$'
go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors ./internal/synctransport -run '^(TestBundleLoadSyncTransportProjectsIndependentDefinition|TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations|TestDestinationTransportDescriptorRefusesChangeCaptureDestinationMode|TestDestinationTransportDescriptorRefusesChangeCaptureDestinationModeRegardlessOfStrategy|TestTransportFamilyHalfPathConformanceRefusesChangeCaptureDestinationBeforeIO|TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration|TestDefinitionConformanceVerifierRefusesAlteredEvidenceBeforeSourceIO|TestDefinitionConformanceVerifierAcceptsEvidenceSelectedByEachDefinition|TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit)$'
go test -count=1 -timeout 20m ./internal/app
go test -race -count=1 -timeout 20m ./internal/synctransport ./internal/app
go test -count=1 -timeout 20m ./internal/cli
go test -count=1 -timeout 20m ./internal/cli -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$'
go test -count=1 -timeout 20m ./cmd/connectorgen
go test -count=1 -timeout 20m ./internal/synctransport ./internal/connectors ./internal/connectors/engine ./internal/app
go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors ./internal/synctransport -run '^(TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations|TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration|TestDestinationTransportDescriptorRefusesChangeCaptureDestinationMode|TestDestinationTransportDescriptorRefusesChangeCaptureDestinationModeRegardlessOfStrategy|TestTransportFamilyHalfPathConformanceRefusesChangeCaptureDestinationBeforeIO|TestDefinitionConformanceVerifierAcceptsEvidenceSelectedByEachDefinition|TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit)$'
go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportRefusesDeclaredChangeCaptureDestinationBeforeIO|TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages|TestDefinitionTransportFactoriesSelectDeclaredEvidence|TestDefinitionTransportFactoriesRegisterSharedSourceOnce)$'
go vet ./...
go build ./cmd/pm
git diff --check

# Generated, boundary, and repository gates
go run ./cmd/connectorgen validate
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen boundary . --json
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
make connector-canon-check
make connectorgen-certification-matrix
make github-parity-artifacts-check

# Live GitHub Gate A (red: retained proof label exposed a non-rerunnable empty baseline)
POLYMETRICS_GITHUB_TOKEN="$(gh auth token)" POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_GITHUB_ISSUE_LABEL_LIVE_PROOF=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli -run '^TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels$'
# failed before mutation: issue 1 retained [pm-db-api-live-add], test wanted exact []

# Live GitHub Gate A (green; a fresh pm binary is built by the test)
POLYMETRICS_GITHUB_TOKEN="$(gh auth token)" POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_GITHUB_ISSUE_LABEL_LIVE_PROOF=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli -run '^TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels$'
# PASS 43.29s: add, keyed set, keyed replay each status=completed records_read=1 records_loaded=1 acknowledgement=checkpoint-recorded read_back=matched
```

`golangci-lint run ./internal/app/... ./internal/synctransport/...` was also
run. It exits nonzero on nine existing staticcheck findings in Arrow fast-path
files and `internal/app/arrow_segment_store.go`, none touched by this branch;
the repository's required `make lint` gate passes. The full `go test ./...`
suite is deliberately not run as one task-runner command: repository guidance
states that it routinely exceeds this runner's command window and requires CI
for full-suite coverage. Changed packages, `internal/cli`, and every
non-test `make verify` gate above were run locally.
