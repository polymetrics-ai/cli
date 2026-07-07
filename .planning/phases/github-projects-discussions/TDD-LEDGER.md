# TDD Ledger

Phase: github-projects-discussions

Record failing test evidence before production code for every behavior-adding task.

## Red Evidence (initial phase)

Command:

```bash
go test ./internal/connectors/engine -run 'TestReadGraphQLBodyResolvesRequestQueryVariables|TestReadGraphQLBodyOmitsEmptyOptionalVariable|TestBundleLoadParsesGraphQLVariableOmitWhenEmpty|TestBundleLoadRejectsGraphQLVariableOmitWhenEmptyNonBoolean|TestGitHubProjectsDiscussionsCommandsMapToGraphQLStreams'
```

Result: failed as expected.

Expected failures:

- `query.number` is an unknown interpolation namespace.
- `omit_when_empty` is rejected as an unsupported GraphQL variable template key.
- GitHub lacks `projects`, `project_items`, `discussions`, and `discussion` GraphQL streams.

## Red Evidence (review-fix slice)

Command:

```bash
go test ./internal/connectors/engine -run 'TestReadGraphQLBodyUsesDefaultForMissingQueryVariable|TestBundleLoadRejectsGraphQLVariableDefaultTypeMismatch'
```

Expected failures before fixes:

- No test exists for an explicitly-empty query variable (`req.Query["number"] = ""`) paired with
  `default` / `omit_when_empty`.
- No bundle validation rejects a GraphQL variable whose `default` string does not match its declared
  `type` (`integer`, `number`, `boolean`).
