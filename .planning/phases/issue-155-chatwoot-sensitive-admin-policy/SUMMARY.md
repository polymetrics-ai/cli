# Summary — Issue #155 Chatwoot sensitive/admin/destructive policy

Status: integrated into parent as `debf010a`; parent review fallback pending.

Completed slice:

- Added root-relative fixed write path handling for `/platform`, `/public`, and `/api/v1/profile` endpoints while preserving existing account-scoped relative write paths.
- Added typed reverse-ETL write action coverage for all remaining non-disallowed/non-duplicate official Chatwoot operation rows.
- Chatwoot now covers 139 of 144 official operations: 7 stream endpoints, 53 bounded direct reads, and 79 typed write endpoints. Only 4 disallowed rows and 1 duplicate row remain blocked.
- Destructive actions include `confirm: destructive` metadata and stay behind reverse ETL plan → preview → approval → execute.
- Updated Chatwoot API/CLI surface, writes, docs/manual/skill, website data, and tests.

Verification passed: targeted checks plus `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify`.

Parent post-integration verification passed: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate internal/connectors/defs`.
