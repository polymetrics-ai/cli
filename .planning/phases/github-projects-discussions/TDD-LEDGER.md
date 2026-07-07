# TDD Ledger

Phase: github-projects-discussions

Record failing test evidence before production code for every behavior-adding task.

## Red Evidence

Command:

```bash
go test ./internal/connectors/engine -run 'TestReadGraphQLBodyResolvesRequestQueryVariables|TestReadGraphQLBodyOmitsEmptyOptionalVariable|TestBundleLoadParsesGraphQLVariableOmitWhenEmpty|TestBundleLoadRejectsGraphQLVariableOmitWhenEmptyNonBoolean|TestGitHubProjectsDiscussionsCommandsMapToGraphQLStreams'
```

Result: failed as expected.

Expected failures:

- `query.number` is an unknown interpolation namespace.
- `omit_when_empty` is rejected as an unsupported GraphQL variable template key.
- GitHub lacks `projects`, `project_items`, `discussions`, and `discussion` GraphQL streams.
