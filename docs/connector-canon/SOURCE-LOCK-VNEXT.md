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
closed published execution generation under
internal/connectors/defs/<connector>/generations/<generation>/
        │ selected through atomic CURRENT when publication is enabled
        ▼
existing execution-only runtime boundary; commandrunner, protocol encoders, approval/auth, warehouse, sync transport
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
| `schema_refs.request` | Shared request schema reference; semantic admission binds it to the effective loaded write or direct-operation request schema. |
| `schema_refs.response` | Shared response schema reference; semantic admission binds it to the effective loaded stream response schema. |
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
Writes use closed request schemas. A request reference must semantically equal
the loaded write `record_schema` or a rendered REST `body_schema` / GraphQL
`variables_schema`. A response reference must semantically equal the loaded
stream schema. The current direct-operation contract has no typed response
schema, so a direct-only response role is rejected rather than admitted as
provenance-only.

Canonicalization rejects structurally impossible role placement before candidate
staging: a record reference requires a stream whose schema matches it; a request
reference requires a write or direct operation; and a response reference
requires a stream or direct operation. Semantic admission then renders one
in-memory execution view, loads it with the existing engine, and proves every
request/response role against the effective runtime schema before binding each
source operation to its rendered stream/write/operation and commands. It
constructs the in-memory connector through the closed native executor and
generated hook authorities, then binds the complete staged manifest/index entry.
GraphQL bindings remain operation-identity based even when routes match.
Unknown provider facts remain opaque authoring data. A failed join names the
source operation and JSON field path, and occurs before candidate staging or
activation.

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

## 3. Canonicalization, staged validation, and publication

`go run ./cmd/connectorgen lock-render <connector>` performs these steps:

1. Open the definitions root and target connector through no-follow descriptors,
   retain the connector descriptor, and decode `source.lock.json` through it
   with unknown fields rejected.
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
5. Render the closed execution set in memory; call `engine.Load`, select the
   closed native executor and generated hook extension, construct its in-memory
   connector, resolve every implemented command through the shared binding
   resolver and `commandrunner.Preflight`, and build the complete selected
   manifest-index input. A supplied complete sync plan is resolved through
   `syncplan.Resolve`; a lock never invents its missing destination or
   Foundation Atlas fact.
6. Sort streams, writes, operations, commands, and source-to-execution
   provenance deterministically. Canonical source-ID order owns staged
   provenance; authored operation-array positions are retained only for
   diagnostics.
7. Render indented JSON with a trailing newline.
8. Build one closed publication candidate containing every rendered execution
   file plus `manifest.json`, `provenance.json`, `atlas.json`, `index.json`,
   and `proof.json`. The author-owned `source.lock.json` is never a candidate
   member.
9. After the entire candidate has been admitted in memory, acquire only the
   target connector's publication lock through that same retained connector
   descriptor, then reread the source lock through it under the lock. The lock
   pathname and its `.connectorgen.lock.anchor` companion must name the same
   verified inode. The companion is linked only when a new lock is created; a
   missing companion for an existing lock refuses rather than rebuilding a
   possibly separate lock domain. Keep the acquired lock descriptor and verify
   both names after nonblocking acquisition and before each publication state
   transition; replacing the lock pathname therefore refuses instead of
   creating a second lock domain. If the source bytes changed during admission,
   refuse the operation and require a retry. Only then open or create the
   `generations/` directory and retain its verified descriptor for the rest of
   the operation. Read lock/control files, stages, leases, and every cleanup
   target through these descriptors only.
   After a successful nonblocking lock acquisition, recheck the signal context
   and unlock before any mutation when it was cancelled. Recover any interrupted
   prior publication from its typed journal before staging a new candidate.
   Publication never spans connectors.
10. Create a same-filesystem temporary directory directly below that connector's
    `generations/` directory. First write and fsync a typed stage-ownership
    marker binding its version, connector, content generation, and stage name.
    Recovery removes only a marker-proven stage; an unknown or malformed
    `.stage-*` directory is preserved and refuses the operation.
11. Write the entire candidate, an empty per-generation lease, and its
    file-digest `integrity.json`; reject unsafe paths and symlinks, then fsync
    every file and staged directory.
12. Validate the physical staged files through the same in-memory loader,
    selected runtime identity, manifest/index construction, and
    `commandrunner.Preflight`. Each validation recomputes the framed
    content-generation address from the closed artifact bytes, requires the
    exact declared directory/member set and empty lease, and performs no
    credential, provider, or transport I/O.
13. Persist and fsync a typed `state:"prepared"` recovery journal
    `{old,new,state}` **before** the same-directory rename that activates the
    completed stage as `generations/<content-digest>/`; no final-generation
    rename is allowed before that journal is durable. Temporary `CURRENT` and
    `JOURNAL` files keep their verified open-file identity through the final
    atomic rename; a swapped temporary refuses without altering the prior
    control. Fsync the `generations/` descriptor after the rename, then
    atomically replace `CURRENT` with the new generation and integrity digest
    and fsync its containing descriptor. `CURRENT`, the journal,
    `integrity.json` (including each file entry), and the stage marker are
    bounded, no-follow, strict JSON documents: duplicate members and trailing
    values are invalid.
