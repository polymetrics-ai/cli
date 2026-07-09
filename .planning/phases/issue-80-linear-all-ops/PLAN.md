# Plan — Issue #80 Linear full-surface all-ops update

Date: 2026-07-09
Branch: `feat/80-linear-cli-parity`
Prompt source: `PI_CONNECTOR_PROMPT.md`

## Objective

Update the Linear parity work to follow the refreshed prompt: every official Linear GraphQL operation must be either executable through a narrow typed surface (stream, direct read, reverse-ETL write, binary-read equivalent) or blocked only with exact duplicate/deprecated/disallowed/auth-internal/product-scope/engine-gap evidence.

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

1. RED: add tests that fail while most Linear mutation operation rows remain blocked with generic pending-review reasons and while all-ops coverage is below the refreshed prompt target.
2. GREEN: generate/curate typed fixed-document GraphQL write actions for every mutation whose variables can be safely represented by explicit scalar/enum/array record fields; update operation ledger coverage for those rows.
3. REFACTOR: leave only exact blocked evidence for raw GraphQL, auth-internal, deprecated/duplicate, binary/upload, or engine-gap rows. Update docs/website/planning.
4. VERIFY: focused Linear all-ops tests, connectorgen validation, Linear conformance, `go vet`, `go test`, `go build`, `make verify`, docs/help parity.
