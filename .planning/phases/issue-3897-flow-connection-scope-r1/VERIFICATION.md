# #3897 Verification Checklist

**Status:** Local GREEN verification complete with correction 1 / 5; pending
no-mistakes and external delivery gates.

## Required acceptance evidence

- [x] RED: two different connection-owned Parquet materializations with the
  same table name fail to resolve through explicitly scoped flow query/action
  reads before implementation.
- [x] GREEN: two different connection-owned materializations with the same
  table name are selected correctly through normal ETL-equivalent warehouse
  ownership and Parquet materialization.
- [x] Explicit query selector returns only the selected owner’s rows through
  Parquet/DuckDB.
- [x] Explicit action source selector returns only the selected owner’s rows
  through `ActionSourceReadRequest.Connection` and reaches the action stub
  only.
- [x] Omitted selector yields `*warehouse.AmbiguousTableError`; its remedy
  names no nonexistent CLI flag.
- [x] `_unattributed` reads only root-owned tables.
- [x] Selectors survive manifest parse/serialize and the action boundary.
- [x] No action/provider mutation occurs during tests.
- [x] Fresh binary local proof validates returned row IDs, not exit status.
- [x] Temporary project roots are removed and verified absent.
- [x] Correction 1 RED: a connection-selected action source with 101 rows
  reaches only 100 rows through the public `QueryTable(..., Limit: 0)` default.
- [x] Correction 1 preserves the public `QueryTable(..., Limit: 0)` 100-row
  default in the same real-Parquet fixture.
- [x] Correction 1 GREEN: the action-only uncapped read delivers all 101
  selected rows, retains owner isolation, and records no success checkpoint
  when the local runner returns an error.

## GSD/manual lifecycle record

- [x] `discuss-phase` prompt resolved; issue decisions captured inline.
- [x] `plan-phase --tdd` prompt resolved; plan and ledger created.
- [x] `execute-phase` RED/GREEN/REFACTOR complete.
- [x] `verify-work` complete; no gaps found.
- [x] `code-review` complete; no in-scope findings.
- [x] Shepherd-compatible evaluation recorded in `SHEPHERD-EVALUATION.md`.
- [ ] no-mistakes child gate complete without `--yes`.

## Verification commands

Record each result and SHA in `RUN-STATE.json` after execution.

## Local commands that passed

- Focused and full changed-package flow/app/CLI tests, including race coverage.
- `go vet ./...`, `go build ./cmd/pm`, and changed-diff Go lint.
- `make tidy-check`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`,
  `make release-workflow-check`, `make lint`,
  `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`,
  `make connector-canon-check`, and `make connector-runtime-preflight`.
- Website `npm run gen:docs` and `npm run typecheck`; the lint task reported
  only pre-existing warnings outside this change.
- Fresh-binary proof recorded in `traces/binary-flow-proof.txt`; its original
  temporary root was moved to the platform Trash and its original path tested
  absent. The smoke temporary root received the same cleanup.

## Correction 1 command record

- RED (exit 1): `go test -v -timeout 20m ./internal/cli -run
  '^TestFlowActionSourceReadsAllSelectedConnectionRows$' -count=1`; both action
  dispatch subtests reported 100 records where 101 selected `acme` rows were
  required.
- GREEN (exit 0): `go test -v -timeout 20m ./internal/app ./internal/flow
  ./internal/cli -run
  '^(TestEnginePassesManifestSourceConnectionSelectors|TestFlowSourceConnectionSelectorsReadOnlyOwningRows|TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed|TestFlowActionSourceReadsAllSelectedConnectionRows)$'
  -count=1` passed.
