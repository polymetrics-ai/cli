# TDD ledger

| Stage | Evidence | Expected result |
| --- | --- | --- |
| Red | Add the reconciliation test before the matrix exists. | `go test ./internal/connectors/defs/notion -run TestNotionSourceLaneMatrixRetainsEveryLockedOperationAndLane -count=1` failed only because `sources/notion-source-lane-matrix.json` was absent. |
| Green | Add the exact source artifacts and explicit matrix. | The focused reconciliation test passes. |
| Edge | Mutate an in-memory matrix copy in subtests. | Missing source row/cell/backlink, invalid lane disposition, boundary drop, and mapping-control restriction drop each fail. |
| Refactor | Normalize source facts and summary only after the assertions pass. | No change to source membership or lane dispositions. |

No live provider, credential, runtime, or certification test is in scope.

## Check-only mapping restrictions

`connectorgen validate internal/connectors/defs/notion --json` rejects the retained v2 `source_operation` payload, and `connectorgen surface-sync internal/connectors/defs --check` reports the absent canonical descriptor. Both results are intentionally retained in the matrix as mapping restrictions for all 49 source rows. They are not runtime-foundation gaps and this Track A change does not modify the shared importer or projection validator.
