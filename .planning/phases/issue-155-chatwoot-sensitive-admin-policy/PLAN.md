# Issue #155 — Chatwoot sensitive/admin/destructive policy

Refs #148 / #155.

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

Implement typed, safety-gated coverage for the remaining official Chatwoot mutation/post-query operation ledger rows while preserving the product safety model:

1. Add fixed `writes.json` actions for remaining non-disallowed/non-duplicate Chatwoot operations.
2. Keep all write execution behind reverse-ETL plan → preview → approval → execute; destructive actions additionally expose a typed confirmation challenge.
3. Add root-relative write path resolution for fixed official `/platform`, `/public`, and `/api/v1/profile` endpoints without adding a raw URL/path escape hatch.
4. Generate command metadata, docs/manual/skill, website data, and operation-ledger metrics so every official operation is covered or intentionally blocked.
5. Leave only duplicate/disallowed rows blocked in `api_surface.json`.

## Out of scope / safety constraints

- No generic HTTP write command, arbitrary endpoint/path/URL flag, shell, SQL write, or binary download tool.
- No credential values in command examples or docs; secret-like write fields are redacted in preview metadata and should be supplied from approved data/credential flows, not prompt text.
- Multipart/binary variants are not enabled as raw file upload flags. JSON/form-compatible typed operations are covered; attachment/profile-avatar binary behavior remains represented only by typed non-binary fields unless a future dedicated binary policy is added.

## TDD plan

1. Red: add write-engine test proving a fixed root-relative write path routes to the configured origin root, not below `/api/v1/accounts/{account_id}`.
2. Red: update Chatwoot API surface metrics to expect only duplicate/disallowed rows left as blocked operations.
3. Green: add root-relative write path resolution, generated write actions/commands/coverage, docs, and tests.
4. Green: validate connector defs, commandrunner write preflight, docs, and website generated data.

## Verification checklist

- `go test ./internal/connectors/engine -run Write -count=1`
- `go test ./internal/connectors/commandrunner -run Write -count=1`
- `go test ./cmd/connectorgen -run Chatwoot -count=1`
- `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- `go test ./internal/connectors/bundleregistry -run Chatwoot -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- `cd website && pnpm run gen:website-data`
- `cd website && pnpm run test:unit -- --run tests/api/connector-data.test.ts`
- `git diff --check`
- Full handoff gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`

## Manual GSD fallback

`2026-07-10`: `scripts/gsd prompt programming-loop ...` and `scripts/gsd prompt quick-validate ...` both returned `scripts/gsd: unknown GSD command`, so this issue uses repo planning/TDD/verification artifacts as the manual GSD fallback. `scripts/gsd list --json` and `scripts/gsd doctor` outputs were captured under `traces/`.
