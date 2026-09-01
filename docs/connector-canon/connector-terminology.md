# Connector terminology and lane contract

The [source.lock vNext architecture](SOURCE-LOCK-VNEXT.md) owns the schema and
pipeline. These terms keep connector authoring and runtime discussions precise.

## Authoring terms

- **Provider source:** immutable provider-published contract material inspected
  by an author. It may describe REST, GraphQL, events, database protocols,
  request/response schemas, media, and pagination.
- **Provider evidence:** authoring-only citations or facts in a vNext lock. It
  explains an authored fact but never grants runtime availability.
- **Source lock:** the connector-local schema-4 immutable authoring input.
- **Canonical descriptor:** the validated per-operation, shared-schema
  intermediate representation. It contains execution content and has no
  provider-evidence field.
- **Shared schema registry:** safe `schemas/...json` entries referenced as
  request, response, or record schemas by canonical operations.
- **Execution bundle:** rendered JSON consumed by the engine: metadata, spec,
  streams/schemas, writes, operations, CLI surface, and optional database,
  changefeed, polling, rate-limit, or sync-transport files.
- **Operation:** one canonical authoring identity that may own multiple runtime
  forms without duplicating the provider fact.
- **Command binding:** one unambiguous CLI command reference to an executable
  operation, stream, or action.

## Seven lanes

- **Direct read:** bounded interactive provider read. It is not a saved
  warehouse extraction.
- **Direct write:** bounded interactive typed mutation with invocation,
  approval, and auth checks before provider I/O.
- **Binary download:** bounded binary read with declared media/status/size and
  confined destination behavior.
- **Binary upload:** bounded typed binary or multipart mutation with declared
  fields/media/size, digest/preview, and approval semantics.
- **ETL:** saved provider-to-DuckDB extraction through a stream, schema,
  continuation contract, bounded batches, and connection-owned warehouse path.
- **Reverse ETL:** saved DuckDB-to-provider delivery through a typed action,
  plan, preview, approval when required, execution, and receipt handling.
- **Sync transport:** compatible saved source and destination executors joined
  through the warehouse with checkpoint, ordering, deletion, idempotency,
  acknowledgement, and mode semantics.

Every lane is `implemented` or `unsupported`. An operation can populate
multiple lanes, but their semantics never collapse. A direct read can also back
an ETL stream; it does not become ETL during interactive execution.

## Warehouse and replication terms

- **Stream:** named ETL record set with a source route, record path, schema,
  scope, and continuation behavior.
- **Action:** closed typed write contract for a provider mutation.
- **Scope:** provider boundary being read or written, such as account,
  workspace, project, repository, or parent resource.
- **Pagination:** documented division of a collection into pages.
- **Continuation:** documented means of reaching the next page or delta.
- **Fan-out:** bounded parent discovery followed by child requests; every
  request contributes to the same execution budget.
- **Checkpoint:** durable source position advanced only after warehouse
  acknowledgement.
- **Stable identity:** documented key or composite key used for correlation.
- **Tombstone:** explicit deletion for a known stable identity and scope.
- **Hydration:** bounded fetch of current state for an event-supported identity.
- **Warehouse receipt:** durable acknowledgement of one bounded workset.

Saved movement is always provider/database → DuckDB → provider/database. A
source and destination never communicate through a hidden direct hop.

## Write-safety terms

- **Plan:** select warehouse rows and map them to typed actions.
- **Preview:** show the exact bounded requests before execution.
- **Approval:** authorize a plan when the execution policy requires it.
- **Execute:** perform approved typed actions and record bounded outcomes.
- **Provider idempotency:** documented guarantee for safe repetition. An
  internal correlation key does not create it.
- **Single-attempt action:** no automatic replay after an ambiguous outcome.
- **Retry-safe action:** replay behavior justified by the typed provider/runtime
  contract.

## Runtime-invalid versus advisory

Runtime-invalid means malformed/missing required execution JSON, ambiguous
binding, missing actual encoder/executor, invalid bounded route/schema,
invalid invocation/approval/auth, or incompatible sync executor/mode.

Authoring diagnostics report citation gaps, unsupported lanes, or render drift.
They do not suppress a runtime-valid command.
