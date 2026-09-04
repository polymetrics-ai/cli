# source.lock vNext architecture and connector-author contract

**Status:** authoritative. This document defines the only connector authoring
pipeline and the only relationship between provider evidence and runtime
execution artifacts.

The architecture has one direction:

```text
immutable source.lock.json
        │ authoring validation
        ▼
canonical per-operation descriptors + shared schema registry
        │ deterministic rendering
        ▼
execution JSON bundle under internal/connectors/defs/<connector>/
        │ engine.Load / engine.LoadAll
        ▼
commandrunner, protocol encoders, approval/auth, warehouse, sync transport
```

`source.lock.json` is authoring and evidence input. It is never embedded in the
runtime filesystem and no runtime or diagnostic path reads it to decide whether
a command is available. Runtime reads execution JSON only. There is no second
reader, fallback, feature flag, admission ledger, evidence hash gate, or
provider-specific exception.

## 1. Source-lock document

The file is `internal/connectors/defs/<connector>/source.lock.json`. Its root is
a closed schema decoded with unknown-field rejection and currently has
`schema_version: 4`.

| Field | Contract |
| --- | --- |
| `schema_version` | Must be `4`. A different version fails authoring validation. |
| `connector` | Canonical connector name. It must match the target directory and connector-name grammar. |
| `lanes` | Exactly seven keys. Every value is `implemented` or `unsupported`; omission and extra lanes fail. |
| `provider_evidence` | Optional immutable citations and provider facts. Authoring-only: it is deliberately absent from the canonical execution descriptor. |
| `metadata` | Complete future `metadata.json` object. |
| `config_schema` | Complete future `spec.json` JSON Schema object. |
| `http` | Optional shared HTTP base, auth candidates, check route, headers, pagination, and error mapping for `streams.json`. |
| `schemas` | Shared registry keyed by safe `schemas/...json` paths. |
| `operations` | Canonical per-operation authoring units described below. |
| `cli` | Optional command-surface root fields other than operation-bound commands. |
| `execution` | Optional named execution objects: `changefeed.json`, `polling_watermark.json`, `sync_transport.json`, `rate_limits.json`, or `database.json`. No other name is accepted. |

The lock is immutable for one authored revision: review a changed lock as a new
input and render it, rather than mutating it during runtime or after credential
resolution. Provider documentation may be cited in `provider_evidence` or an
operation's `source` object. Citations do not grant execution capability and
are not copied into execution JSON.

### Canonical operation unit

Every entry in `operations` has a unique non-empty `id`. An entry may populate
multiple lanes so the provider fact is authored once while its distinct runtime
forms remain explicit.

| Field | Meaning |
| --- | --- |
| `id` | Stable authoring identity within this lock. |
| `source` | Optional operation-local provider citation/facts; authoring-only. |
| `schema_refs.request` | Shared request schema reference. |
| `schema_refs.response` | Shared response schema reference. |
| `schema_refs.record` | Shared record schema reference. |
| `stream` / `stream_order` | One ETL stream entry and its deterministic position in `streams.json`. |
| `write` / `write_order` | One typed write action and its deterministic position in `writes.json`. |
| `operation` / `operation_order` | One direct or binary execution operation and its position in `operations.json`. |
| `commands[].command` / `commands[].order` | CLI command objects and their deterministic order in `cli_surface.json`. |

At least one execution form is required: stream, write, operation, or command.
All referenced schemas must exist in the shared registry. Operation objects and
commands must be JSON objects and must not contain authoring evidence.

### Shared schema registry

Schema keys start with `schemas/`, end in `.json`, remain path-clean, and may
not traverse directories. A request, response, or record reference is valid
only when its exact key exists. Streams normally point at record schemas.
Writes use closed request schemas. Direct operations may use request and
response references appropriate to their REST, GraphQL, multipart, or binary
encoder contract.

Canonicalization rejects structurally impossible role placement before any
rendered-file replacement: a record reference requires a stream whose schema
matches it; a request reference requires a write or direct operation; and a
response reference requires a stream or direct operation. Semantic admission
then renders one in-memory execution view, loads it with the existing engine,
and uses the runtime's exact binding resolver and command preflight. It binds
each source operation to its rendered schemas, stream/write/operation and
commands; GraphQL bindings remain operation-identity based even when routes
match. Unknown provider facts remain opaque authoring data. A failed join names
the source operation and JSON field path, and occurs before output replacement.

