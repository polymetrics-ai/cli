# Connector Operation Kernel

Status: operation metadata and fixed direct-read execution foundation.

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

Operation metadata is loaded and validated for every bundle. A command becomes
executable only when its `cli_surface.json` entry is `availability: "implemented"`,
`intent: "direct_read"`, references a `rest_read` operation, and the connector
implements `OperationDirectReader`; otherwise `commandrunner` returns a blocked
command error. Implemented operation direct reads are bounded to connector-relative
GET/POST REST endpoints, require a supported `output_policy`, and reject raw
method/path/body flags. Non-direct-read operation kinds remain planned until their
typed executors land.

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

## Runtime Behavior

If a command references an operation outside the implemented direct-read contract,
`commandrunner` returns a blocked command error naming the operation ID and why the
executor is unavailable. This fail-closed behavior is deliberate: docs, validation,
and parity planning can land before any new side-effecting executor is available.

## Example

```json
{
  "id": "github.projects.list",
  "kind": "graphql_query",
  "summary": "List GitHub Projects using a fixed GraphQL query.",
  "risk": "low",
  "approval": "none",
  "output_policy": "json_redacted",
  "graphql": {
    "operation_name": "ListProjects",
    "document": "query ListProjects($owner: String!, $first: Int!, $after: String) { organization(login: $owner) { projectsV2(first: $first, after: $after) { nodes { id number title url closed updatedAt } pageInfo { hasNextPage endCursor } } } }"
  }
}
```
