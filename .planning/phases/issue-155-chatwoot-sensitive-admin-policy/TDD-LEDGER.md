# TDD Ledger — Issue #155 Chatwoot sensitive/admin/destructive policy

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestWriteRootRelativeEndpointUsesConfiguredOrigin -count=1` | Captured red failure: root-relative write under a scoped base path dispatched below `/api/v1/accounts/{account_id}` instead of the Chatwoot origin root. | Updated write request path resolution for fixed root-relative official paths while preserving existing scoped relative write paths. | passed |
| 2 | `go test ./cmd/connectorgen -run TestChatwootAPISurfaceOperationLedgerMetrics -count=1` | Captured red failure after metadata generation until expected counts moved from blocked admin/sensitive/destructive rows to write coverage. | Generated typed write coverage for all non-disallowed/non-duplicate operation rows. | passed |
| 3 | `go run ./cmd/connectorgen validate internal/connectors/defs` | Validator initially found required object fields without CLI mappings; added fixed metadata mappings for required fields. | Validated all connector defs. | passed |

## Notes

- No live Chatwoot credentials are used.
- Destructive commands keep typed confirmation metadata and remain subject to reverse ETL approval flow.
