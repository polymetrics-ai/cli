# Connector terminology and lane contract

This document is the single owner of the vocabulary used to build, map, validate, and prove API connectors.
Read it before changing a connector definition, source lock, lane mapping, certification rule, or missing-foundation record.
The construction and proof procedure lives in the [Implementation Procedure](IMPLEMENTATION-PROCEDURE.md).

## Authority and boundaries

A connector is a source-backed provider integration with an executable PM-CLI definition.
The provider source defines what the provider documents.
The source lock records provider facts and citations.
The mapping declares how those facts occupy connector lanes.
The projected connector JSON drives execution.
Certification verifies an already-mapped executable path.
Certification, importer behavior, a hash, a credential, or a destructive HTTP method must never erase a provider operation from the source denominator.

The runtime reads projected connector JSON rather than raw documentation or the source lock.
The source lock remains the authority for every provider fact used by a projected declaration.
System policy such as approval, retry behavior, warehouse mediation, and scheduler behavior belongs in explicit runtime policy rather than in provider provenance.

## Source facts and identities

### Provider source

A provider source is the provider-published material that defines an API contract.
It can be an OpenAPI or Swagger document, GraphQL schema, discovery document, or retained rendered reference page.
A documentation index, login wall, redirect response, or error page is not a provider source.

### Retained artifact

A retained artifact is the exact provider bytes used as evidence for a source lock.
A digest pins those bytes to a particular source-lock revision.
A digest proves provenance and change detection, not request executability.

### Citation

A citation identifies the exact provider artifact and operation, request field, response field, media type, pagination rule, or lifecycle fact that supports a mapping.
Every declared provider-facing fact needs a citation.

### Source lock

A source lock is the frozen provider-fact authority for one connector and one source revision.
It contains or cites source operations, request and response facts, media details, pagination and event evidence, and explicit provider-information gaps.
A source lock revision can be re-pinned when the provider changes its documentation, but a mapping always names the exact revision it used.
The source lock does not contain PM-specific execution policy.

### Source descriptor and importer

A source descriptor is a normalized, source-backed view of the retained source that makes operations and their documented facts easier to map.
An importer reads a retained source and produces or validates that descriptor.
An importer is a source-reading tool, not a policy authority.
An importer failure must retain the source operation with a named parsing or projection gap instead of silently dropping it.

### Descriptor-free source-accounting retention

`retention_only` is a narrow mapping-admission contract for a connector whose
immutable source lock and exact source-lane accounting are retained but whose
historical canonical descriptor is intentionally absent. It is valid only when
the primary lock's digest/byte identity and exact source-operation-ID
reconciliation pass, every lane remains nonimplemented, and each lane cites
only the source-lock artifact. It does not create a command, stream, write,
transport, certification result, or executable connector.

Source operation IDs are opaque provider evidence, not filenames. Preserve
ordinary spaces and slash-bearing provider spellings exactly; reject only empty
or control-character IDs, and never normalize an ID to make it path-safe.

### Source operation and `source_operation_id`

A source operation is one provider API identity, such as `GET /tasks/{task_gid}` or `POST /projects`.
`source_operation_id` is the stable provenance identifier for that provider operation.
The word `source` is intentional because the identifier answers which provider contract fact PM is implementing.

### Provider operation ID

A provider operation ID is the provider's own operation name when the source publishes one.
It can be used alongside the source operation ID, but it does not replace the canonical source identity when the provider has duplicate, missing, or unstable names.

### Connector operation and `connector_operation_id`

A connector operation is PM's typed executable declaration derived from provider facts.
`connector_operation_id` identifies the declaration inside the connector definition.
One source operation can produce more than one connector operation when PM offers aliases or distinct typed representations.
One connector operation must retain its exact source-operation binding.

### Source-operation binding

A source-operation binding is the explicit link from a projected connector declaration back to the exact `source_operation_id`, method, path, and citation that justify it.
It is required for read, write, binary, ETL, reverse-ETL, and transport mappings.
Binding a write or binary operation is provenance, not permission to issue arbitrary HTTP.

### Lane cell