14. Revalidate the selected `CURRENT` generation, including its recomputed
    content address, mark the journal committed, and clear it only after the
    active generation is complete. A failed active validation restores the old
    `CURRENT` (or removes it for a first publish) and removes the rejected
    generation. The validated `CURRENT` or `JOURNAL` descriptor remains bound
    through cleanup; a replacement refuses and preserves both objects.
15. Prune only stale generations whose closed tree and integrity prove publisher
    ownership and whose per-generation lease can be acquired exclusively. A
    validated stage, generation, and lease keep their descriptor identities
    through recovery, pruning, and rollback removal; a replacement refuses
    rather than deleting either the original or replacement. A reader holds
    that lease from reading `CURRENT` until it releases its handle, so an
    in-use old generation—and any unowned directory—remains intact.

The `CURRENT` pointer is the only generation selection authority for a
publication root. A reader observes either the prior complete generation or the
new complete generation—never an index from one generation with an artifact
from another. The prepared journal and typed stage marker form one recovery
state machine: a crash before final rename removes the owned stage and restores
the old selection; a crash after final rename but before `CURRENT` restores the
old selection and removes the new generation; a valid `CURRENT` pointing at the
new generation completes recovery. Recovery treats an interrupted stage as
stale only after its durable ownership proof; it never accepts an incomplete,
renamed, self-consistent, symlinked, or otherwise non-exact tree. Optional files
are members of the closed set: a generation that retains an optional file absent
from the candidate fails validation rather than silently inheriting it.

The outputs inside a published generation are:

- `metadata.json` and `spec.json`;
- `streams.json` and registry schemas when streams or HTTP configuration exist;
- `writes.json` when write actions exist;
- `operations.json` when direct/binary operations exist;
- `cli_surface.json` when CLI root content or commands exist;
- the explicitly allowed optional execution files;
- publication metadata: `manifest.json`, `provenance.json`, `atlas.json`,
  `index.json`, and `proof.json`;
- publication control members: `integrity.json`, `.lease`, and the durable
  `.connectorgen-stage.json` ownership marker; none is a runtime artifact.

`--check` performs the same canonicalization, semantic admission, rendering,
closed-set construction, and staged-file validation in memory. It then requires
the selected `CURRENT` generation to contain exactly those bytes and integrity
metadata; it never writes. A fresh reference corpus that intentionally has no
published generation is proved by the in-memory deterministic renderer test,
not by a flat-file fallback reader.

Authors create schema-4 locks directly from immutable provider facts. There is
no execution-JSON-to-lock importer: such a reverse path would create a second
source of truth and could silently preserve obsolete runtime fields.

## 4. Runtime boundary

`internal/connectors/defs` embeds execution JSON only. Its inventory rejects
`source.lock.json` and other evidence files. `engine.Load` and `engine.LoadAll`
parse the execution bundle and construct the existing runtime objects.

Generation publication is an authoring-only foundation in this checkpoint. It
does not materialize the checked-in connector corpus, alter `defs.FS`, or add a
runtime reader for `CURRENT`, the journal, leases, integrity metadata, or a
source lock. A later runtime adoption may consume a selected generation only
through `CURRENT`; it requires its own approved change and cannot introduce a
flat-file fallback.

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
rejects an existing-but-swapped request/response schema, a request/response
role with no effective runtime schema, a missing or cross-source
stream/write/operation/command join, an invalid runtime route or encoder, an
incomplete staged executor/extension identity, a mismatched staged identity,
and malformed `rate_limits.json`; each fails before candidate staging or
activation. Publication additionally rejects unsafe or reserved artifact paths,
symlinks, an incomplete/digest-mismatched selected tree, an unexpected stale
member, a malformed pointer or journal, and an unavailable stale-generation
lease. A failed publication preserves or restores the prior complete
generation.

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

6. Review the lock and the selected closed generation. Execution and
   publication-metadata changes must be explainable by the lock; provider
   evidence must not appear in any generation member. Confirm `--check` does
   not write and observes the exact selected generation.
7. Prove discovery without credentials or provider I/O, malformed-artifact
   rejection, lane reporting, binding preflight, and each configured runtime
   executor. Use a fake server for wire-shape/credential-boundary tests and
   DuckDB for saved-flow proofs.
8. Run connector-local tests, renderer/check tests, engine/commandrunner tests,
   CLI proofs, and broader affected-package tests. Commit and push only when
   all required checks are green.

## 7. Proof requirements

A reference connector is green only when all of the following are demonstrated:

- a published connector's `lock-render --check` is byte-stable and read-only;
- a deliberately unmaterialized reference corpus has in-memory renderer parity
  proof rather than a flat-file fallback;
- fault recovery yields the old complete or new complete generation only;
- a held reader lease prevents stale-generation pruning, and concurrent readers
  observe one complete selected generation;
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
