# Summary — Issue #154 Chatwoot advanced query / binary engine

Status: full verified; stacked PR pending.

Completed slice:

- Added root-relative direct-read origin handling so official `/api/v2`, `/platform`, and `/public` Chatwoot paths route to the configured Chatwoot origin rather than under the account API base path.
- Converted all remaining official Chatwoot `direct_read` ledger rows into bounded JSON direct-read command coverage.
- Chatwoot now has 53 implemented bounded JSON direct-read commands, plus 7 streams and 6 existing write actions.
- Kept duplicate/disallowed rows blocked and left mutation/admin/sensitive/destructive write enablement to #155.
- Updated Chatwoot API/CLI surface, generated manual/skill/catalog/website data, docs, and tests.

Verification passed: targeted checks plus `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify`.
