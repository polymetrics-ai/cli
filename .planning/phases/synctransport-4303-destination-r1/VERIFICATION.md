# Refs #4303 — Verification

## Result

Passed locally. No checks were skipped and `/no-mistakes` was not run.

The application-dispatch reconciliation and complete-result-output slice also
passed locally. The first full run identified only generated CLI golden
transcript drift; after regenerating that repository-owned snapshot, the
second full `make verify` run exited `0`.

## Focused TDD checks

- `go test -count=1 -timeout 20m ./internal/app -run '^TestDefinitionTransportFactories(RunTypedDestinationFromDefinition|SelectDistinctTypedDestinationEvidence|RefuseTypedDestinationDeclarationsBeforeIO)$'` — passed.
- `go test -count=1 -timeout 20m ./internal/app -run '^Test(DefinitionTransportFactories(RunTypedDestinationFromDefinition|SelectDistinctTypedDestinationEvidence|RefuseTypedDestinationDeclarationsBeforeIO)|DeclarativeTypedDestinationRefusesInvalidWorksetsBeforeProviderWrite)$'` — passed.
- `go test -count=1 -timeout 20m ./internal/app -run '^Test(OpenRegistersDefinitionOwnedProductionTransports|DefinitionTransportFactories(SelectDeclaredEvidence|RegisterSharedSourceOnce|RunTypedDestinationFromDefinition|SelectDistinctTypedDestinationEvidence|RefuseTypedDestinationDeclarationsBeforeIO))$'` — passed.

## Review remediation check

- `go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/app -run '^(TestBundleLoadRejectsDuplicateWriteActionNames|TestOperationDirectWriteHonorsDeclaredJSONAndNoneResponsePolicies|TestValidateRecordSchemaFieldMappingRequiresExactCompleteFields|TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields|TestDeclarativeTypedDestinationPreflightRejectsIncompleteMappingAndFullOverwriteBeforeIO|TestDirectWriteCommandHonorsDeclaredJSONAndNoneResponsePolicies)$'` — passed. It covers exact raw selected-action property names, required mapping coverage before I/O, default full-overwrite refusal, duplicate definition names, and bodyless `none` responses retaining status and provider receipt headers.

## Affected-package and binary checks

- `go test -count=1 -timeout 20m ./internal/app` — passed (249.108s).
- `go test -count=1 -timeout 20m ./internal/connectors` — passed.
- `go test -count=1 -timeout 20m ./internal/connectors/engine` — passed.
- `go test -count=1 -timeout 20m ./internal/synctransport` — passed.
- `go test -count=1 -timeout 20m ./internal/cli` — passed (529.662s).
- `go vet ./...` — passed.
- `go build ./cmd/pm` — passed.

## Application dispatch and result-output checks

- `go test -count=1 -timeout 20m ./internal/app -run '^TestPersistedConnectionSelectsDeclarativeTypedDestinationAction$'` — passed. The production-shaped synthetic bundles prove two named actions in one connector and one in another are selected only by the persisted `destination_action`; missing, foreign, unlisted, and stale selections refuse before I/O. The same test proves persisted `Run.destination_results` preserves ordinary response status, headers, nested fields, large numeric values, and tier-specific fields while masking concrete configured credential material; separately generated diagnostics remain secret-taint-safe.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields$'` — passed. Exact schema-valid provider-owned names, including snake_case, camelCase, and `http`, are admitted only when the selected action declares them; whitespace, malformed, unknown, cross-action, runtime-selected, and undeclared names refuse before I/O.
- `go test -count=1 -timeout 20m ./internal/connectors -run 'TestSanitize.*Output'` — passed. Complete output retains provider-returned structure and ordinary values while masking concrete configured credential material; separately generated diagnostics remain secret-taint-safe.
- `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteHonorsDeclaredJSONAndNoneResponsePolicies$'` — passed. `output_policy: none` retains complete non-empty output and accepts an intentional bodyless success without losing status or headers.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestDirectWriteCommandHonorsDeclaredJSONAndNoneResponsePolicies$'` — passed.
- `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'` regenerated the tracked CLI snapshots; the same command without the update variable then passed.

