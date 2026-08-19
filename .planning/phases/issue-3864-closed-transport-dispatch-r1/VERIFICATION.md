# #3864 verification checklist

## Status: correction-loop 5 focused evidence recorded; broader outer gates and child-local delivery gate pending

- [x] TDD RED outputs are recorded before production code.
- [x] Correction loops 3 through 5 retain the closed inspection, descriptor, checkpoint-identity,
  and durable-acknowledgement runtime boundaries.
- [ ] The outer executor must rerun the broader post-review package, CLI/help, docs, and
  delivery gates; this review step intentionally ran only T15–T21's focused regression commands.
- [x] The correction-loop 5 focused count-one `internal/app` regression command passed with
  `-timeout 20m`.
- [x] Transport package race test and cancellation regression pass.
- [x] `go vet` and build pass.
- [x] Required non-suite `make verify` components pass individually.
- [x] `make connector-runtime-preflight`, connector canon, connectorgen validation, and
  surface-sync checks pass.
- [x] `pm connectors`, `pm help connectors`, `pm connectors --help`, and
  `pm connectors inspect sample --json` are checked in an initialized project without
  credentials; docs/website parity is checked.
- [x] Manual `verify-work` outcome (zero automated gaps), code-review findings/dispositions, and
  supervisor-compatible local evidence are recorded in `UAT.md`, `REVIEW.md`, and
  `SUPERVISOR-EVIDENCE.md` using the documented manual-GSD fallback.
- [ ] Child-local `no-mistakes axi run --intent <complete issue intent> --skip=push,pr,ci`
  result (without `--yes`) is recorded. Push, sub-PR creation to
  `feat/3862-any-to-any-transport`, automated-review coverage, and CI belong to the outer
  delivery owner; this child must not create another parent/default PR or merge.

## Local evidence

- Correction loop 5 first recorded distinct RED failures for T20 and T21. Its sole focused GREEN
  command was `go test -count=1 -timeout 20m ./internal/app -run
  '^(TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume|TestRunETLTransportPersistsActiveCheckpointBeforeSourceFailureForAllModes|TestRunETLTransportAdvancesInterimCheckpointAcrossPages|TestRunETLTransportPreservesUnrelatedStateDuringInterimCheckpointCommit|TestRunETLTransportRejectsStaleCheckpointWriter|TestRunETLTransportDistinguishesMissingAndPresentStreamState|TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation|TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails|TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure)$'`.
  It passed, proving typed rebootstrap before state writes, per-stream compare-and-swap against
  the prior/last interim entry, raw opaque-byte comparison, metadata preservation, unrelated
  state preservation, stale-writer rejection, cancellation ordering, state-store outcomes, and
  all seven modes. Broader post-review gates remain outer-executor work.
- Correction loop 4 first recorded independent RED failures in T15–T19. Its sole focused GREEN
  command was `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m
  ./internal/app ./internal/connectors ./internal/synctransport ./internal/cli -run
  '^(TestRunETLTransportPersistsActiveCheckpointBeforeSourceFailureForAllModes|TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation|TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails|TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure|TestSyncTransportEligibilityProjectsDeclaredNoneAcknowledgement|TestSyncTransportEligibilityProjectsValidRolesIndependently|TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead|TestOrchestratorRejectsInvalidSiblingDescriptorBeforeProviderAccess|TestNewRegistryFailsClosedForTypedNilConformanceVerifier|TestCloneRecordCopiesBinaryValuesAtEveryNestingLevel|TestOrchestratorCommitsAcknowledgedPageBeforeReturningCancellation|TestConnectorsHelpExplainsDeclaredNoneInspectionPolicy|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestGoldenTranscripts|TestConnectorInspectProjectsUnsupportedSyncTransport)$'`.
  It passed after regenerating `docs/cli` and the transcript fixture, covering all seven mode
  checkpoint cases, both descriptor roles, typed nil, byte isolation, runtime help, and docs goldens.
- `go test -timeout 20m ./internal/connectors ./internal/synctransport`,
  `go test -timeout 20m ./internal/app`, and `go test -timeout 20m ./internal/cli` passed.
