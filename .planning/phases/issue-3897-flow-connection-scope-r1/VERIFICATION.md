# #3897 Verification Checklist

**Status:** RED established; GREEN pending

## Required acceptance evidence

- [x] RED: two different connection-owned Parquet materializations with the
  same table name fail to resolve through explicitly scoped flow query/action
  reads before implementation.
- [ ] GREEN: two different connection-owned materializations with the same
  table name are selected correctly through normal ETL-equivalent warehouse
  ownership and Parquet materialization.
- [ ] Explicit query selector returns only the selected owner’s rows through
  Parquet/DuckDB.
- [ ] Explicit action source selector returns only the selected owner’s rows
  through `QueryTableRequest.Connection` and reaches the action stub only.
- [ ] Omitted selector yields `*warehouse.AmbiguousTableError`; its remedy
  names no nonexistent CLI flag.
- [ ] `_unattributed` reads only root-owned tables.
- [ ] Selectors survive manifest parse/serialize and the action boundary.
- [ ] No action/provider mutation occurs during tests.
- [ ] Fresh binary local proof validates returned row IDs, not exit status.
- [ ] Temporary project roots are removed and verified absent.

## GSD/manual lifecycle record

- [x] `discuss-phase` prompt resolved; issue decisions captured inline.
- [x] `plan-phase --tdd` prompt resolved; plan and ledger created.
- [ ] `execute-phase` RED/GREEN/REFACTOR complete.
- [ ] `verify-work` complete; gaps planned/executed if needed.
- [ ] `code-review` complete; all findings dispositioned.
- [ ] Shepherd-compatible evaluation recorded.
- [ ] no-mistakes child gate complete without `--yes`.

## Verification commands

Record each result and SHA in `RUN-STATE.json` after execution.
