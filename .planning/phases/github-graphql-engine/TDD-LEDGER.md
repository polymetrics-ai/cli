# TDD Ledger

Phase: github-graphql-engine

Record failing test evidence before production code for every behavior-adding task.

## Planned Red Tests

| Behavior | Test | Expected initial failure |
| --- | --- | --- |
| GraphQL read body | `TestReadGraphQLBodySendsFixedDocumentAndVariables` | stream body not sent / GraphQL field missing |
| GraphQL read errors | `TestReadGraphQLErrorsFailClosed` | top-level `errors[]` ignored |
| GraphQL write body | `TestWriteGraphQLBodyUsesFixedDocumentAndDeclaredVariables` | `body_type: graphql` treated as JSON record body |
| GraphQL document safety | `TestWriteGraphQLBodyIgnoresRecordQueryField` | record `query` leaks into body before fixed document support |

## Evidence

### Red 1 — GraphQL request spec missing

Command:

```bash
go test ./internal/connectors/engine -run 'TestReadGraphQL|TestWriteGraphQL'
```

Result: failed as expected.

Key failure:

```text
unknown field GraphQL in struct literal of type StreamSpec
undefined: GraphQLRequestSpec
unknown field GraphQL in struct literal of type WriteAction
```

Conclusion: production engine bundle types do not yet model safe fixed GraphQL request payloads for
read streams or write actions.

### Green 1 — GraphQL request spec implemented

Command:

```bash
go test ./internal/connectors/engine -run 'TestReadGraphQL|TestWriteGraphQL|TestBundleLoad.*GraphQL'
```

Result: passed.

Covered:

- fixed GraphQL read payload with declared variables
- GraphQL read top-level `errors[]` fail-closed behavior
- fixed GraphQL write payload with declared variables
- record-provided `query` cannot override the fixed bundle document
- GraphQL write top-level `errors[]` fail-closed behavior
- bundle load validation for GraphQL stream/write shapes
- typed GraphQL variables can coerce declared templates to `integer`, `number`, or `boolean`
