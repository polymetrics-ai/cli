# Summary: Chatwoot Operation Ledger

Status: implemented and locally verified.

- Created #152 GSD/TDD plan and verification checklist before production edits.
- Added `TestChatwootAPISurfaceOperationLedgerMetrics` covering 144 official operations, 89 paths, method/model/risk/status counts, covered stream/write rows, and blocked-by-default invariants.
- Tightened `PUT /api/v1/profile` operation metadata to call out blocked multipart avatar/profile policy.
- Scope stayed operation-ledger metadata and tests only; no runtime execution, direct reads, binary transfer, or new writes were added.
- Verification passed: targeted Chatwoot/GitHub API surface tests, connectorgen validation, Chatwoot conformance, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `git diff --check`.