A lane cell is one pair of a source operation and a connector lane.
One source operation may occupy multiple lane cells.
For example, a paginated `GET /tasks` can occupy both `direct_read` and `etl`.
A `POST /tasks` can occupy both `direct_write` and `reverse_etl`.
This is why the number of lane cells can exceed the number of source operations.

### Mapping manifest

A mapping manifest is the structured, source-lock-bound record that assigns source operations to lane cells, connector operations, projected artifacts, and named gaps.
It is derived from source facts and binds the source-lock digest or revision it uses.
Its file name may vary with the current connector-definition schema, but its meaning must remain stable.

### Projection and materialization

Projection is the deterministic transformation from source-backed mappings into executable connector-definition JSON.
Materialization is the creation or update of the resulting JSON artifacts.
Projection must not invent a request field, response field, schema, media type, stable key, or endpoint that the source does not document.

## The seven connector lanes

Every enabled connector declares all seven lanes.
A lane can be provider-evidenced unsupported or blocked by a named runtime foundation, but it cannot disappear silently.

### `direct_read`

A direct read is one bounded provider read that returns provider output.
The command must bound the number of provider requests, pages, and fan-out children it can issue.
A direct read can return one page of a collection.
A direct read is not ETL merely because the underlying endpoint supports pagination.

### `direct_write`

A direct write is one exact typed provider mutation for one input record.
It constructs a bounded request from documented fields and returns the provider response.
Delete, remove, archive, and destructive provider operations remain visible when the provider documents them.
Their source status is not changed by whether an operator later approves execution.

### `binary_download`

A binary download is the bounded binary form of a direct read.
It requires a source-backed media contract and a bounded request and response shape.

### `binary_upload`

A binary upload is the typed binary form of a direct write.
It requires documented request media, fields, and provider response behavior.

### `etl`

ETL is a saved provider-to-DuckDB extraction pipeline.
It starts from a source-backed read mapping, exhausts the declared source scope, and durably materializes records in DuckDB.
ETL is not an interactive direct-read command and is not a connector-to-connector API shortcut.

A paginated collection is an ETL candidate when the source documents the collection records, a continuation mechanism, and the required scope or parent inputs.
An ETL candidate still needs a typed stream mapping, record path, schema, continuation semantics, and executable warehouse proof before it is called implemented.
Every documented paginated collection must be classified explicitly as ETL, direct-only, fan-out-required, blocked with a named foundation, or unsupported with provider evidence.

### `reverse_etl`

Reverse ETL is a saved DuckDB-to-provider pipeline.
It reads warehouse rows, maps one row to one bounded typed provider write action, previews the exact requests, consumes approval when required, and dispatches the action.
Reverse ETL can reuse the same provider mutation as direct write without being the same user interaction.

### `sync_transport`

Sync transport composes saved sides as source to DuckDB to destination.
It never creates a hidden provider-to-provider API shortcut.
A source transport owns source scope, continuation, checkpoints, ordering, deletion semantics, and durable warehouse acknowledgement.
A managed destination transport owns eligible reverse-ETL actions, acknowledgement, partial-failure behavior, ordering, idempotency, retry, and ambiguous-result behavior.

## Warehouse and replication terms

### Stream

A stream is a named ETL record set with a declared source operation, scope, schema, record path, and extraction behavior.

### Scope

Scope is the documented provider boundary being extracted or written, such as a workspace, project, group, account, or parent resource.

### Pagination and continuation

Pagination divides a collection into provider pages.
Continuation is the documented mechanism for reaching the next page, such as an offset, cursor, `next_page`, or `has_more` response signal.
A limit parameter alone does not prove that PM can exhaust the collection.

### Fan-out

Fan-out is a source pattern in which one parent collection supplies inputs for child requests.
It is not automatically safe or unbounded by default.
The mapping must state the parent scope, child request, request bound, and execution semantics.

### Full refresh

A full refresh exhausts one declared source scope before reporting success.
Full-refresh overwrite replaces destination state only after the complete read succeeds.
Full-refresh append adds another complete snapshot.

### Incremental ETL

