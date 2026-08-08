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

- Status: pending the captain-policy implementation and rerun.
- Test: run the same real-binary sweep against the generated PersistIQ bundle,
  plus validation, surface-sync check, runtime preflight, and batch gate.
- Captured materializer output (verbatim):

  ```text
  connectorgen batch materialize: 0 connector(s) materialized, 1 dropped; report .planning/phases/persistiq-artifact-materialize-pilot-r1/materialize-report.json
  exit status 1
  ```

  Captured drop (verbatim):

  ```text
  executable coverage GET /v1/mailboxes is absent from the cited artifact
  ```

  This is historical Red evidence for the superseded policy. The generated
  bundle and its commands did not exist in that run; no Green claim is made
  from it and no generated command is counted reachable.

## Refactor / safety

- No new generator, generic write tool, credentials, or gate weakening.
- Any failed batch stage remains a failed pilot result; it is not papered over
  by editing the test or source artifact.
