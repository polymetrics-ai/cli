# TDD Ledger — Issue #151 Chatwoot stream runner

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestReadFanOutRequestPaginationOverrideAllowsDifferentChildPagination -count=1` | Fan-out parent id-list request reuses child cursor pagination and cannot model Chatwoot messages' parent conversation page sweep independently. | Add optional `fan_out.ids_from.request.pagination` and use it for id-listing requests. | planned |
| 2 | `go test ./internal/connectors/conformance -run TestChatwootStreamRunnerSweep -count=1` | Chatwoot message stream still models message pagination with legacy `page` query and cursor metadata does not prove request-param resume. | Update Chatwoot messages stream to official `after` cursor pagination, message cursor field to `id`, and fixtures to include empty cursor-stop pages. | planned |
| 3 | `go run ./cmd/connectorgen validate internal/connectors/defs` | New DSL field must be accepted by `streams.schema.json`; Chatwoot fixtures/schema must remain valid. | Update schema and bundle structs; validate all connector defs. | planned |

## Notes

- Red/green work is fixture-only and non-credentialed.
- Generic API/write escape hatches remain disallowed; this slice only affects stream reads and declarative pagination metadata.
