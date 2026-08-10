# #3897 TDD Ledger

**Status:** RED established; no production edit has begun.  
**Correction rounds:** 0 / 5

| Slice | Red evidence | Green evidence | Refactor / result |
|---|---|---|---|
| 1. Selected source rows | **RED 2026-08-11:** `go test -timeout 20m ./internal/cli -run '^TestFlowSourceConnectionSelectorsReadOnlyOwningRows$' -count=1` failed for both explicit selectors. DuckDB reported bare `records` absent and suggested only a qualified owner view. | Pending: each selector returns only its owner’s IDs through DuckDB/warehouse. | Pending |
| 2. Omitted and root selectors | Pending: omitted duplicated source must be typed ambiguity; `_unattributed` must not read a connection table. | Pending: honest flow-manifest remedy and root-only success. | Pending |
| 3. Serialization/action boundary | Pending: parse/serialize and runner-boundary selector assertions fail. | Pending: selectors survive JSON and the action source request/step boundary. | Pending |
| 4. Public proof | Pending: binary flow fixture has no scoping syntax/returns wrong rows. | Pending: fresh `pm` reports only the selected owner’s returned IDs. | Pending |

## Red: required first executable test

Create `acme` and `globex` materializations through normal local ETL with the
same `records` table name. The test must drive the flow query and action source
read boundaries with an explicit declared selector and assert the actual rows
from the Parquet/DuckDB path. It is RED until the current selector drop is
removed. No provider action, approval token, or external network call is part
of the test.

### Recorded RED output

`TestFlowSourceConnectionSelectorsReadOnlyOwningRows` creates two connection
records plus two `warehouse.Location` ownership records and writes each
`records.parquet` table through `warehouse.WriteTable`. The query and action
source reads both failed at the real DuckDB boundary:

```text
flow: step failed: step <step>: execute query: Catalog Error:
Table with name records does not exist!
Did you mean "records__conn_<owner-id>"?
```

This is the target defect: the declared `connection: "acme"` and JSON
`action_cfg.source_connection: "globex"` were not carried into the source
read. The full non-secret result is in `traces/red-flow-source-selectors.txt`.

## Green safety constraints

- `warehouse.FindTable` remains the owner-selection authority.
- `_unattributed` is forwarded unchanged and never becomes a real connection.
- SQL receives no user-built identifier interpolation.
- Omitted selection never falls back to an arbitrary owner.
- An action remains a local stub boundary; #3994 owns dispatch.
