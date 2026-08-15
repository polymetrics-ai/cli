# Verification — #4093

## Checklist

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
- [ ] Branch is rebased on `origin/integration/4015-mvp-flat-r1` immediately
  before the final push; PR base is API-verified.

## Acceptance matrix

| Acceptance criterion | Evidence | Observable assertion |
| --- | --- | --- |
| A definition-owned transport loads and projects safely | `TestBundleLoadSyncTransportProjectsIndependentDefinition` | Exact source/destination references appear in `Definition`; mutating one returned projection leaves the next unchanged. |
| Unknown or malformed declarations do not register anything | `TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations`, `TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration` | Unknown/version/role declarations are errors; source/destination builds, registry entries, reads, plans, and applies all remain zero. |
| Composition accepts declarations, not connector names | `TestRegisterDeclaredTransportsRegistersTwoDefinitionOwnedPairs`, `TestOpenRegistersDefinitionOwnedProductionTransports` | A synthetic second definition produces a second registered pair; real PostgreSQL and GitHub preflights resolve the declared executor IDs. |
| Evidence is externally admitted | `TestDefinitionConformanceVerifierRefusesAlteredEvidenceBeforeSourceIO` | Altered evidence fails preflight with zero source reads. |
| Destination change capture stays closed | `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination`, descriptor and composition tests | A `change_apply` declaration produces one source read/stage/plan/apply; a `change_capture`/`append` declaration has zero builds, registrations, reads, plans, and applies. |
| PostgreSQL remains live-proven | mandated tagged native PostgreSQL test | Docker/Colima integration suite passed for the definition-owned production source adapter before the CI repair; the post-repair mandatory run is recorded below. |
| Local warehouse is a closed destination | production preflight + local executor test | PostgreSQL-to-warehouse preflight resolves the warehouse-owned adapter; apply writes the connection-owned Parquet table and read-back confirms its durable digest and row count. |
| Legacy routes stay outside a half-transport | full app suite | One-sided declarations take the legacy path; malformed two-sided declarations reach preflight; the full app suite remains green. |
| GitHub inspection reports the declaration | CLI regression test + docs parity | `connectors inspect github --json` reports both statuses as `declared`; runtime help, CLI docs, generated connector docs, and website guidance describe that fact without claiming certification. |
