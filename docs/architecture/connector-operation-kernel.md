# Connector Operation Kernel

Status: foundation slice for GitHub CLI parity (#56).

## Purpose

`operations.json` is optional per connector metadata for provider-style CLI
commands that are not naturally represented by a stream or existing write
action alone. It is a reviewed command execution contract, not a raw escape
hatch.

Command surface entries may reference exactly one executable target:

- `stream`
- `write`
- `operation`

API surface rows that are already executable through fixed direct-read command
metadata use `covered_by.direct_read` or `covered_by.direct_reads`. Blocked
`api_surface.operation` rows remain ledger-only and are not an execution
allowlist.

The #56 foundation loads and validates operation metadata but keeps operation
execution blocked by default. Later issues add executors for fixed REST,
GraphQL, XML, binary/file, local git, local file, browser, and composite
operations.

## Supported Operation Kinds

- `stream_etl`
- `rest_read`
- `rest_write`
- `graphql_query`
- `graphql_mutation`
- `xml_export`
- `xml_import`
- `binary_download`
- `file_upload`
- `local_git`
- `local_file`
- `browser_open`
- `composite`

Unknown kinds are rejected at load time. There is intentionally no generic
shell, unrestricted HTTP write, generic SQL write, or arbitrary GraphQL kind.

## Safety Contract

- Operations must be fixed, connector-scoped definitions.
- Mutations must keep plan, preview, approval, execute.
- Secrets must not appear in operation metadata, fixtures, logs, examples, or
  review comments.
- GraphQL operations must use fixed documents and checked variables.
- File and binary operations must define bounded output policy before becoming
  executable.
- Local git/file operations must use allowlisted structured actions, never a
  shell string.
- Generated candidates from provider specs are not executable until reviewed
  and promoted to production metadata.

## Runtime Behavior In #56

If a command references `operation`, `commandrunner` returns a blocked command
error naming the operation ID and explaining that its executor is not yet
implemented. This fail-closed behavior is deliberate: it lets docs, validation,
and parity planning land before any new side-effecting executor is available.

## Example

```json
{
  "id": "github.projects.list",
  "kind": "graphql_query",
  "summary": "List GitHub Projects using a fixed GraphQL query.",
  "risk": "low",
  "approval": "none",
  "output_policy": "json",
  "graphql": {
    "operation_name": "ListProjects",
    "document": "query ListProjects($owner: String!, $first: Int!, $after: String) { organization(login: $owner) { projectsV2(first: $first, after: $after) { nodes { id number title url closed updatedAt } pageInfo { hasNextPage endCursor } } } }"
  }
}
```
