---
phase: github-parity-extract-r1
plan: "09"
type: tdd
wave: 9
depends_on: ["08"]
autonomous: true
files_modified:
  - internal/connectors/connectors.go
  - internal/connectors/engine/bundle.go
  - internal/connectors/engine/connector.go
  - internal/connectors/engine/direct_read.go
  - internal/connectors/engine/direct_write.go
  - internal/connectors/engine/graphql_operation.go
  - internal/connectors/engine/operation_endpoint_ledger.go
  - internal/connectors/engine/write_gate.go
  - internal/connectors/engine/*_test.go
  - internal/connectors/commandrunner/runner.go
  - internal/connectors/commandrunner/*_test.go
  - internal/connectors/engine/schema/operations.schema.json
  - internal/app/github_sync_modes_test.go
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
---

# Plan — typed GraphQL operation runtime and write gate

## Goal

Make a bundle-declared GraphQL `Query` or `Mutation` executable through the same bounded direct
read and reverse-ETL plan/preview/approval/typed-confirmation paths as REST operations. This is a
provider-neutral engine capability; it accepts only a fixed document, fixed connector-relative
endpoint, and a schema-validated variables object from a declared operation. It must not introduce
a caller-supplied GraphQL document, selection, endpoint, header, or generic write escape hatch.

This is runtime infrastructure, not a claim that all 305 GitHub GraphQL roots are already emitted
as commands. The source ledger remains the authority for each root's current factual state.

## Contract

- A `graphql_query` operation has a fixed `graphql` declaration: operation name, document,
  connector-relative endpoint, response cap, and closed variables schema. Direct reads may supply
  only values accepted by that schema. `--page-cursor` is permitted only for a declared cursor
  contract; `--page` remains refused.
- A `graphql_mutation` uses the same fixed declaration and variables schema, but is prepared and
  executed only through `OperationDirectWrite`: preview binds the exact request, approval is
  single-use, and the existing destructive confirmation derives from the operation's mutation
  class/destructive metadata. Its command binding is the declaration-owned physical `POST`
  `graphql.path` endpoint; a command whose API surface differs is rejected before plan creation.
  It never accepts a URL or raw transport selector.
- Query responses retain bounded `data` even when GitHub supplies an `errors[]` array, and expose
  a sanitized partial-response indicator plus at most the bounded error metadata. A mutation with
  any GraphQL `errors[]` fails its approved write rather than reporting completion. Both paths
  preserve response bounds and existing redaction.
- The runtime extracts declared GraphQL rate-limit fields when present and reports them as response
  metadata; rate-limit accounting remains in the existing requester/limiter rather than a
  GitHub-only path.

## TDD slices

### Red 29a — closed GraphQL query request and partial response

Add an `httptest` bundle with a fixed query, endpoint, nested variables schema, cursor declaration,
and a response containing both `data` and `errors`. Require that the engine sends only the fixed
document/operation name, rejects undeclared variables and unsupported page navigation before the
server is reached, carries partial data/error metadata, caps response bytes, and derives the next
cursor from the declared connection path.

### Red 29b — GraphQL mutation shares the approved write gate

Add a fixed mutation test that proves preview has no network call; unapproved/replayed approval has
no network call; approved execution sends the fixed document plus typed variables exactly once; a
GraphQL error envelope fails the operation; and destructive confirmation remains required whenever
the declared mutation is destructive. Include a command-runner preflight test that refuses a
physical POST/path binding unless the declared `graphql_mutation` matches it exactly.

### Green 29 — provider-neutral GraphQL execution

Implement the strict schema/runtime support and operation metadata needed for the tests. Keep
existing REST behavior byte-for-byte compatible. Update the operation schema and command runner
only as necessary to make the real lifecycle accept a fixed GraphQL operation; do not add a generic
`pm graphql`, raw document flag, or arbitrary endpoint/HTTP command.

### Verification repair — local GitHub ETL fixture rate-limit cohort

The full local `internal/app` suite may exercise GitHub's declared authenticated-user rate-limit
policy. Its loopback credential must therefore provide the policy's required non-secret
`rate_limit_account` cohort key. If it is absent, repair the test fixture with a fixed synthetic
identifier and retain the full five-sync-mode test; never disable the policy or weaken the ETL
assertions to make the runtime gate pass.

## Verification

- No provider/PM/browser/GitHub request occurs in this phase; all tests use `httptest`.
- Run focused engine and command-runner tests, then `go test -timeout 20m ./internal/connectors/engine`,
  `go test -timeout 20m ./internal/connectors/commandrunner`, `go test -timeout 20m ./internal/app`,
  `go run ./cmd/connectorgen validate internal/connectors/defs`, `surface-sync --check`, the source
  ledger check, and `git diff --check`.
- Required skills: `golang-how-to`, `golang-graphql`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, and `golang-cli`.
