# TDD ledger: Issue #3866

## Planned cycle

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Family matrix contract | `go test -count=1 -timeout 20m ./internal/synctransport -run '^TestTransportFamilyHalfPathConformance$'` failed with `undefined: transportFamilyConformanceCases` in `family_conformance_test.go`. | `go test -count=1 -timeout 20m ./internal/synctransport -run '^TestTransportFamilyHalfPathConformance'` passes 4 named source/destination family half-paths, each over `synccontract.AllModes()`, with exact staged/applied record IDs and committed checkpoint values. | The helper and every far-side executor are private test fixtures; no production surface is required. |
| Mode and invariant rows | The explicit unbound-source row initially failed during development because the matrix had no private fixture. | Focused `synctransport` and `coordination` tests assert `*DestinationSourceIneligibleError` with zero source/stage/plan/apply/commit calls; `ErrDownstreamAcknowledgementRequired` with no commit; `context.Canceled` with no apply/commit; `ErrAuthCohortFenced`; `ErrRateLimitParked`; and durable `ErrRateParkingConflict`, together with exact resume checkpoint/no replay. | All cases use memory stores/schedulers and connector-neutral fakes; none contact a provider/database. |
| Sensitivity | After `go run ./cmd/connectorgen validate internal/connectors/defs` passed (`552 connector(s) checked, 0 findings`), a scratch edit changed only the `api_source_to_warehouse` fixture's source family from `api` to `database`. `go test -count=1 -timeout 20m ./internal/synctransport -run '^TestTransportFamilyHalfPathConformance/api_source_to_warehouse/full_append$'` failed: `source family = "database", want "api"`. | The exact fixture binding was restored, then the same named command passed. | Scratch change was restored before any commit. |

No production edit has been made before this ledger.
