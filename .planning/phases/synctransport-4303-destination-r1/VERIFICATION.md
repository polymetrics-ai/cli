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

## Affected-package and binary checks

- `go test -count=1 -timeout 20m ./internal/app` — passed (249.108s).
- `go test -count=1 -timeout 20m ./internal/connectors` — passed.
- `go test -count=1 -timeout 20m ./internal/connectors/engine` — passed.
- `go test -count=1 -timeout 20m ./internal/synctransport` — passed.
- `go test -count=1 -timeout 20m ./internal/cli` — passed (529.662s).
- `go vet ./...` — passed.
- `go build ./cmd/pm` — passed.

## Application dispatch and result-output checks

- `go test -count=1 -timeout 20m ./internal/app -run '^TestPersistedConnectionSelectsDeclarativeTypedDestinationAction$'` — passed. The production-shaped synthetic bundles prove two named actions in one connector and one in another are selected only by the persisted `destination_action`; missing, foreign, unlisted, and stale selections refuse before I/O. The same test proves persisted `Run.destination_results` preserves ordinary response status, headers, nested fields, large numeric values, and tier-specific fields, while credential material remains present as an explicit mask marker.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields$'` — passed. Exact schema-valid snake_case and camelCase names are admitted only for the selected action; malformed, unknown, cross-action, runtime-selected, generic/shell/http, and undeclared names refuse before I/O.
- `go test -count=1 -timeout 20m ./internal/connectors -run 'TestSanitize.*Output'` — passed. Complete output retains ordinary provider results and masks only credential-bearing values in place.
- `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteHonorsDeclaredJSONAndNoneResponsePolicies$'` — passed. `output_policy: none` controls parsing only and cannot suppress successful ordinary provider output.
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
