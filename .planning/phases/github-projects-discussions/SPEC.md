# SPEC: GitHub Projects And Discussions GraphQL Reads

## Goal

Implement the first GraphQL-backed GitHub CLI parity slice for safe read commands:

- `pm github project list`
- `pm github project item-list --project-id <id>`
- `pm github discussion list`
- `pm github discussion view --number <n>`

Project and discussion mutations remain planned/direct-write until fixed mutation schemas and
reverse-ETL approval policies exist.

## Runtime Contract

- Reads are stream-backed ETL commands, not operation-executor commands.
- GraphQL documents are fixed bundle metadata; command flags may only bind variables.
- GraphQL response `errors[]` remains fail-closed.
- Cursor variables may be omitted on the first page to preserve nullable GraphQL variable semantics.
- Command flag values reach GraphQL variables through `query.*` interpolation.
- GitHub Projects require `read:project` scope for private/project data.

## Non-Goals

- Generic GraphQL endpoint execution.
- Project or discussion mutation execution.
- Generic operation ledger executor.
- New dependencies.

## Sources

- GitHub GraphQL API overview: https://docs.github.com/en/graphql
- GitHub Projects API docs: https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects
- GitHub Discussions GraphQL docs: https://docs.github.com/en/graphql/guides/using-the-graphql-api-for-discussions
