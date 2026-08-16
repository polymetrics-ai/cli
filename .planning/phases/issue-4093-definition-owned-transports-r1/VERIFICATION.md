# Verification — #4093

## Checklist

### R2 continuation

- [x] Reconcile each #4093 clause with merged #4156/#4186 and the current
  integration branch; record the residual synthetic-bundle proof plus the
  history/CDC current-state divergence in the PR body.
- [x] Record a RED and GREEN run for the loaded synthetic connector route and
  the residual transient-stage lifecycle implementation.
- [x] Run the requested connector-boundary, generated-definition, scoped test,
  vet/build, website generation, and repository gates; record exact commands
  and results below.

### R3 certification regression repair

- [x] Preserve the certification proof's connection-owned durable receipt
  across ordinary `Open`, without touching the owned certification stage.
- [x] Reconcile committed receipts through the generic pre-execution path and
  prove it occurs before source I/O.
- [x] Run the failing certify-timing test plus the changed consumer/package and
  boundary/generated gates; record the exact results below.

- [x] Targeted `synctransport` and connector test packages pass:
  `go test -timeout 20m ./internal/synctransport/... ./internal/connectors/...`.
- [x] Race coverage passes for the changed registries/loaders and app
  composition: `go test -race -timeout 20m ./internal/synctransport
  ./internal/connectors/engine ./internal/connectors ./internal/app`.
- [x] Docker-backed PostgreSQL native integration test was rerun using the
  explicit Colima Unix endpoint from the task brief:
  `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock
  go test -tags=databaseintegration -count=1 -timeout 20m -v
  ./internal/connectors/native/postgres`. The transport source lanes passed,
  while the unrelated existing managed-target control assertion failed with
  `database history route source and destination do not match the declared
  managed-target driver` (tracked as #4158; the failing file is not changed by
  this branch). No unrelated PostgreSQL repair was made.
- [x] `go vet ./internal/synctransport/... ./internal/connectors/... ./internal/app`
  and `go build ./cmd/pm` pass.
- [x] `go test -count=1 -timeout 20m ./internal/app` passes after the generic
  two-sided routing repair; the focused local warehouse tests also pass under
  `-race`.
- [x] Required non-full-suite repository gates pass individually: `make
  tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make
  agent-contract-check`, `make connectorgen-validate`, `make
  connectorgen-surface-sync`, `make connector-boundary`, and `make
  release-workflow-check`.
- [x] `connectorgen validate` and `surface-sync --check` pass through the
  corresponding individual repository gates.
- [x] The Connector Boundary CI feedback was repaired with connector-local
  provider discovery; `go run ./cmd/connectorgen boundary . --json` passes on
  the repaired commit.
- [x] Full CI RED exposed omitted embedded transport declarations. The repair
  adds `*/sync_transport.json` to `defs.FS`; `go test -count=1 -timeout 20m
  -run TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle
  ./internal/cli`, the PostgreSQL declaration/registration tests, the full app
  and required transport/connector package sweeps, `connectorgen validate`,
  `surface-sync --check`, `connector-boundary`, vet, and CLI build pass after
  the repair.
- [x] CLI inspection parity: `TestConnectorInspectProjectsDeclaredSyncTransport`
  observes the production GitHub source/destination statuses as `declared`;
  `pm help connectors`, bare `pm connectors --json`, and `pm connectors --help`
  pass. `docs/cli/connectors.md`, generated GitHub connector docs/catalog, and
  `website/content/docs/agent-guide.mdx` document the metadata-only status;
  `make docs-check` passes.
- [x] Website generated data: `cd website && pnpm run gen:website-data` updates
  the agent-guide projection in `website/lib/docs.generated.ts` with no other
  tracked generated-data drift.
- [x] `./pm docs generate --dir docs/cli` regenerated the warehouse manual and
  skill after its declared destination status changed; `make docs-check` then
  passed. The golden transcript was regenerated after rebasing #4071 and its
  checking test passes.
- [x] Complete generated-artifact closure before the final push: ran `go run
  ./cmd/agentcontractgen sync`, `go run ./cmd/connectorgen gen`, `go run
  ./cmd/connectorgen surface-sync`, `go run ./cmd/connectorgen
  certification-matrix --all`, `./pm docs generate --dir docs/cli`, `./pm
  skills generate --dir docs/skills`, golden-transcript regeneration, and
  `cd website && npm run gen:website-data`. The matching local drift checks
  pass: agent-contract check, connector validation/surface/certification,
  connector docs validation, `TestSkillsGenerateMatchesTrackedSkills`,
  `TestGoldenTranscripts`, `TestGoldenDocsGenerateMatchesTrackedCLIManuals`,
  `TestDocsGenerateIncludesConnectorCatalog`, GitHub parity artifacts, and a
  generated-path diff. Only the expected generated GitHub and warehouse skill
  files changed.
- [x] Inline `verify-work` maps every acceptance criterion to an observable
  passing test or live output below.
- [x] Inline code review has no unresolved actionable findings; see REVIEW.md.
- [x] Branch was refreshed against `origin/integration/4015-mvp-flat-r1`
  immediately before the final push; both local HEAD and remote base resolved
  to `ff6a8710199c10f209d9d47cce87e5c8f7c429e6`. PR #4196 was created against
  `integration/4015-mvp-flat-r1`; the read-only GitHub API reports that exact
  base ref.

