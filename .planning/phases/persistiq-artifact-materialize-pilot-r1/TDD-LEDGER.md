# PersistIQ artifact materialization pilot - TDD ledger

> This ledger is maintained before and during the pilot. It must contain
> observed Red and Green evidence, not plan-shaped claims.

## Scope lock

- One connector: `persistiq`.
- No credentials, provider data, or live API exercise.
- 21 ledger operations; pre-fetch mapping: ETL 11, direct_read 1,
  reverse_etl 7, direct_write 2, binary_download 0, unclassified 0.

## RED

- Status: confirmed before production bundle edits.
- Test: build/run the real `pm` binary against the baseline PersistIQ bundle;
  assert the pilot command surface is not fully reachable.
- Command: `.planning/phases/persistiq-artifact-materialize-pilot-r1/pm-baseline persistiq leads list --help`
- Captured output (verbatim):

  ```text
  error: unknown command "persistiq"
  ```

  This is the baseline red: the current binary has no reachable PersistIQ
  command namespace before materialization.

## Captain-ruling policy change

- The old materializer's Red was a fail-closed refusal when the existing
  `/v1/mailboxes` stream was absent from the fetched artifact.
- New required Red tests: artifact operations are retained when unsupported,
  each `not_implemented` command has a machine-checkable named dependency, and
  source-surface-only operations remain with the exact discrepancy marker.
- Missing executor/foundation is a visible gap, not a drop. The implemented
  availability invariant remains enforced by runtime preflight.

### Batch gate placement

- Materialization is intentionally not a gate: it does not invoke the
  repository-wide runtime-preflight sweep per connector. The report's
  `runtime_preflight_commands` remains zero until `batch gate` runs over the
  staged result.
- This preserves review-sized batch boundaries without paying the full sweep
  once for every candidate.

### Captain-ruling Red run

- Command: `go test -timeout 20m ./cmd/connectorgen -run
  'TestBatchMaterializeMapsUnsupportedOperationsAndFlagsSurfaceDiscrepancies|TestValidate_CLISurfaceNotImplementedRequiresNamedDependency'`
- Observed Red: the materializer returned `0 connector(s) materialized, 1
  dropped` for the source-only endpoint, and the new availability value was
  rejected by the CLI schema enum. These are the two intended pre-change
  failures: complete inventory was still refused and named dependency
  validation had not yet been wired.

## GREEN

- Status: confirmed under the captain complete-inventory policy.
- Tests: `go test -timeout 20m ./cmd/connectorgen`,
  `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner`,
  and the real PersistIQ generated-bundle sweep.
- Captured materializer output (verbatim):

  ```text
  connectorgen batch materialize: 1 connector(s) materialized, 0 dropped; report .planning/phases/persistiq-artifact-materialize-pilot-r1/rerun-2026-08-08/materialize-report.json
  ```

- Captured static/runtime evidence: `connectorgen validate` returned zero
  findings; `surface-sync --check` reported no drift; `batch gate` included
  one candidate with 21 runtime-preflight checks; and the real binary reached
  all 24 generated help paths (21 implemented plus 3 intentionally blocked).
- The three blocked commands were each rejected with
  `availability=not_implemented` before credentials/network, with no unknown
  command result.
- Captured materializer output (verbatim):

  ```text
  connectorgen batch materialize: 0 connector(s) materialized, 1 dropped; report .planning/phases/persistiq-artifact-materialize-pilot-r1/materialize-report.json
  exit status 1
  ```

  Captured drop (verbatim):

  ```text
  executable coverage GET /v1/mailboxes is absent from the cited artifact
  ```

  This is historical Red evidence for the superseded policy. It is retained
  only to show why the captain ruling changed the materializer.

## REFACTOR / safety

- No new generator, generic write tool, credentials, or gate weakening.
- Materialization now maps unsupported operations into typed metadata and
  named-dependency commands, while the final gate remains authoritative for
  implemented claims.
- The generated PersistIQ bundle was installed only into a temporary embedded
  binary test and the original source bundle was restored exactly.
