# Issue #154 — Chatwoot advanced query / binary engine

Refs #148 / #154.

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

Implement the advanced read/query side of the remaining Chatwoot official surface without exposing raw generic HTTP:

1. Extend bounded direct reads to safely address root-relative official paths (`/api/v2`, `/platform`, `/public`) from the configured Chatwoot origin without duplicating the account API base path.
2. Add typed command metadata for remaining official Chatwoot GET operations that can be represented as bounded JSON direct reads.
3. Add path/query flags from the official Swagger metadata where available; path variables remain required by the direct-read executor.
4. Keep disallowed/duplicate rows blocked, and leave write/admin/sensitive/destructive mutation enablement to #155.
5. Refresh docs/manual/website generated data and lock counts with tests.

## Out of scope

- No generic raw HTTP command, arbitrary URL/path input, shell tool, SQL write, or generic write action.
- No binary download to disk; Chatwoot's current official inventory has no safe non-JSON binary download endpoint in the ledger. Multipart avatar/profile write remains blocked for #155/#binary policy.
- No sensitive/admin/destructive write execution; #155 owns mutation policy.

## TDD plan

1. Red: add direct-read test proving root-relative official paths dispatch to the configured Chatwoot origin instead of under `/api/v1/accounts/{account_id}`.
2. Red: add Chatwoot direct-read surface test expecting all non-platform/public/report safe GETs plus explicitly bounded report/public/platform GETs to be covered by direct-read commands.
3. Green: update engine root-relative direct-read path handling and generate Chatwoot direct-read command/API coverage metadata.
4. Green: regenerate docs/website data and update operation-ledger count tests.

## Verification checklist

- `go test ./internal/connectors/engine -run DirectRead -count=1`
- `go test ./internal/connectors/commandrunner -run DirectRead -count=1`
- `go test ./cmd/connectorgen -run Chatwoot -count=1`
- `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- `cd website && npm run test:unit -- --run tests/api/connector-data.test.ts`
- `git diff --check`
- Full handoff gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`

## Manual GSD fallback

`2026-07-10`: `scripts/gsd prompt programming-loop ...` and `scripts/gsd prompt quick-validate ...` both returned `scripts/gsd: unknown GSD command`, so this issue uses the repo planning/TDD/verification artifacts as the manual GSD fallback. `scripts/gsd list --json` and `scripts/gsd doctor` outputs were captured under `traces/`.
