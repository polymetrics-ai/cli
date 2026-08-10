# #3897 TDD Ledger

**Status:** Planned; no production edit has begun.  
**Correction rounds:** 0 / 5

| Slice | Red evidence | Green evidence | Refactor / result |
|---|---|---|---|
| 1. Selected source rows | Pending: two connection-owned Parquet tables named `records`; explicit flow query/action selector cannot choose one. | Pending: each selector returns only its owner’s IDs through DuckDB/warehouse. | Pending |
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

## Green safety constraints

- `warehouse.FindTable` remains the owner-selection authority.
- `_unattributed` is forwarded unchanged and never becomes a real connection.
- SQL receives no user-built identifier interpolation.
- Omitted selection never falls back to an arbitrary owner.
- An action remains a local stub boundary; #3994 owns dispatch.