Incremental ETL reads only a documented provider delta after a durable checkpoint.
It requires a provider-documented cursor, event token, watermark, or equivalent continuation semantics.
Local content hashes, timestamps inferred from output, and page ordinals do not create an incremental cursor.

### Incremental dedupe with history

Incremental dedupe with history requires documented stable identity and sufficient provider ordering, version, or cursor semantics to preserve the claimed history behavior.
It must not be inferred from a changing content hash.

### Event source and event token

An event source is a provider-documented change feed.
An event token is the documented durable position used to resume that feed.
Event-token support requires evidence that the subscription scope and emitted resource types cover the mapped stream.

### Checkpoint

A checkpoint is the durable position PM commits only after DuckDB acknowledges the associated source window.
It is a replication state value, not a direct-command parameter.

### Stable provider identity

A stable provider identity is a provider-documented primary key or documented composite identity.
It supports record correlation and, when source evidence permits, deduplication and tombstone application.

### Tombstone

A tombstone records a provider-documented deletion of a resource with known stable identity and deletion scope.
`removed` describes a relationship or scope change and is not automatically a resource tombstone.

### Hydration

Hydration fetches the current provider representation for an event-supported resource when the documented event contract requires it.
A hydration endpoint alone does not prove event coverage or incremental eligibility.

## Write reliability terms

### Approval

Approval is an explicit runtime execution-policy control that is consumed before provider I/O when a mode requires it.
Approval does not change source membership or mapping eligibility.

### Idempotency

Provider idempotency is a provider-documented guarantee that repeating a request has the documented safe result.
An internal PM correlation key is not provider idempotency.
PM must not invent an idempotency header, replay guarantee, or retry-safe claim.

### Single-attempt action

A single-attempt action maps one row to one bounded request, disables automatic replay, reports partial failure, and requires operator review after an ambiguous result.
Lack of provider idempotency does not by itself prevent the first approved write.

### Retry-safe action

A retry-safe action has source-backed or otherwise proven semantics that make the configured replay policy safe.
It must state the provider evidence and exact retry behavior.

### Plan, preview, approval, and execute

Plan selects and maps warehouse rows to typed actions.
Preview shows the exact bounded requests that would be sent.
Approval authorizes the selected plan when the execution policy requires it.
Execute performs the approved actions and records bounded provider outcomes.

## Foundations and proof

### Foundation

A foundation is reusable runtime code that enables a class of typed connector mappings to execute.
It is distinct from source retention, parsing, projection, certification, and documentation.

### Mapping restriction

A mapping restriction is an importer, schema, admission, or certification rule that rejects or hides a source-backed mapping without representing a genuine execution limitation.
Mapping restrictions must be repaired so the source operation remains mapped with its real state.

### Runtime foundation gap

A runtime foundation gap exists only when no existing executor, extension seam, or connector-specific hook can execute a source-backed mapping safely and truthfully.
A genuine new shared runtime foundation requires captain approval before implementation.

### Connector-specific hook or extension seam

A connector-specific hook or extension seam contains behavior that is genuinely provider-specific and cannot be expressed through the shared typed contract.
It must retain its source evidence and must not be used to bypass a reusable foundation or admit raw request/response execution.

### Foundation Atlas

The Foundation Atlas is the CLI-owned catalog of reusable foundations, their owner symbols, constraints, supported contracts, and proof tests.
It is an authoring and discovery aid, not a runtime loader, mapping-admission gate, or certification-admission gate.
`proof_tests` are references to evidence that a foundation works, not a requirement that a connector already be complete.

### `missing-foundation.json`

`missing-foundation.json` records a precise source-cited runtime gap after the Atlas has been checked.
Each entry must identify the affected source operation or lane cell, the missing contract, the Atlas classification, and the exact evidence needed to close it.
It must not be used as a generic deferred bucket.

### Certification

Certification is a credential-bound verification overlay for an exact source-lock revision, projected definition revision, executor, scope, resource type, and mode.
It proves that a reachable command reaches the credential boundary and behaves as expected at that boundary.
Certification cannot create a provider fact, configure runtime behavior, decide source membership, or turn an unmapped declaration into an implementation.

### Credential-boundary proof

