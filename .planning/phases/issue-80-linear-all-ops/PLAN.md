# Plan — Issue #80 Linear full-surface all-ops update

Date: 2026-07-09
Branch: `feat/80-linear-cli-parity`
Prompt source: `PI_CONNECTOR_PROMPT.md`

## Objective

Update the Linear parity work to follow the refreshed prompt: every official non-deprecated Linear GraphQL field (514 total: 156 query, 358 mutation) must be either executable through a narrow typed surface (stream, direct read, reverse-ETL write, binary-read equivalent) or blocked only with exact duplicate/deprecated/disallowed/auth-internal/product-scope/engine-gap evidence. Deprecated live-schema rows should also be inventoried or explicitly blocked so no root field is silently absent.

## GSD path

- `scripts/gsd doctor` — pass.
- `scripts/gsd verify-pi` — pass.
- `scripts/gsd list --json` — pass.
- `scripts/gsd prompt programming-loop issue-80-linear-all-ops --skip-research` — unavailable (`unknown GSD command: programming-loop`).
- Manual PLAN → RED → GREEN → REFACTOR → VERIFY fallback remains active.

## Skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-graphql`, `golang-documentation`.

## Safety constraints

- No secrets, no credentialed Linear checks, no live Linear writes.
- No raw arbitrary GraphQL query/mutation execution.
- No generic JSON body write escape hatch.
- Reverse ETL remains plan → preview → approval → execute.
- Admin/destructive/sensitive mutations need typed actions with risk/confirmation metadata or exact blocked evidence when the current schema/engine lacks required policy support.

## TDD plan

1. RED: tighten Linear ledger tests to the prompt's official counts (156 query + 358 mutation non-deprecated fields) and live-schema inventory counts so the current 465-field SDK-document slice fails.
2. GREEN: generate/curate fixed-document GraphQL streams for missing query fields and typed reverse-ETL write actions for missing non-deprecated mutation fields using explicit record schemas; add exact deprecated blocked evidence for the remaining deprecated live-schema gap.
3. REFACTOR: keep raw arbitrary GraphQL as disallowed, keep docs/website/planning counts aligned with the prompt, and avoid adding raw GraphQL or generic JSON write escapes.
4. VERIFY: focused Linear all-ops tests, connectorgen validation, Linear conformance, `go vet`, `go test`, `go build`, `make verify`, docs/help parity.