## Acceptance matrix

| Acceptance criterion | Evidence | Observable assertion |
| --- | --- | --- |
| A definition-owned transport loads and projects safely | `TestBundleLoadSyncTransportProjectsIndependentDefinition` | Exact source/destination references appear in `Definition`; mutating one returned projection leaves the next unchanged. |
| Unknown or malformed declarations do not register anything | `TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations`, `TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration` | Unknown/version/role declarations are errors; source/destination builds, registry entries, reads, plans, and applies all remain zero. |
| Composition accepts declarations, not connector names | `TestRegisterDeclaredTransportsRegistersTwoDefinitionOwnedPairs`, `TestOpenRegistersDefinitionOwnedProductionTransports` | A synthetic second definition produces a second registered pair; real PostgreSQL and GitHub preflights resolve the declared executor IDs. |
| A loaded synthetic second connector routes generically | `TestAppCompositionRoutesLoadedSyntheticDefinitionConnector` | A real test bundle supplies `sync_transport.json`; generic App composition selects its named hooks by declared family/ID/evidence and the generic orchestrator reads, stages, plans, applies, read-backs, and commits one record. |
| Evidence is externally admitted | `TestDefinitionConformanceVerifierRefusesAlteredEvidenceBeforeSourceIO` | Altered evidence fails preflight with zero source reads. |
| Destination change capture stays closed | `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination`, descriptor and composition tests | A `change_apply` declaration produces one source read/stage/plan/apply; a `change_capture`/`append` declaration has zero builds, registrations, reads, plans, and applies. |
| Committed transient stages reconcile safely | `TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit`, `TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages`, `TestRunETLTransportReconcilesCommittedStagesBeforeSourceIO` | An opt-in stage has zero retirements on persistence failure and one on success. Connection-owned evidence remains through ordinary Open; before the next source read, generic reconciliation removes only a matching committed manifest/WAL/Parquet and leaves an active receipt intact. A reconciliation error has zero source reads and destination applies. |
| PostgreSQL remains live-proven | mandated tagged native PostgreSQL test | Docker/Colima integration suite passed for the definition-owned production source adapter before the CI repair; the post-repair mandatory run is recorded below. |
| Local warehouse is a closed destination | production preflight + local executor test | PostgreSQL-to-warehouse preflight resolves the warehouse-owned adapter; apply writes the connection-owned Parquet table and read-back confirms its durable digest and row count. |
| Legacy routes stay outside a half-transport | full app suite | One-sided declarations take the legacy path; malformed two-sided declarations reach preflight; the full app suite remains green. |
| GitHub inspection reports the declaration | CLI regression test + docs parity | `connectors inspect github --json` reports both statuses as `declared`; runtime help, CLI docs, generated connector docs, and website guidance describe that fact without claiming certification. |

## R2 commands and results

All commands below exited 0 unless stated otherwise.

```text
# RED before the lifecycle production patch
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit$'
# failed: checkpoint committed reported "retired receipts = 0, want 1"

# R2 focused GREEN
go test -count=1 -timeout 20m ./internal/app -run '^TestAppCompositionRoutesLoadedSyntheticDefinitionConnector$'
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit$'
go test -count=1 -timeout 20m ./internal/app -run '^TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages$'
go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations$'
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestPreflightReturnsTypedDestinationSourceIneligibleErrorBeforeExecutorAccess$'
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration$'
go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination$'

# Scoped suites and build quality
go test -count=1 -timeout 20m ./internal/synctransport
go test -count=1 -timeout 20m ./internal/connectors/engine
go test -count=1 -timeout 20m ./internal/connectors
go test -count=1 -timeout 20m ./internal/app
go test -count=1 -timeout 20m ./internal/cli
go test -count=1 -timeout 20m ./cmd/connectorgen
go vet ./...
go build ./cmd/pm
go run ./cmd/connectorgen boundary . --json
pnpm --dir website run gen:docs  # run twice; both generated 12 pages and left no diff

# Individual make verify gates
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
make connectorgen-certification-matrix
make connector-canon-check
make github-parity-artifacts-check
git diff --check
```

The repository instruction explicitly advises against one monolithic
`go test ./...`/`make verify` invocation in this per-command environment; the
changed packages plus `internal/cli` and every non-test `make verify` gate were
therefore run separately above. Full-suite execution remains CI coverage.

## R3 certification regression commands and results

All exited 0 after the recorded CI RED:

```text
go test -count=1 -timeout 20m ./internal/connectors/certify -run '^TestCertificationDeclaredTransportPairResolvesAndExecutes$'
go test -count=1 -timeout 20m ./internal/app -run '^(TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages|TestRunETLTransportReconcilesCommittedStagesBeforeSourceIO)$'
make certify-timing
go test -count=1 -timeout 20m ./internal/connectors/certify
go test -count=1 -timeout 20m ./internal/app
go vet ./internal/app ./internal/connectors/certify
go run ./cmd/connectorgen boundary . --json
git diff --check
```
