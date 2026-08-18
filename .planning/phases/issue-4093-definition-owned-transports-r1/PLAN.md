---
phase: issue-4093-definition-owned-transports-r1
plan: 01
type: tdd
status: completed
---

# Plan — #4093 definition-owned production transports

## R2 continuation — task delivery header

- **Issue:** `Refs #4093 — feat(synctransport): load definition-owned transports, register production adapters, and retire transient stages`
- **Base branch:** `integration/4015-mvp-flat-r1`
- **Merges into:** `integration/4015-mvp-flat-r1 → main`
- **Delivery:** direct PR open against `integration/4015-mvp-flat-r1`, API-verified after creation.
- **Working branch:** `fm/cli-4093-definition-owned-transports-r1`
- **Task:** Reconcile the current branch with #4093 and close the residual proof gap: a second connector must be loaded from its own `sync_transport.json`, discovered through the generic factory-provider path, registered by App without connector-name dispatch, and run through the generic orchestrator. Close the remaining transient-stage lifecycle gap by retiring only exact receipts after checkpoint persistence and reconciling a process killed between those two durable events.
- **Verification:** focused RED/GREEN tests in `internal/app` and `internal/synctransport`, connector boundary, generated-definition checks, vet/build, scoped suites, and repository gates recorded in `VERIFICATION.md`.

### Reconciliation before R2 implementation

- Merged **#4156** (`11fd27b4b`) already supplies versioned strict loading,
  clone-safe `engine.Definition` projection, exact family/identifier factory
  selection, definition-selected evidence admission, generic registration, and
  production PostgreSQL/GitHub hooks.
- Merged **#4186** (`281560ca1`) adds GitHub's definition-owned
  `destination_transport.source_bindings`; its source eligibility no longer
  belongs in shared Go.
- Merged **#4187** and **#4188** made the advertised history modes executable.
  The original issue's route-specific history refusal is therefore stale and
  must not be reintroduced. CDC/change-capture remains role-aware: a
  destination cannot independently capture changes; the existing closed
  `change_apply` strategy is the destination half of a source `change_capture`
  run.
- Existing `TestRegisterDeclaredTransportsRegistersTwoDefinitionOwnedPairs`
  proves generic registration with Go-authored test connectors, but it does
  not load a throwaway bundle, invoke `App.composeTransportRegistry`, or run a
  production composition result through the generic orchestrator. That is the
  residual this continuation owns.

### R2 TDD slice

1. **Residual proof:** add a test-only connector whose sole transport
   declaration is `sync_transport.json`; inject it into an App registry solely
   through the ordinary factory-provider interface. Baseline reconciliation
   found the existing generic composition already satisfies this behavior, so
   this is a GREEN acceptance proof rather than a contrived production-code
   rewrite. It observes one build per role, one staged source record, one
   destination apply/read-back, and one acknowledged checkpoint; it fails if a
   connector-name filter or dispatch bypass is introduced.
2. **RED:** make an accepted destination effect whose checkpoint persistence
   fails. The stage must not be retired. Before the lifecycle implementation,
   the successful sibling case also reported zero retirements.
3. **GREEN:** make the optional owned-stage retirement happen only after the
   durable commit, then prove a fresh App safely and boundedly reconciles an
   exact committed receipt after a simulated process kill while preserving an
   active receipt.

## R2 evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A synthetic connector is definition-owned and generic | fake — deterministic in-memory bundle | `engine.Load` parses the synthetic `sync_transport.json`; App discovers its named test hook via `DefinitionFactoryProvider`, and generic orchestration stages/applies/reads back one record and commits one checkpoint. An in-memory bundle is necessary because `defs.FS` is compile-time embedded; the assertion is against the real bundle loader, App composition, registry, and orchestrator, not a hand-authored descriptor. |
| A committed transient stage is retired without overreach | fake — deterministic owned project fixture | `TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit` proves that an opt-in generic stage receives no retirement before a persistence failure and one exact receipt after a successful commit. `TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages` proves that connection-owned evidence survives ordinary `Open`, then its bounded pre-execution reconciler removes only the committed manifest/WAL/Parquet and preserves the active receipt. |

## R3 certification regression repair

CI's `certify-timing` target found that eager post-commit retirement and eager
`Open` reconciliation delete the exact connection-owned durable receipt which
`app.ProbeDeclaredTransportForCertification` must reopen and inspect. The
declared GitHub pair still executed (one apply and a committed checkpoint), but
the certification proof correctly failed with zero retained manifest and
Parquet artifacts. This repair is confined to generic stage registration and
runtime lifecycle; `internal/connectors/certify/stages_transport.go` is owned
by the concurrent database-certification lane and will not be changed.

1. **RED:** `go test -count=1 -timeout 20m ./internal/connectors/certify -run
   '^TestCertificationDeclaredTransportPairResolvesAndExecutes$'` fails with
   `warehouse_manifests=0 warehouse_parquet=0` after a successful execution.
