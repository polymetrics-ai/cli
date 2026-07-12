# TDD Ledger — Issue #151 Chatwoot stream runner

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestReadFanOutRequestPaginationOverrideAllowsDifferentChildPagination -count=1` | Captured red build failure: `unknown field Pagination in struct literal of type FanOutIDsRequest`. | Added optional `fan_out.ids_from.request.pagination` to bundle schema/structs and used it for id-listing requests. | passed |
| 2 | `go test ./internal/connectors/conformance -run TestChatwootStreamRunnerSweep -count=1` | Locked the Chatwoot stream sweep after cycle 1 made the DSL expressible; this would have failed against legacy `messages` page pagination/cursor metadata. | Updated Chatwoot messages stream to official `after` cursor pagination, message cursor field to `id`, and fixtures to include empty cursor-stop pages. | passed |
| 3 | `go run ./cmd/connectorgen validate internal/connectors/defs` | New DSL field had to be accepted by `streams.schema.json`; Chatwoot fixtures/schema had to remain valid. | Updated schema and bundle structs; validated all connector defs. | passed |

## Notes

- Red/green work is fixture-only and non-credentialed.
- Generic API/write escape hatches remain disallowed; this slice only affects stream reads and declarative pagination metadata.