- Review correction #4029 first reproduced the declared-`none` inspection failure with
  `go test -v -timeout 20m ./internal/connectors -run
  '^TestSyncTransportEligibilityProjectsDeclaredNoneAcknowledgement$'`. The focused GREEN
  command `go test -v -timeout 20m ./internal/connectors ./internal/synctransport -run
  '^(TestSyncTransportEligibilityProjectsDeclaredNoneAcknowledgement|TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead)$'`
  passed, covering both projection and the unchanged runtime rejection.
- `go test -race -timeout 20m ./internal/synctransport -run
  '^(TestRegistryPreflightIsRaceSafeDuringRegistration|TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches|TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply)$'`
  passed.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`,
  `make docs-check`, `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, `make release-workflow-check`, and
  `make connector-runtime-preflight` passed.
- `golangci-lint run --new-from-rev=origin/feat/3862-any-to-any-transport
  ./internal/app/... ./internal/cli/... ./internal/connectors/... ./internal/synctransport/...`
  returned `0 issues`. A broader non-diff sweep reports pre-existing findings in unrelated
  app/connsdk files; they are not changed or waived by this child.
- A freshly built binary in an initialized temporary project rendered the three connector
  help forms with `sync_transport` wording and `connectors inspect sample --json` with
  both roles `unsupported`. `docs/cli/connectors.md`, its generated transcript fixture,
  and `website/content/docs/agent-guide.mdx` carry the same non-certification boundary.
- `scripts/gsd prompt verify-work issue-3864-closed-transport-dispatch-r1` and
  `scripts/gsd prompt code-review issue-3864-closed-transport-dispatch-r1` were executed.
  The Pi adapter cannot create the mandated worker roles for this non-numbered issue phase, so
  `UAT.md` and `REVIEW.md` record the inline/manual fallback and its results.

## Explicit limits

Correction loop 1/5 is tracked in [#4021](https://github.com/polymetrics-ai/cli/issues/4021):
an authored invalid descriptor must reach closed preflight rather than legacy routing. Its RED and
GREEN command evidence is in T11 of `TDD-LEDGER.md`.

Correction loop 2/5 is tracked independently in [#4023](https://github.com/polymetrics-ai/cli/issues/4023):
the closed descriptor must reject `generic-http` just as it rejects `generic_http`. Its RED and
GREEN command evidence is in T13 of `TDD-LEDGER.md`. Shared correction commit `9775f420c`
references #4021, #4023, #3864, and #3862 while retaining each issue's bounded scope.

Correction loop 3/5 is tracked in [#4029](https://github.com/polymetrics-ai/cli/issues/4029):
inspection must report a structurally valid destination `acknowledgement: none` as declared
without admitting it to runtime execution. T14 records the focused RED/GREEN evidence. It stays
in this child alongside #4021 and #4023; a topology restart is not a product correction loop.

Correction loop 4/5 remains in the same child: [#4046](https://github.com/polymetrics-ai/cli/issues/4046)
persists acknowledged active-stream checkpoints, [#4045](https://github.com/polymetrics-ai/cli/issues/4045)
full-validates selected descriptors, [#4048](https://github.com/polymetrics-ai/cli/issues/4048) rejects
typed-nil verifiers, [#4047](https://github.com/polymetrics-ai/cli/issues/4047) isolates binary record
values, and #4029 aligns canonical help and its artifacts. T15–T19 record separate RED/GREEN evidence;
this is not a PR split and a topology restart is not a product correction loop.

Correction loop 5/5 remains in the same child and under [#4046](https://github.com/polymetrics-ai/cli/issues/4046):
T20 rejects acknowledgement-stamped checkpoints that do not match the active source identity or
generation, while T21 prevents stale target-stream state replacement without blocking unrelated
project updates. Both RED/GREEN records remain in this one child; no extra PR or topology restart
was created.

This verification can prove only fake-backed dispatch and metadata surfaces. It cannot
truthfully assert executable #3810 conformance, a real API/database transport, a live
provider flow, automatic Shepherd certification, or a green GitHub CI/review state until
those gates actually run.