2. **GREEN:** keep the generic optional immediate-retirement interface for
   stages that opt in, but leave connection-owned receipt evidence available
   after a completed run and reopen. Run its bounded committed-receipt cleanup
   before the next generic transport execution, before source I/O. Prove the
   certification pair passes and a kill-after-commit receipt is cleaned only by
   that deferred reconciliation, never by ordinary `Open`.

## Scope

This foundation issue owns bundle loading/projection, strict role validation,
definition-neutral production registration, and the GitHub/PostgreSQL
declarations and adapters required to exercise them. It deliberately excludes
#4125, #4136, #4090 implementation changes, and #4154 apply-boundary work.

## Delivery record

Required-skill routing and the Go `how-to`, testing, error-handling, safety,
security, design-patterns, structs/interfaces, database, dependency-injection,
CLI, and documentation skills were reviewed. The GSD adapter was validated
with `scripts/gsd doctor`; its numeric-phase-only verify/review commands cannot
target this issue-named directory, so verification and code review are recorded
as the documented inline/manual fallback in `VERIFICATION.md` and `REVIEW.md`.

## TDD slices

1. **Planning checkpoint:** create CONTEXT, this plan, TDD ledger,
   verification checklist, and run state before production changes.
2. **RED loader/projection:** prove a versioned declaration is present in the
   loaded bundle and definition, survives a round trip, and cannot be mutated
   through one definition projection to affect another. Prove unknown JSON
   members, a missing/wrong schema version, and unsupported descriptor values
   are errors.
3. **RED atomic composition:** use stateful fake factories/registries to prove
   a declared role produces a real registration, while an unknown executor or
   malformed role produces an error with no factory build and no source or
   destination registration. Add a synthetic second connector to demonstrate
   that the composer requires no App/orchestrator/dispatch edit.
4. **GREEN loader and role rule:** add `sync_transport.json` schema, loader,
   bundle field, clone-safe `Definition` projection, and require a destination
   `change_capture` declaration to bind the closed `change_apply` strategy.
5. **GREEN production composition:** replace the hard-coded GitHub wrapper
   composition with a declarative factory table; collect connector-local
   factories through a generic provider interface; register actual GitHub
   source/destination adapters and the actual PostgreSQL snapshot source from
   their declared definition roles. The verifier must accept only the factory's
   external exact evidence references.
6. **Definition parity:** move GitHub and PostgreSQL descriptor data into their
   own `defs/*/sync_transport.json`, delete obsolete Go-authored descriptors,
   and retain the existing typed adapter contracts.
7. **Verification/review:** run focused/race tests, the required live PostgreSQL
   proof, build/vet and individual repository gates, then inline `verify-work`
   and `code-review`. Rebase immediately before push and open a direct PR.
8. **CI repair:** preserve legacy GitHub-to-local-warehouse execution by
   declaring the closed transport modes in the canonical contract order, and
   correct the destination role rule so a legitimate `change_capture` route
   reaches its descriptor-owned `change_apply` strategy.
9. **Warehouse destination repair:** give the local `warehouse` primitive its
   own closed destination declaration and a reference-bound production adapter.
   The adapter writes a connection-owned Parquet materialization from the
   reopened workset, durably re-reads that materialization, and only then
   returns its acknowledgement. No preflight exception or legacy dispatch
   fallback is allowed.
10. **Closed-pair routing repair:** require declarations on both endpoints
    before App selects transport dispatch. A destination declaration cannot
    divert a legacy source into a preflight that it never declared; a
    two-sided malformed pair still reaches preflight and refuses there.

## Acceptance evidence

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Definition declaration loads and projects clone-safely | fake bundle | A loaded `sync_transport.json` has exact source/destination refs; mutating one returned `Definition` does not change the next projection. |
| Unknown or malformed declaration fails closed | fake bundle/composer | Loader errors; composer asserts build calls and registry source/destination counts are both zero. |
| Registration is definition-owned, not connector-name dispatched | fake + production | A synthetic additional declared connector registers from the same composer without App/orchestrator/dispatch modification; real GitHub/PostgreSQL preflight reaches their registered executors. |
| Evidence admission is external and selected by declaration | fake | Exact factory evidence preflights; altered evidence fails preflight before source read. |
| Destination `change_capture` stays closed | fake bundle + app transport | A `change_apply` declaration reaches destination planning and apply; an `append` declaration fails before any registration or I/O. |
| Local warehouse is a real closed destination | production app + local filesystem | PostgreSQL-to-warehouse preflight resolves the warehouse-owned reference; apply writes a connection-owned Parquet table and read-back observes its digest and row count. |
| PostgreSQL production path remains live | live database | Docker PostgreSQL test emits bounded snapshot rows/pages/checkpoint through the registered native source. |

## Verification commands

```text
go test -timeout 20m ./internal/synctransport/... ./internal/connectors/...
go test -race -timeout 20m ./internal/synctransport ./internal/connectors/engine ./internal/connectors ./internal/app
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
  go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
go vet ./internal/synctransport/... ./internal/connectors/... ./internal/app
go build ./cmd/pm
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```
