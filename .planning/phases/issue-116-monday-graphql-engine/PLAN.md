# Plan — issue #116 Monday GraphQL/direct-read engine

## Objective

Add a constrained fixed-document GraphQL direct-read execution path that can run vetted operation-ledger `graphql_query` operations from CLI direct-read commands without exposing arbitrary GraphQL or writes.

## GSD mode

- `scripts/gsd prompt plan-phase issue-116-monday-graphql-engine --skip-research` generated this prompt.
- `programming-loop` command unavailable; manual TDD fallback active.

## Slice

1. Red tests: commandrunner allows implemented direct-read commands with a typed `operation`; engine executes a fixed GraphQL query through POST, passes declared flag variables, fails closed on GraphQL `errors`, enforces response byte limit, and rejects missing/non-query/unsafe operations.
2. Green: add `graphql_json` direct-read output policy and an `Operation` field on `DirectReadRequest`; route only `graphql_query` operation specs, never mutations.
3. Refactor: keep REST direct reads unchanged; no raw GraphQL document input from users.

## Safety

- Fixed document comes from bundled `operations.json` only.
- User flags can only supply declared variables; no user-provided GraphQL document.
- Mutations remain blocked; GraphQL `errors` fail closed; max response bytes are clamped.

## Verification

```bash
go test ./internal/connectors/engine -run 'TestDirectReadGraphQL' -count=1
go test ./internal/connectors/commandrunner -run 'TestRunDirectReadGraphQLOperation' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```
