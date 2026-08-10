# #3897 Verification Checklist

**Status:** Local GREEN verification complete; pending no-mistakes and
external delivery gates

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
  through `QueryTableRequest.Connection` and reaches the action stub only.
- [x] Omitted selector yields `*warehouse.AmbiguousTableError`; its remedy
  names no nonexistent CLI flag.
- [x] `_unattributed` reads only root-owned tables.
- [x] Selectors survive manifest parse/serialize and the action boundary.
- [x] No action/provider mutation occurs during tests.
- [x] Fresh binary local proof validates returned row IDs, not exit status.
- [x] Temporary project roots are removed and verified absent.

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