The renderer copies each shared schema byte-deterministically to the same path
in the execution bundle. Shared references prevent command, stream, write, and
operation variants from silently drifting into incompatible shapes.

## 2. The seven lanes

Every lock declares every lane. `unsupported` is an explicit empty declaration,
not a hidden omission and not a promise to add a runtime route later.

| Lane | Execution meaning | Authored evidence of implementation |
| --- | --- | --- |
| `direct_read` | One bounded interactive provider read through a real REST, GraphQL, or other registered read encoder. | A command with `intent: direct_read` bound unambiguously to an operation/stream the commandrunner can preflight. |
| `direct_write` | One bounded typed provider mutation with invocation validation, approval where required, and auth before provider I/O. | A command with `intent: direct_write` and a real operation or action binding. |
| `binary_download` | Bounded binary provider response with declared media/status/size and destination confinement. | A command with `intent: binary_download` and a binary-download operation. |
| `binary_upload` | Bounded typed multipart/binary request with declared parts, media, size, digest/preview, and approval semantics. | A command with `intent: binary_upload` and a binary-upload operation. |
| `etl` | Saved provider-to-DuckDB extraction through a declared stream, schema, paginator, bounded batches, and warehouse ownership. | At least one authored stream. |
| `reverse_etl` | Saved DuckDB-to-provider delivery through a typed write action, plan, preview, approval, execution, and receipts. | A command with `intent: reverse_etl` plus the referenced executable action. |
| `sync_transport` | Source-to-warehouse-to-destination composition using compatible registered executors, modes, checkpoint/delivery, and acknowledgement rules. | A valid `execution["sync_transport.json"]`. |

One operation can populate more than one lane. A collection read can back both
an interactive direct-read command and an ETL stream; a mutation can back both
direct write and reverse ETL. Those consumers remain separate execution
semantics. A direct operation is never converted into a warehouse pipeline
merely because an ETL lane exists.

Lane state is checked against authored content. A claimed implemented lane with
no corresponding execution content fails, as does execution content paired
with `unsupported`.

## 3. Canonicalization and deterministic rendering

`go run ./cmd/connectorgen lock-render <connector>` performs these steps:

1. Decode the root with unknown fields rejected.
2. Validate version, connector identity, all seven lane declarations, JSON
   object shapes, unique operation IDs, schema paths/references, optional
   execution filenames, and lane/content consistency.
3. Reject evidence-only keys recursively from execution objects. In
   particular, `conformance`, `source_operation`, and `source_cli_path` cannot
   cross the authoring boundary. Provider citations stay only in authoring
   fields. A lock may record `source_cli` or `destination_cli` citations as
   authoring provenance; canonicalization removes those keys before the CLI
   surface is rendered, so changing a citation cannot change runtime bytes.
4. Clone execution-relevant content into a canonical descriptor while retaining
   raw operation-local source facts only for authoring admission.
5. Render the closed execution set in memory; call `engine.Load`, construct its
   in-memory connector, resolve every implemented command through the shared
   binding resolver and `commandrunner.Preflight`, and build the selected
   manifest-index input. A supplied complete sync plan is resolved through
   `syncplan.Resolve`; a lock never invents its missing destination or
   Foundation Atlas fact.
6. Sort streams, writes, operations, commands, and source-to-execution
   provenance deterministically.
7. Render indented JSON with a trailing newline.
8. Atomically replace destination files through a same-directory temporary
   file only after every earlier step succeeds.

The outputs are:

- `metadata.json` and `spec.json`;
- `streams.json` and registry schemas when streams or HTTP configuration exist;
- `writes.json` when write actions exist;
- `operations.json` when direct/binary operations exist;
- `cli_surface.json` when CLI root content or commands exist;
- the explicitly allowed optional execution files.

`--check` performs the same canonicalization, semantic admission, and rendering
in memory, then requires byte equality with every required output. It does not
write. A second render of unchanged input must produce identical bytes.

Authors create schema-4 locks directly from immutable provider facts. There is
no execution-JSON-to-lock importer: such a reverse path would create a second
source of truth and could silently preserve obsolete runtime fields.

## 4. Runtime boundary

`internal/connectors/defs` embeds execution JSON only. Its inventory rejects
`source.lock.json` and other evidence files. `engine.Load` and `engine.LoadAll`
parse the execution bundle and construct the existing runtime objects.

The existing runtime remains responsible for:

- command discovery and one unambiguous command-to-operation/action binding;
- invocation and bounded route/schema validation;
- REST, GraphQL, multipart, and binary encoders;
- credential resolution, auth selection, approval consumption, and the rule
  that invalid requests fail before provider I/O;
