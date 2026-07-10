# Issue #153 — Chatwoot direct read (CLI parity)

Refs #148 / #153.

## Required skills loaded

- `gsd-core` (manual fallback; `scripts/gsd prompt programming-loop` unavailable)
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-structs-interfaces`
- `golang-context`
- `golang-documentation`
- CLI/help/docs/website parity reference

## Scope

Implement a bounded, typed direct-read slice for Chatwoot without adding any raw generic HTTP escape hatch:

1. Add a connector-fixed `bounded_json` direct-read output policy that caps response bytes and recursively redacts secret-shaped keys.
2. Ensure direct-read execution works for connectors whose declarative base URL includes a scoped path prefix (Chatwoot base is `/api/v1/accounts/{account_id}` while official surface paths include the same prefix).
3. Mark safe Chatwoot direct reads implemented:
   - `conversation view` → `GET /api/v1/accounts/{account_id}/conversations/{conversation_id}`
   - `contact view` → `GET /api/v1/accounts/{account_id}/contacts/{id}`
   - `contact search` → `GET /api/v1/accounts/{account_id}/contacts/search?q=...`
4. Keep reports, audit logs, public inbox reads, binary/file endpoints, and admin/sensitive/destructive operations blocked for later slices.
5. Update API-surface coverage, CLI surface metadata, docs/manual/website data, and tests.

## Out of scope

- No generic HTTP command or arbitrary URL/path input.
- No unbounded pagination sweep for reports/audit logs.
- No binary/file download or multipart output.
- No reverse-ETL write/admin/sensitive policy changes.

## TDD plan

1. Red: engine direct-read test for `bounded_json` redaction fails because the policy is unsupported.
2. Red: engine direct-read test for scoped base-path stripping fails because Chatwoot-like paths duplicate `/api/v1/accounts/{account_id}`.
3. Red: commandrunner/CLI tests for Chatwoot implemented direct reads fail while commands remain planned/unsupported.
4. Green: implement bounded JSON policy, scoped path handling, Chatwoot metadata coverage, docs/generated data, and tests.

## Verification checklist

- `go test ./internal/connectors/engine -run DirectRead -count=1`
- `go test ./internal/connectors/commandrunner -run DirectRead -count=1`
- `go test ./internal/cli -run Chatwoot -count=1`
- `go test ./cmd/connectorgen -run Chatwoot -count=1`
- `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- `cd website && npm run test:unit -- --run tests/api/connector-data.test.ts`
- `git diff --check`
- Full handoff gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`

## Manual GSD fallback

`2026-07-10`: `scripts/gsd prompt programming-loop ...` and `scripts/gsd prompt quick-validate ...` both returned `scripts/gsd: unknown GSD command`, so this issue uses the repo planning/TDD/verification artifacts as the manual GSD fallback. `scripts/gsd list --json` and `scripts/gsd doctor` outputs were captured under `traces/`.
