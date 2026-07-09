# Plan: Bitbucket full-surface implementation

Parent issue: #79
Branch: `feat/79-bitbucket-cli-parity`
PR: https://github.com/polymetrics-ai/cli/pull/128
Date: 2026-07-09

## GSD command path

- `scripts/gsd doctor` — passed.
- `scripts/gsd list --json` — passed.
- `scripts/gsd prompt plan-phase issue-79-bitbucket-cli-parity --skip-research --tdd` — generated and followed for parent planning context.
- `scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-full-surface --dry-run` — unavailable: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual fallback remains active using `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `golang-graphql`
- `golang-lint`

## Objective

Move Bitbucket from representative parity (30 covered operations) to GitHub-style full typed surface coverage for the official 331-operation Swagger ledger. Preserve hard safety gates: no raw generic HTTP write, no shell/git/browser/local arbitrary executor, no secrets in docs/tests/fixtures/logs, no credentialed checks, and reverse ETL remains plan → preview → approval → execute.

## GitHub pilot pattern to mirror

- Implement operations as typed streams, direct reads, or reverse-ETL write actions.
- Leave only duplicate/deprecated/disallowed/internal/out-of-product-scope operations blocked.
- Keep raw API commands `unsafe_or_disallowed`.
- Model sensitive/admin/destructive writes as named reverse-ETL actions with risk text, redaction metadata, typed confirmation policy, and no generic arbitrary endpoint executor.

## Implementation approach

1. Add fail-first tests that require Bitbucket API coverage to approach full Swagger coverage and reject broad blocked ledgers.
2. Generate typed direct-read command entries for every GET endpoint not already covered by a stream/direct-read command.
3. Add a bounded Bitbucket binary direct-read output policy that returns base64 JSON with content metadata, no filesystem writes, no archive extraction, and max-byte enforcement.
4. Generate typed reverse-ETL write actions for every POST/PUT/DELETE endpoint not already modeled, using explicit operation names, path fields, shallow Swagger-derived record schemas, risk text, destructive confirmations, and sensitive policy metadata where applicable.
5. Update `api_surface.json` so every official endpoint has `covered_by` rather than blocked operation rows unless truly duplicate/deprecated/disallowed (target: 331/331 covered for this request).
6. Regenerate Bitbucket docs/catalog/website data and update GSD evidence.

## Safety constraints

- Do not expose `pm bitbucket api` or generic raw endpoint/method/body execution.
- Do not add generic shell, SQL write, browser, local git, or arbitrary local filesystem tools.
- Binary reads are API reads only and return bounded JSON/base64 to stdout; no destination path, overwrite, or extraction behavior.
- Sensitive record fields must be redacted in command-plan JSON and operation metadata.
- All mutations remain reverse-ETL write actions and require approval before execution.
- No live Bitbucket credentials/checks are used in verification.

## Verification checklist

- `jq . internal/connectors/defs/bitbucket/*.json internal/connectors/defs/bitbucket/schemas/*.json`
- `go test ./cmd/connectorgen -run Bitbucket -count=1`
- `go test ./internal/cli -run Bitbucket -count=1`
- `go test ./internal/connectors/engine -run DirectRead -count=1`
- `go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `gofmt -w cmd internal`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `./pm help bitbucket`, `./pm bitbucket`, `./pm bitbucket --help`
- `npm --prefix website run gen:website-data`
