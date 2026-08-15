# Verification — #4093

## Checklist

- [x] Targeted `synctransport` and connector test packages pass:
  `go test -timeout 20m ./internal/synctransport/... ./internal/connectors/...`.
- [x] Race coverage passes for the changed registries/loaders and app
  composition: `go test -race -timeout 20m ./internal/synctransport
  ./internal/connectors/engine ./internal/connectors ./internal/app`.
- [x] Docker-backed PostgreSQL native integration test passes using the explicit
  Colima Unix endpoint from the task brief:
  `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock
  go test -tags=databaseintegration -count=1 -timeout 20m -v
  ./internal/connectors/native/postgres`.
- [x] `go vet ./internal/synctransport/... ./internal/connectors/... ./internal/app`
  and `go build ./cmd/pm` pass.
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
| Destinations cannot claim change capture | `TestDestinationTransportDescriptorRefusesChangeCapture`, composition refusal test | The role is rejected before construction, registration, or executor I/O. |
| PostgreSQL remains live-proven | mandated tagged native PostgreSQL test | Docker/Colima integration suite passes using the definition-owned production source adapter. |