## GitHub real-provider parity

With the explicit local Docker socket and an authenticated GitHub token held
only in the test process environment:

`go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli -run '^TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels$'`

passed (40.589s). The fresh `pm` binary completed add, set, and keyed-replay
runs against `karthik-sivadas/pm-parity-proof-db-to-api`; independent GitHub
API reads confirmed the retained labels after each acknowledgement and
checkpoint.

## Repository gates

- Detached, polled `make connector-boundary` — passed: `outcome: clean`, 552
  connectors, 293 files, zero findings.
- `make verify` — passed. This includes formatting/tidy, `go vet ./...`, the
  complete `go test -timeout 20m ./...` suite, build, docs validation, smoke,
  lint, agent-contract validation, generated parity/certification checks,
  connector-boundary, connector canon, and release-target checks.

The final full run reported `ConnectorBoundaryReport.outcome: clean`, 552
connectors loaded, 294 files checked, and zero findings.

## 2026-08-20 reconciliation revalidation

The published reconciliation head was revalidated locally before rollup. No
new code change was required: the persisted App dispatch and the closed
definition-owned boundary remained intact under a fresh run.

- `go test -count=1 -timeout 20m ./internal/app -run '^(TestPersistedConnectionSelectsDeclarativeTypedDestinationAction|TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields|TestDefinitionTransportFactories(RunTypedDestinationFromDefinition|SelectDistinctTypedDestinationEvidence|RefuseTypedDestinationDeclarationsBeforeIO)|TestDeclarativeTypedDestinationRefusesInvalidWorksetsBeforeProviderWrite)$'` — passed.
- `go test -count=1 -timeout 20m ./internal/connectors -run '^(TestDestinationTransportDescriptorSelectsPersistedActionWithinMode|TestSanitize.*Output)$'` — passed.
- `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteHonorsDeclaredJSONAndNoneResponsePolicies$'` — passed.
- `go test -count=1 -timeout 20m ./internal/cli -run '^(TestETLTransportBareAndLeafHelpAreContextual|TestDeclarativeTypedDestinationTransportRejectsCallerActionBeforeProjectIO|TestGoldenTranscripts)$'` — passed.
- Detached and polled `make connector-boundary` — passed: `outcome: clean`, 552 connectors, 294 files, no findings or warnings. The full result is retained at `traces/connector-boundary-rerun-20260820.log`.
- `make verify` — passed locally. The complete result is retained at `traces/make-verify-rerun-20260820.log`; it includes the full test tree, docs, smoke, lint, generator, certification, connector-boundary, canon, and release-target gates.
- `git diff --check main...HEAD` — passed.
- `gh api /repos/polymetrics-ai/cli/pulls/4304 --jq .base.ref` — returned `main`, matching the task delivery header.

## Follow-up foundation r1 checklist

