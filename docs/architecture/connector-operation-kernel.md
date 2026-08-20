# Connector Operation Kernel

Started as the foundation slice for GitHub CLI parity (#56). The executor
contract below describes what the runtime does today; `commandrunner` and
`internal/connectors/engine` remain the authority when this document and the
code disagree.

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
`api_surface.operation` rows remain ledger-only. An API-surface operation row
is never an execution allowlist: an implemented direct write must independently
pass its declared operation and command preflight before it can enter the write
lifecycle.

Operation execution is opt-in per intent, not blanket-enabled: an operation runs
only through an intent whose executor exists. Direct reads, declared `rest_write`
direct writes, and bounded binary downloads have executors today; GraphQL, XML,
local git, local file, browser, and composite operations do not, and remain
blocked.

## Supported Operation Kinds

- `stream_etl`
- `rest_read`
- `rest_status`
- `rest_write`
- `provider_search`
- `graphql_query`
- `graphql_mutation`
- `xml_export`
- `xml_import`
- `binary_download`
- `text_export`
- `file_upload`
- `local_git`
- `local_file`
- `browser_open`
- `composite`

Unknown kinds are rejected at load time. There is intentionally no generic
shell, unrestricted HTTP write, generic SQL write, or arbitrary GraphQL kind.

`provider_search` is a read that carries a fixed POST body containing bounded
lists; every array must declare `maxItems`, the body schema must be closed
(`additionalProperties: false`), and the method/path are fixed by the bundle.
It is a distinct kind rather than a convention over `rest_read` so its bound
rules are enforceable at load time.

## Safety Contract

- Operations must be fixed, connector-scoped definitions.
- Mutations must keep plan, preview, approval, execute.
- Secrets must not appear in operation metadata, fixtures, logs, examples, or
  review comments.
- GraphQL operations must use fixed documents and checked variables.
- Binary and file operations are bounded by a byte cap and an explicit
  caller-supplied destination, never by an output policy: the response becomes a
  file on disk, not a JSON body.
- A caller-provided operation header is an exact, declared bounded string
  parameter. Authorization, cookie, routing, connection/proxy, and other
  runtime-owned headers are never caller-selectable. Declared response metadata
  is bounded by name and byte cap; every ordinary admitted value is retained,
  while established credential/transport-secret headers retain presence with an
  explicit redaction marker.
- Local git/file operations must use allowlisted structured actions, never a
  shell string.
- Generated candidates from provider specs are not executable until reviewed
  and promoted to production metadata.

## Runtime Behavior

`commandrunner` decides whether an operation-backed command executes, and it
decides before any network or filesystem access:

- `intent:"direct_read"` with `availability:"implemented"` executes as a bounded
  REST read under the command's `output_policy`. It issues exactly one request
  and returns one page: the page size and the next-page context are derived from
  the connector's own declared `streams.json` pagination spec, the result
  reports whether that page is the whole collection, and the runtime-owned
  `--page`/`--page-cursor` flags reach the rest. See AGENTS.md, "Direct Reads
  Return One Page, And Say So".
- `intent:"direct_write"` with `availability:"implemented"` executes one bounded
  `rest_write` only through the connector-command plan → preview → approval →
  execute lifecycle. The command and operation must declare matching, explicit
  output policies; their intent-specific choices are defined in the
  [connector authoring conventions](../migration/conventions.md#2-authoring-rules).
- `intent:"binary_download"` with `availability:"implemented"` executes through
  `connectors.OperationBinaryDownloader`, which the declarative engine satisfies
  with `engine.OperationBinaryDownload`. The endpoint must be a single
  connector-relative GET, the caller must supply a destination root, and the
  byte cap is the request value clamped by the operation's declared maximum and
  then by the engine's own ceiling. The operation declares its accepted success
  statuses, response media types, optional bounded response headers, and any
  bounded redirect policy before a file can be created.
- `intent:"status_check"` with `availability:"implemented"` executes one
  declared connector-relative HEAD operation and returns its fixed status plus
  only declared bounded response metadata, never a body.
- A `text_export` operation uses the bounded download executor with a declared
  CSV media type and exact charset, declared successful statuses, explicit destination, atomic
  file completion, and the same response-metadata projection. A failed
  media/charset or byte check leaves no output file.
- `intent:"direct_write"` with `availability:"implemented"` can enter the
  plan → preview → approval → execute lifecycle for one declared `rest_write`
  operation. Disk-backed bundles cross-check the operation's fixed method/path
  against an `api_surface.json` operation entry. Shipped builds derive endpoint
  validation only from embedded `rest_write` declarations because
  `api_surface.json` is not embedded; that proves internal declaration
  consistency, not provider documented-surface provenance. #3773 owns the
  separate per-operation `api_surface` provenance foundation. Command preflight
  requires one connector-relative mutating endpoint declaration.
  `commandrunner` never dispatches the write directly.
- Every other command that references an `operation` returns a blocked command
  error naming the operation ID and explaining that its executor is not
  implemented. This fail-closed default is deliberate: it lets docs, validation,
  and parity planning land before any new side-effecting executor is available.

`availability: "implemented"` is a runtime claim, not a label.
`TestEveryImplementedCommandPassesRuntimePreflight` in
`internal/connectors/commandrunner/runner_test.go` sweeps every bundle in
`defs.FS` through the real `commandrunner.Preflight`, so a command cannot claim
it while the runtime blocks it. See AGENTS.md, "Command Surface Must Stay
Executable".

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
