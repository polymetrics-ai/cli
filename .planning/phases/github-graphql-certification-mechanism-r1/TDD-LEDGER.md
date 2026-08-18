# TDD ledger — GraphQL certification mechanism

## Baseline

- Branch base: `integration/4015-mvp-flat-r1` at `e7ae907ec6962920ebf42dc52c27c6014de6031a`.
- Existing sweep reports 305 GitHub GraphQL commands as fixture-required. This is an input to verify, not a count to trust.
- Required skills: `golang-how-to`, `golang-graphql`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`.

## Red — observed before production implementation

Command:

```text
go test -timeout 20m -run TestGraphQLCertificationInventoryRejectsIncorrectProducedValueAfterSchemaConformance ./internal/connectors/certify
```

Result:

```text
internal/connectors/certify/stages_graphql_schema_internal_test.go:6:20: undefined: graphQLCertificationInventoryFor
FAIL polymetrics.ai/internal/connectors/certify [build failed]
```

The test fixes the intended externally observable contract before implementation: compile the connector-owned source-pinned schema inventory, classify all 305 generated commands as 29 schema-conformant / 2 live-required / 274 fixture-bound, then evaluate the connector-owned `viewer.__typename` produced-value assertion. The later negative run changes only that expected value after schema compilation; it must fail and is restored before green verification.

## Green — observed after implementation

Restored declaration command:

```text
go test -timeout 20m -run TestGraphQLCertificationInventoryRejectsIncorrectProducedValueAfterSchemaConformance -v ./internal/connectors/certify
```

Result: `PASS` (`ok polymetrics.ai/internal/connectors/certify`). The compiled source-pinned schema inventory reports exactly 29 schema-conformant, 2 live-required, and 274 fixture-bound rows.

Post-compilation negative proof changed only `/response/viewer/__typename` from `User` to `Organization` in the connector definition, then ran the same test. It compiled the schema inventory and failed at the assertion:

```text
produced-value assertion failed after schema conformance: graphql_query_viewer: declared output at /response/viewer/__typename does not match
FAIL
```

The connector-owned assertion was restored to `User`, and the same test passed again. The unexecutable seam test records `Status=unexecutable` rather than a benign skip when a declared source lock cannot compile.

## Refactor — observed

Shared code is input-driven: no provider identifier, endpoint, query document, or assertion is present outside the connector definition. `go run ./cmd/connectorgen boundary` passed.