- rate-limit coordination and bounded retry behavior;
- DuckDB/Parquet materialization, connection ownership, WAL/checkpoints, and
  warehouse-mediated ETL/reverse-ETL;
- compatible sync source/destination executors and modes.

Only true runtime-invalid conditions suppress or reject execution: a malformed
or missing required execution artifact; ambiguous binding; missing actual
encoder/executor; invalid bounded route or schema; invalid invocation,
approval, or auth; and incompatible sync executor/mode. Authoring citations,
content hashes, review records, or absence of external proof do not suppress a
documented command whose execution contract is otherwise valid.

Diagnostics may read the vNext lock offline and report missing citations,
unsupported lanes, or render drift. Such a report is advisory and non-binding.
It must not load any predecessor-format source descriptor and must not feed a
runtime decision.

## 5. Error handling

Authoring errors include unsupported schema version, target-name mismatch,
unknown root fields, missing/extra/invalid lanes, lane/content contradiction,
duplicate/empty operation IDs, operation with no execution form, invalid JSON
object, unsafe/missing schema reference, unsupported optional artifact, and
evidence leakage into execution content. Semantic admission additionally
rejects a missing or cross-source schema/stream/write/operation/command join,
an invalid runtime route or encoder, a mismatched staged identity or executor,
and malformed `rate_limits.json`; each fails `lock-render` before any file
replacement.

Runtime errors remain typed at the closest execution boundary. The authoring
admission uses the same in-memory loader, exact binding resolver, and command
preflight without provider I/O, credential resolution, transport, staging, or
activation. Missing or ambiguous bindings fail command preflight. Unsupported
encoders/executors fail before provider I/O. Invocation, approval, and
credential failures retain their existing typed boundaries. Warehouse and sync
incompatibilities fail planning/preflight rather than being reclassified as an
authoring-evidence failure.

A missing genuine shared runtime capability is recorded as a concrete
foundation gap. Keep the source fact in the lock, declare the affected lane
`unsupported`, and do not add a connector-specific bypass or new shared
foundation without its own approved design.

## 6. Connector-author workflow

1. Inspect immutable provider material without fetching mutable documentation
   during a deterministic build. Record truthful citations in authoring-only
   evidence fields.
2. Create or update one schema-4 `source.lock.json`. Declare all seven lanes
   and preserve every actual provider fact needed by an authored request,
   response, media, pagination, or sync contract.
3. Model each provider operation once. Attach all applicable stream, write,
   direct/binary operation, command, and shared schema references to that unit.
4. Use only `implemented` when the authored content has a registered real
   runtime path. Use `unsupported` for an honest empty lane or a concrete
   missing shared foundation.
5. Render and check:

   ```bash
   go run ./cmd/connectorgen lock-render <connector>
   go run ./cmd/connectorgen lock-render <connector> --check
   go run ./cmd/connectorgen validate internal/connectors/defs
   ```

6. Review the diff in both the lock and rendered JSON. Execution changes must
   be explainable by the lock; provider evidence must not appear in rendered
   artifacts.
7. Prove discovery without credentials or provider I/O, malformed-artifact
   rejection, lane reporting, binding preflight, and each configured runtime
   executor. Use a fake server for wire-shape/credential-boundary tests and
   DuckDB for saved-flow proofs.
8. Run connector-local tests, renderer/check tests, engine/commandrunner tests,
   CLI proofs, and broader affected-package tests. Commit and push only when
   all required checks are green.

## 7. Proof requirements

A reference connector is green only when all of the following are demonstrated:

- `lock-render --check` is byte-stable;
- runtime inventory contains execution JSON and excludes authoring evidence;
- `engine.Load` discovers the connector without consulting a lock or external
  evidence;
- a command reaches the credential/approval boundary with provider I/O
  disabled or a local fake server;
- malformed required execution JSON is rejected;
- each of the seven lanes is accurately `implemented` or `unsupported`;
- configured direct read/write and binary encoders are exercised;
- ETL and reverse ETL traverse DuckDB rather than a direct provider hop;
- sync transport is exposed only when its source/destination executor and mode
  contracts are compatible;
- focused Go tests, affected broader suites, the CLI build, and generated-output
  checks are green.

Live provider credentials are not required for deterministic authoring or
runtime reachability proofs. If live proof is separately authorized, it adds
operational confidence but never becomes runtime admission state.