- [x] Action-owned mapping schema and loader validation: three same-source actions and one other connector action pass only their exact persisted action bindings; missing and cross-selected bindings refuse before I/O. Focused command: `go test -count=1 -timeout 20m ./internal/app -run '^(TestDeclarativeTypedDestinationRequiresActionOwnedSourceBindings|TestDeclarativeTypedDestinationBindsThreeActionsAndSeparateConnectorBeforeIO)$'` — passed.
- [x] Batch boundaries and partial receipts: `TestDeclarativeTypedDestinationBatchLimitIsDefinitionOwned` proves normal and tombstone worksets above their declared two-record limit refuse before provider I/O; `TestDeclarativeTypedDestinationPersistsPartialProviderResultsOnFailedApply` retains ordered 2xx/5xx receipt facts. Engine retry/cancellation and orchestrator commit/cancellation regressions remain covered by `TestDeclarativeWriteRetryPolicy`, `TestWriteCtxCancelMidLoopAccounting`, `TestTransportFamilyHalfPathConformanceAcknowledgementAndCancellation`, and `TestOrchestratorCommitsAcknowledgedPageBeforeReturningCancellation`.
- [x] Tombstone create/update/delete separation: `go test -count=1 -timeout 20m ./internal/app -run '^(TestDeclarativeTypedDestinationTombstonesRequireDeclaredDeleteAction|TestDeclarativeTypedDestinationTombstoneAppliesOnlyDeclaredDeleteAndReadsBackAbsence)$'` — passed. It uses a persisted project plan/preview/approval and a synthetic provider to prove only the declared delete action receives a durable tombstone key and provider read-back confirms absence.
- [x] App and CLI persisted boundaries: `TestPersistedConnectionSelectsDeclarativeTypedDestinationAction` and `go test -count=1 -timeout 20m ./internal/cli -run '^(TestETLTransportBareAndLeafHelpAreContextual|TestDeclarativeTypedDestinationTransportRejectsCallerActionBeforeProjectIO)$'` — passed. The route remains stored-connection selected; no caller action override is accepted.
- [x] Focused schema/engine/App/CLI/orchestration tests: `go test -count=1 -timeout 20m ./internal/connectors ./internal/synctransport` (passed), `go test -count=1 -timeout 20m ./internal/connectors/engine` (passed), and the focused App command above (passed). Race commands for changed core packages passed: `go test -race -count=1 -timeout 20m ./internal/connectors ./internal/synctransport` and the focused App race groups.
- [x] Generator checks: `go run ./cmd/connectorgen validate` — `552 connector(s) checked, 0 findings`; `go run ./cmd/connectorgen surface-sync --check` — `552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`.
- [x] One post-green tracked-artifact regeneration: `./pm docs generate --dir docs/cli` updated `docs/connectors/catalog/all-connectors.json`; no production connector definition changed. The command initially refused without its required `--dir`; `pm help docs` supplied the exact non-secret invocation.
- [x] Final App regression and package verification: the full `go test -count=1 -timeout 20m ./internal/app` suite passed in `255.389s` after the defaulted-action admission fix. The focused regression set covering three action-owned mappings, cross-connector selection, tombstones, batch limits, partial provider output, persisted selection, default selected action, and malformed worksets also passed.
- [x] Changed-package race verification: `go test -race -count=1 -timeout 20m ./internal/synctransport` passed (`2.181s`), and the focused App action-binding/tombstone/persisted-path group passed (`48.143s`).
- [x] Final CLI verification: `go test -count=1 -timeout 20m ./internal/cli -run '^(TestETLTransportBareAndLeafHelpAreContextual|TestDeclarativeTypedDestinationTransportRejectsCallerActionBeforeProjectIO)$'` passed (`6.173s`).
- [x] Static and repository gates: `go vet ./...`, `go build ./cmd/pm`, `make lint` (zero issues), `make connector-canon-check`, and `git diff --check` passed. The prior post-green chained run also passed `docs-check`, `smoke-no-build`, `agent-contract-check`, generator validation and `surface-sync --check`, `connector-boundary` (552 connectors; clean), and `release-workflow-check`.
- [x] Manual-GSD code review: inline standard-depth review found no remaining actionable issue; the only full-package finding was the defaulted-action admission ordering above, which was red-tested, fixed, and revalidated before this evidence was recorded.

## 2026-08-23 merge integration (in progress)

- [x] Non-rewriting merge preparation: `scripts/gsd doctor`, all required
  source/prompt resolution, and `go run ./cmd/agentcontractgen check` passed.
- [x] Focused merged-tree tests: `go test -count=1 -timeout 20m ./internal/app`
  (270.191s), `./internal/connectors/engine` (13.250s),
  `./internal/coordination` (4.333s), `./internal/connectors ./internal/synctransport`
  (1.096s / 1.169s), and `./internal/cli` (459.905s) passed.
- [x] `cd website && pnpm run gen:website-data` passed and produced no
  working-tree delta, proving generated website data is current for the merged tree.
- [ ] Full `make verify`, including standalone/polled `connector-boundary`, and
  the post-push PR workflow rollup remain required before delivery.