A credential-boundary proof resolves the real CLI or App command, constructs the typed request, reaches the credential boundary, and records the observed result without relying on a JSON parse alone.
For saved ETL and reverse ETL, the proof also covers isolated DuckDB planning, preview, execution, checkpoint, and bounded request-count behavior.

## Connector-definition artifacts

The exact schema may evolve, but a complete API connector definition must retain the following categories of artifacts.

| Artifact category | Typical current files | Purpose |
| --- | --- | --- |
| Source retention | `sources/<connector>-operation-source-lock.json`, retained artifacts, source descriptor | Preserve provider facts and citations. |
| Core definition | `metadata.json`, `spec.json`, `api_surface.json`, `cli_surface.json` | Identify the connector and expose typed surfaces. |
| Direct reads | `operations.json` and source-operation bindings | Declare bounded provider reads. |
| Direct writes and reverse ETL | `writes.json` and source-operation bindings | Declare bounded provider mutations. |
| ETL | `streams.json` and stream schemas | Declare provider-to-DuckDB extraction. |
| Binary lanes | Source-backed media records plus operations or writes declarations | Declare bounded binary transfers. |
| Saved transport | `sync_transport.json` and, when applicable, `event_source_contract.json` | Declare warehouse-mediated source and destination behavior. |
| Completion and gaps | `enabled_connector_contract.json` and `missing-foundation.json` | Reconcile all seven lanes and named runtime gaps. |

An artifact may be empty only when its lane declaration explicitly carries a provider citation and a truthful unsupported or named-gap state.
An enabled connector must not omit a lane merely because no implementation has been written yet.

## States and completion

### Source-accounting state

Every locked source operation must be accounted for exactly once as `implemented`, `blocked_with_named_foundation`, or `unsupported_with_provider_evidence`.
Aliases are counted separately from provider operations and cannot inflate source coverage.

### Mapping-completeness state

Mapping completeness records whether every source operation and lane cell has a source-operation binding and projected artifact.
`unmapped_mapping` is a mapping deficit, not evidence that the provider operation is absent.

### Runtime-execution state

Runtime execution records whether the typed executor and saved-connection path can perform the mapped lane.
An implemented runtime lane can still have partial source coverage, and that partial coverage must remain visible.

### Complete connector parity

Complete connector parity means every locked source operation is accounted for, every applicable lane cell is mapped or truthfully named as a gap, all seven lane declarations are present, and the executable lanes have the appropriate direct or DuckDB proof.
It does not mean that every provider API is safe for unattended automation.
It does mean that provider-documented delete, reverse-ETL, and other destructive surfaces remain correctly represented.

## Required build sequence

1. Retain and validate the provider source.
2. Freeze the source-lock denominator and normalize a source descriptor. A
   `retention_only` contract is the narrow non-executable exception: it keeps
   exact-ID source accounting without a historical descriptor only after lock
   identity and full reconciliation pass.
3. Classify each source operation into one or more lane cells or a cited unsupported state.
4. Create source-operation bindings and project the required connector-definition artifacts.
5. Consult the Foundation Atlas before recording any missing foundation.
6. Prove bounded direct commands, provider-to-DuckDB ETL, and DuckDB-to-provider reverse ETL through the correct execution path.
7. Run the seven-lane enabled-connector contract validation and credential-boundary proof.
8. Record certification as evidence after mapping and execution, never as a prerequisite for them.

## Non-negotiable no-miss rules

- Do not classify all `GET` endpoints as ETL, because a singleton read is not an exhaustive collection extraction.
- Do not classify all mutations as reverse ETL, because each warehouse row still needs a documented one-to-one bounded write mapping.
- Do not omit a documented paginated collection, because it must receive an explicit ETL classification or a cited reason.
- Do not use a content hash as a provider cursor, primary key, or idempotency guarantee.
- Do not let importer, certification, hash, credential, or destructive-method restrictions hide a source operation.
- Do not create generic raw JSON or arbitrary HTTP escape hatches to work around a missing typed foundation.
- Do not call a runtime foundation missing before checking the Foundation Atlas and its owner code.
- Do not implement a genuinely new runtime foundation without captain approval.
