# TDD ledger

| Stage | Evidence | Expected result |
| --- | --- | --- |
| Red | Add the reconciliation test before the matrix exists. | `go test ./internal/connectors/defs/notion -run TestNotionSourceLaneMatrixRetainsEveryLockedOperationAndLane -count=1` failed only because `sources/notion-source-lane-matrix.json` was absent. |
| Green | Add the exact source artifacts and explicit matrix. | The focused reconciliation test passes. |
| Edge | Mutate an in-memory matrix copy in subtests. | Missing source row/cell/backlink, invalid lane disposition, boundary drop, and mapping-control restriction drop each fail. |
| Refactor | Normalize source facts and summary only after the assertions pass. | No change to source membership or lane dispositions. |

## Independent-review repair: semantic POST direct reads

| Stage | Evidence | Expected result |
| --- | --- | --- |
| Red | Tighten the connector-local reconciliation test so all four retained semantic POST reads must be `direct_read` `source_candidate` / `mapped_unproven`, while the matrix still marks them not-applicable because of their method. | The focused test fails at `notion.rest.post-database-query` with `lane direct_read must be mapped_unproven`. |
| Green | Map `post-database-query`, `post-search`, `query-meeting-notes`, and `introspect-token` as semantic POST direct reads, with the immutable source-lock backlink and source-fact classification. | The focused reconciliation test passes with 24 mapped-unproven direct-read cells and 25 not-applicable cells. |
| Edge | Mutate an in-memory `post-page` direct-read cell, remove the retained `200 application/json` response from `post-search`, and assert the meeting-notes direct-read plus restricted-ETL dual classification. | Mutation POST promotion and missing source response facts fail; the meeting-notes ETL continuation restriction remains required. |

No live provider, credential, runtime, or certification test is in scope.

## Check-only mapping restrictions

`connectorgen validate internal/connectors/defs/notion --json` rejects the retained v2 `source_operation` payload, and `connectorgen surface-sync internal/connectors/defs --check` reports the absent canonical descriptor. Both results are intentionally retained in the matrix as mapping restrictions for all 49 source rows. They are not runtime-foundation gaps and this Track A change does not modify the shared importer or projection validator.
