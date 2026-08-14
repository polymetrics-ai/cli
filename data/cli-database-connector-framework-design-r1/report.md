# Centralized database connector framework for `pm`

**Design report r1 — 2026-08-10**<br>
**Repository:** `polymetrics-ai/cli`<br>
**Inspected commit:** `0600dfefba4fc5c1c274755c8a8c6a5e1221a9bb`<br>
**CDC follow-up inspected:** `origin/fm/cli-postgres-cdc-logical-replication-r1` at
`aeafb4ff0f403cff5aa4498265c177e57bf654fd` (13 behind / 2 ahead of the inspected `main`)<br>
**Scope:** design and implementation plan only; no production code was changed

## Executive recommendation

Build a typed database framework in `internal/connectors/database/`, make PostgreSQL its reference
driver, and attach it to the native execution admission already present in `internal/synccontract`.
Do **not** create a generic SQL executor, a database-shaped REST operation, a second sync-mode enum,
or a separate repository.

The framework should own all repeatable policy:

- normalized catalog and stable keyset read planning;
- a lossless logical type system and bidirectional mapping validation;
- immutable target/write plans, preview digests, drift detection, and destructive classification;
- structurally connection-scoped managed-table addresses, in-database ownership assertions, and a
  versioned Polymetrics system-column schema;
- closed dialect rendering from a typed statement model;
- managed-table creation, staging, key validation, mode application, transaction commit/abort, and
  confirmed-commit durability receipts;
- native-executor registration and a reusable conformance corpus.

Each database should supply only what cannot be configuration-driven:

- a wire/session driver;
- native metadata introspection and row decoding;
- transaction execution and its safest bulk-load primitive;
- a strict `database.json` declaring closed dialect, type, and capability choices;
- a database-specific changefeed adapter when applicable, admitted through the shared transaction
  and checkpoint contract.

PostgreSQL-to-PostgreSQL should initially use a durable, warehouse-mediated path:

```text
PostgreSQL source
  → bounded stable-keyset read
  → connection-owned JSONL WAL (fsync)
  → atomically published Parquet table + immutable input identity
  → reverse-ETL plan
  → preview
  → approval
  → PostgreSQL managed-table session
      (assert/create owner → stage → validate → apply → COMMIT)
  → durability receipt
  → delivery checkpoint
```

This deliberately spends local disk and a second I/O pass to gain replay, bounded memory,
independent extraction/load retry, immutable preview input, and an unambiguous durability boundary.
A future low-latency path may omit Parquet, but it may not omit a durable replay spool or advance the
source before the target commit is known durable.

Phase one should deliver the shared write foundation, PostgreSQL write, and a live
PostgreSQL-source-to-PostgreSQL-destination vertical slice through that path. Its destination is
always a **Polymetrics-created, Polymetrics-owned managed table**; phase one neither adopts nor
mutates an arbitrary pre-existing customer table. PostgreSQL CDC is no longer a dependency decision:
the captain approved `pglogrepl`, and the active branch already contains a substantial logical-
replication source. Preserve that protocol/lifecycle core, harden the transaction/event boundary
described in the correction addendum, and use it as an existing source adapter rather than deferring
or reimplementing it.

## Bottom line from the current code

The brief's structural diagnosis is correct:

- PostgreSQL `Write` is an unsupported stub at
  `internal/connectors/native/postgres/connector.go:97-101`.
- MySQL is identical at `internal/connectors/native/mysql/connector.go:43-46`.
- PostgreSQL and MySQL both declare `capabilities.write: false` at
  `internal/connectors/defs/postgres/metadata.json:7-18` and
  `internal/connectors/defs/mysql/metadata.json:8-22`.
- The declarative operation-kind enum has REST/GraphQL/XML/file operations but no database write at
  `internal/connectors/engine/schema/operations.schema.json:19-36`.
- A current bundle scan found 240 connectors with write actions: **239 API connectors and one queue
  connector; zero databases**. This slightly corrects the literal “every one is an API” statement,
  but not its conclusion.
- PostgreSQL and MySQL are separately registered native packages at
  `internal/connectors/native/nativeset/factories.go:20-29`, after declarative bundles are loaded at
  `internal/connectors/bundleregistry/registry.go:15-28`.

There is, however, a strong foundation that the new work must reuse rather than replace:

- `internal/synccontract/mode.go:8-31` defines the product's seven closed modes.
- `internal/synccontract/native.go:15-120` defines concrete native executor references, descriptors,
  conformance admission, and explicitly excludes generic query fields.
- `internal/synccontract/commit.go:11-31,51-96` prevents checkpoint commit without a durable
  downstream acknowledgement.
- `internal/app/app.go:1038-1050` refuses a contract mode until a matching native executor has
  completed conformance.
- `internal/app/app.go:1368-1483`, `1734-1806`, and `2106-2181` already implement the reverse-ETL
  plan, preview, approval, input revalidation, approval consumption, and execution sequence.

The missing unit is therefore not another top-level command or mode contract. It is a typed
**database write executor/session** that can make several bounded batches one atomic target
operation and return durability evidence coupled to the actual commit.

One existing contract needs a deliberate refinement. `NativeSyncExecutor` combines admission
(descriptor + conformance) with `RunNativeSync`, but its request carries only mode/resume/checkpoint
state (`internal/synccontract/native.go:42-71`), not a sealed target plan or batch source. Do not hide
those behind a generic map or opaque SQL payload. Extract the common admission half and let source
sync and table write expose separate typed execution methods, as detailed below.

## Design principles

1. **One semantic owner.** `internal/synccontract` remains the sole owner of sync modes, checkpoint
   identity, resume/rebootstrap, delete/history semantics, and acknowledgement order.
2. **One irreversible-write entrance.** A database write is a reverse-ETL execution and must pass
   plan → preview → approval → execute. No direct `pm postgres sql` write path is introduced.
3. **Configuration selects code; it does not contain code.** Dialect configuration may select a
   closed renderer primitive. It may not contain SQL snippets, templates, functions, or raw commands.
4. **Exact capability claims.** `capabilities.write` stays false until the actual runtime executor,
   its admitted modes, shared conformance evidence, and live tests all pass.
5. **Fail closed on identity, type, order, and durability.** A missing destination identity,
   ambiguous cursor order, unmappable type, target drift, absent key, or unknown commit result is an
   error. None has a fallback.
6. **Bounded data, explicit lifetime.** Reads and staging use bounded batches, every blocking call
   accepts `context.Context`, and connections/rows/transactions have deterministic closure.
7. **Framework policy, driver mechanics.** Mode behavior and safety live once in the framework;
   drivers provide protocol and database mechanics, not bespoke interpretations of append/upsert.
8. **No false “configuration-only” promise.** MariaDB may reuse MySQL's protocol driver. SQL Server
   and Oracle still need new protocol drivers. What disappears is a new hand-written connector and
   policy stack for every engine.

## Recommended component architecture

### Package and file layout

```text
internal/connectors/database/
  definition.go          DatabaseDefinition, CapabilitySet, strict validation
  logical_type.go        closed parameterized LogicalType algebra
  type_plan.go           source/target TypePlan compilation and value guards
  identifiers.go         structured CatalogRef/SchemaRef/TableRef/ColumnRef
  ownership.go           TargetOwner, ManagedTargetRef, fail-closed identity checks
  managed_columns.go     ManagedRecord/ManagedRowBatch + versioned reserved-column contract
  provisioning.go       managed namespace/table/control-registry state machine
  catalog.go             normalized catalog and key/constraint representation
  read_plan.go           snapshot + deterministic keyset ReadPlan
  reader.go              bounded framework read orchestration
  write_plan.go          immutable DatabaseWritePlan and target/input fingerprints
  write_session.go       WriteSession, DurabilityReceipt, commit outcome handling
  writer.go              framework-owned mode/stage/validate/apply orchestration
  statement.go           opaque compiled statement model; no caller SQL
  render.go              closed renderer primitives chosen by dialect config
  registry.go            driver + definition + admitted write-executor registry
  errors.go              typed mapping/drift/key/order/commit/rebootstrap failures
  conformance/
    definition.go        static capability/type/renderer tests
    read.go              catalog, bounds, composite order, cancellation fixtures
    write.go             modes, rollback, durability, drift, context fixtures
  drivers/postgres/
    driver.go            pgx connection/transaction adapter
    introspect.go        PostgreSQL native metadata → normalized catalog
    ownership.go         pg_namespace/pg_class identity and managed registry adapter
    provision.go         transactional managed schema/table/index creation
    bulk.go              PostgreSQL bulk-load primitive
  drivers/mysql/         phase-three protocol adapter

internal/connectors/database/schema/database.schema.json
internal/connectors/defs/postgres/database.json
internal/connectors/defs/mysql/database.json

internal/connectors/native/postgres/
  connector.go           temporary thin compatibility factory
  cdc.go                 approved pglogrepl source; migrate through transaction sink

internal/connectors/native/mysql/
  connector.go           temporary thin compatibility factory
  cdc.go                 existing engine-specific changefeed adapter
```

Add `internal/ownership/identity.go` as the one implementation of the opaque
`{workspace, connector, connection}` identity and `SameIdentity` rule. `warehouse.Owner` and the
database `TargetOwner` should project that shared value; the database framework must not restate the
comparison and drift from `internal/warehouse/layout.go:128-144`.

`internal/connectors/database` may import `internal/synccontract`; the reverse must never occur.
The exact subpackage split can move to avoid cycles, but ownership should not.

`database.schema.json` is a new schema owned by the native database framework. It should not add
`database_write` to the API-oriented `operations.json` enum. The executable declaration should still
be a `synccontract.NativeCommandContract`, for example:

```text
protocol: postgres-wire
command: managed-table-apply
executor: { kind: native, id: postgres.table_write.v1 }
modes: [full_overwrite, full_append, incremental_append,
        incremental_upsert, incremental_dedupe]
```

Those names are concrete enough for the existing native admission rule and carry no SQL payload.

### Named interfaces

The important interfaces should be small and consumer-owned. First, split common native admission
from source-sync dispatch in `internal/synccontract/native.go`:

```go
type NativeExecutorAdmission interface {
    NativeSyncExecutorDescriptor() NativeSyncExecutorDescriptor
    NativeSyncConformanceEvidence() ConformanceEvidence
}

type NativeSyncExecutor interface {
    NativeExecutorAdmission
    RunNativeSync(context.Context, NativeSyncRequest) (NativeSyncResult, error)
}
```

The existing registry should validate and store `NativeExecutorAdmission`; its source `Execute`
method then requires `NativeSyncExecutor`. A database registry resolves that **same admitted object**
and requires `DatabaseWriteExecutor`. This keeps one descriptor/evidence/mode admission path without
forcing a target plan through the source-oriented `NativeSyncRequest`.

The database package shape is:

```go
package database

type Driver interface {
    Open(context.Context, RuntimeConfig) (Connection, error)
}

type Connection interface {
    Introspect(context.Context, CatalogRequest) (Catalog, error)
    OpenRows(context.Context, CompiledSelect) (RowStream, error)
    Begin(context.Context, TransactionOptions) (Transaction, error)
    Close() error
}

type RowStream interface {
    Next(context.Context) (Row, error)
    Close() error
}

type Transaction interface {
    Execute(context.Context, CompiledStatement, ...Value) error
    BulkLoad(context.Context, TableRef, []ColumnRef, RowBatch) error
    Commit(context.Context) (CommitConfirmation, error)
    Rollback(context.Context) error
}

type WriteSession interface {
    Stage(context.Context, ManagedRowBatch) error
    Commit(context.Context) (DurabilityReceipt, error)
    Abort(context.Context) error
}

type DatabaseWriteExecutor interface {
    synccontract.NativeExecutorAdmission
    BeginWrite(context.Context, ApprovedDatabaseWritePlan) (WriteSession, error)
}
```

`ApprovedDatabaseWritePlan` is an execution capability, not a public struct callers can assemble. The
application constructs it only after plan, persisted preview, approval validation, and immediate
pre-execution drift checks. It binds the plan/preview/input/target digests, approval evidence,
`TargetOwner{Workspace, Connector, Connection}`, the derived physical managed-table address, the
observed native object identity, managed-column schema version, and expected owner/control records.

`CompiledSelect` and `CompiledStatement` are opaque framework values with unexported construction
state. Connector callers cannot use them to submit raw SQL. Renderer code constructs them only from
typed plans and structured identifiers. The protocol adapter executes them and binds data values;
values are never interpolated into statement text.

The framework's concrete `database.Executor` should implement `DatabaseWriteExecutor` and its
admission descriptor/evidence. The compatibility `database.Connector` delegates to that executor and
implements these existing product-facing contracts:

- `connectors.Connector` for check/catalog/read and compatibility write dispatch;
- `connectors.WriteValidator` for type/key/target/schema checks;
- `connectors.DryRunWriter` for a mutation-free preview;
- exact `synccontract.NativeCommandContract` admission through the shared native registry;
- the durability destination contract, but only by converting a successful `DurabilityReceipt`, not
  by fabricating an acknowledgement in a later unrelated method call.

`connectors.Connector.Write` remains a one-shot compatibility adapter only for a sealed managed
target: begin, stage the supplied records, commit, and translate the receipt. It must refuse missing
owner identity, managed-target plan, or mismatched approval/preview evidence, so calling it directly
cannot become an arbitrary-table safety bypass. The application uses `BeginWrite` directly for
bounded multi-batch execution. The current `connectors.WriteRequest` is a useful compatibility DTO, but at
`internal/connectors/connectors.go:564-572` it has only stream/table/action/overwrite/config/key and
approval. It needs a typed database plan behind it; it must not grow arbitrary SQL.

### Framework versus driver/config responsibilities

| Component | Framework provides | Driver provides | `database.json` provides |
|---|---|---|---|
| Connection | lifecycle, timeouts, retry classification | wire open/close and auth adapter | protocol ID, safe defaults |
| Identity | shared owner triple, sealed destination fingerprint; no secrets | server/catalog and native object identity observation | managed namespace strategy |
| Catalog | normalized tables, columns, keys, constraints | native metadata introspection | native→logical type rules |
| Read | snapshot and keyset algorithms, bounded delivery | execute/scan/cancel | bind style, ordering capabilities, limits |
| Type planning | exact/lossless/unsupported classification | native value decode/encode | closed mapping entries and bounds |
| SQL construction | typed AST and statement assembly | execute opaque compiled statement | quote/bind/feature enums, never SQL text |
| Bulk load | batch size, validation, staging lifecycle | fastest safe primitive (`COPY`, batched bind, array bind) | parameter/batch limits |
| Sync modes | managed columns, stage/dedupe/delete/apply policy and guarantees | transactional primitives | admitted modes and atomic strategies |
| Ownership | structural address, `SameIdentity`, and control-registry semantics | native schema/table identity and locks | closed ownership primitive |
| Durability | confirmed-commit receipt and checkpoint ordering | native commit confirmation | required durability capability |
| CDC | transaction/event/checkpoint contract and durable-spool adapter | database-specific log protocol | capability true only after exact admission |
| Testing | one reusable conformance corpus | engine fixtures + live adapter | tested capability matrix |

### Strict declarative definition

`database.json` should use `additionalProperties: false` and contain only closed values:

- driver and native protocol IDs;
- catalog/schema support and qualification order;
- identifier quote family, case-fold behavior, maximum byte length, and reserved-word policy;
- bind-marker family;
- transaction and savepoint support;
- whether DDL is transactional;
- staging-table kinds;
- managed ownership namespace/address strategy, native object-identity capability, and provisioning
  recovery strategy;
- the versioned system-column capability;
- atomic target-replacement strategy;
- conflict/update-insert strategy;
- maximum bind parameters and safe default/max batch size;
- supported `synccontract.Mode` values;
- exact native→logical and logical→native type mappings with precision, scale, length, timezone,
  collation, and nullability conditions.

It must not contain:

- SQL text or interpolation fragments;
- arbitrary function names;
- raw query hooks;
- a silent default type such as `VARCHAR`;
- a capability not tied to a renderer primitive and conformance fixture.

If an engine needs behavior outside the closed model, the framework adds a named primitive with red
tests and code review before configuration may select it. This is how “configuration-driven” remains
safe instead of becoming code injection encoded as JSON.

## Read/write symmetry

### Shared lifecycle

Both sides begin with the same normalized `Catalog`, `LogicalType`, structured identifier, and
destination/source identity. A read plan and a write plan should be mirror images:

```text
ReadPlan                              DatabaseWritePlan
────────                              ─────────────────
source identity + generation          destination identity + credential revision
catalog/schema/table refs             owner triple + derived managed table ref
observed source schema hash           observed target schema hash
selected logical columns              selected logical→native TypePlans
snapshot barrier                      pinned warehouse input identity/digest
cursor + tie-breaker ordering         mode + target key + deterministic winner order
bounded page size                     bounded stage batch size
candidate checkpoint                  confirmed-commit receipt + delivery checkpoint
```

The framework must fix a current asymmetry: PostgreSQL reads use only `cursor > lower_bound` and
`ORDER BY cursor` at `internal/connectors/native/postgres/reader.go:62-84`. Equal cursor values can be
skipped across pages. MySQL already demonstrates the better shape—cursor plus primary key at
`internal/connectors/native/mysql/reader.go:273-383,455-474`—but unnecessarily requires a
single-column primary key at lines 401-405. The framework should support a composite stable unique
tie-breaker.

### Reverse-ETL integration

Database writes must plug into the existing flow, not bypass it:

1. **Plan**
   - Resolve the destination without exposing credential values.
   - Resolve the pinned source connection to the same opaque
     `{workspace, source connector, connection ID}` used by `warehouse.LocationFor`; a display name
     is not identity.
   - Derive the owner-scoped physical namespace/table from that identity and the logical stream.
   - Introspect the Polymetrics control registry and physical target. The only admitted states are
     “both absent, create proposed” or “present and owned by the exact identity.” A table with a
     missing, unreadable, or foreign owner record is a hard refusal and is never adopted.
   - Establish the destination server/database fingerprint and native table object identity.
   - Compile all source logical types to target types.
   - Add the versioned, non-overridable `_polymetrics_*` system columns and their indexes/constraints.
   - Resolve mode, key, cursor/tie-breaker, mandatory creation/provisioning plan, batch bounds, and
     target mutation summary.
   - Bind the warehouse input manifest/digest, destination credential revision/configuration digest,
     owner identity, target identity/object ID, managed schema hash, and full `DatabaseWritePlan`
     into the plan seal.
2. **Preview**
   - Re-read metadata only; do not create schemas, tables, or staging objects.
   - Re-check type/key/capability constraints, control-registry ownership, native object identity,
     managed-column version, and target drift.
   - Report rows, bytes, business and system columns, safe owner identity, logical and physical
     target, whether creation is required, exact DDL/indices/control rows, mode, locks expected,
     insert/update/delete estimates where provable, warnings, and a digest.
   - A destructive preview must identify the same approval target that execution will mutate.
3. **Approval**
   - Use the existing one-time token and typed destructive confirmation.
   - Consume approval before the first mutation or stage creation.
4. **Execute**
   - Re-resolve owner identity, destination identity, credential revision, native object/schema
     identity, and pinned input digest.
   - Begin one `WriteSession`; under a target-scoped lock, reassert ownership or transactionally
     create the PostgreSQL control rows, owner namespace, table, system columns, and indexes.
   - Feed bounded warehouse batches, validate the stage, and apply the mode in the same target
     transaction.
   - Return success only with a confirmed `DurabilityReceipt`; then and only then advance delivery
     state.

The current app already hashes and re-reads the approved input and consumes approval before
`writer.Write` (`internal/app/app.go:1734-1806,2106-2181`). Phase one should preserve those gates
while making **persisted preview mandatory for every database plan**, not only destructive actions,
and replacing the all-records-at-once call with a pinned-input batch iterator feeding one write
session. `ReversePlan` also needs typed catalog/schema/table/key/mode fields; a plain `Name` and
string `Action` cannot fully seal a database mutation.

The ordinary `runConnectorETL` batch writer is not an alternate database-destination path: it has no
approved target/owner plan and currently calls `Write` without approval
(`internal/app/app.go:1111-1185`). Phase one must refuse a database destination there and route
database-to-database through warehouse materialization followed by reverse ETL. `ReversePlan` must
persist the opaque source connection ID and connector, not merely its display-name selector, so the
database target and local warehouse share the exact owner triple. A later scheduled database
delivery needs a persisted, explicitly approved plan contract; it cannot inherit approval from
connection creation.

### Write-session state machine

```text
PLANNED
  └─ preview + approval ─→ APPROVED
       └─ BeginWrite ────→ OWNERSHIP_ASSERTED
            ├─ absent target ──→ PROVISIONING → OWNED
            ├─ matching target ─→ OWNED
            └─ absent/foreign/unreadable owner → REFUSED
       └─ OWNED ─────────→ STAGING
            ├─ Stage(batch)*
            ├─ validate row count, type bounds, non-null keys, duplicate winners
            └─ Apply(mode) ─→ COMMITTING
                   ├─ confirmed COMMIT ─→ DURABLE → receipt → checkpoint
                   ├─ known rollback/failure ─→ ABORTED (no receipt/checkpoint)
                   └─ outcome unknown ─→ RECONCILIATION_REQUIRED
                                        (no receipt/checkpoint; never blind success)
```

`Abort` is idempotent. Cleanup errors are joined with, and never replace, the primary failure. Context
cancellation stops staging but cannot turn an unknown commit into a known rollback.

## Exact sync-mode semantics

The requested names should be product profiles over the canonical modes at
`internal/synccontract/mode.go:13-20`; do not add another enum.

| Product profile | Canonical mode | Required inputs | Target guarantee after confirmed commit | Missing key / non-unique cursor behavior | Retry semantics |
|---|---|---|---|---|---|
| `full_refresh` | `full_overwrite` | Complete approved snapshot and managed schema; no source PK required | Create the owner-scoped table when absent. Otherwise require exact ownership, stage all rows, and transactionally replace only that managed table's contents. Every row carries the managed lineage columns. | PK irrelevant. If the dialect cannot make provisioning and replacement recoverable/atomic, the mode is unsupported. | Repeating the same approved snapshot converges table contents; unknown commit still requires operator-visible reconciliation. |
| `append` (full input) | `full_append` | Complete approved input and managed target; no source PK required | Create when absent; otherwise assert ownership. Append all rows atomically for the write session. Existing rows remain. | PK not required; `_polymetrics_raw_id` is lineage, not a substitute source key. | **At-least-once across an unknown commit, per captain ruling.** A retry can duplicate until a separately approved destination idempotency ledger exists. Managed run IDs aid diagnosis but are not proof of commit outcome. |
| `incremental` | `incremental_append` | Cursor plus deterministic total order and committed checkpoint | Append only records strictly after `(cursor, tie_breaker)`, including `_polymetrics_cursor` and lineage. | A unique cursor may stand alone. Otherwise require a stable non-null unique tie-breaker, normally the source PK; fail preflight if absent. | At-least-once at the target; no equal-cursor records may be skipped. A replay may duplicate after unknown commit. |
| `incremental_upsert` | `incremental_upsert` | Cursor order and non-empty target key | Upsert current state in the owned table. A tombstone updates the managed delete state (or applies the separately admitted delete policy) without losing lineage. | Absent/null key is a preflight/runtime error. Non-unique cursor still needs a stable tie-breaker. | Reapplication converges by key, though triggers/external effects may repeat. |
| `incremental_dedupe` | `incremental_dedupe` | Cursor/event order, non-empty stable source PK, deterministic winner order, and managed metadata schema | Collapse repeated source rows for a key within the admitted window; select the greatest ordered event, including tombstones; atomically upsert one current managed row per key with generation/cursor/delete lineage. | **No PK: fail. Null PK: abort. No stable total order: fail. Never downgrade to append or ordinary upsert.** `_polymetrics_raw_id` cannot repair an ambiguous source order. | Target state converges by key for the same ordered window; external side effects are not claimed exactly once. |

`incremental_dedupe_history` and `change_capture` remain canonical future capabilities. They should
not be advertised merely because their names exist. Deletes and history-window behavior continue to
use the existing `synccontract` tombstone/history contract.

The brief uses `incremental_dedup`; the captain ruled that the repository's existing
`incremental_dedupe` spelling is the only public and persisted spelling. Reject the shorter form at
input instead of creating an alias. There is one mode name and one contract.

## Type mapping: lossless or no write

### Logical type algebra

The existing PostgreSQL catalog maps unknown types to `string` at
`internal/connectors/native/postgres/cataloger.go:127-147`. That is acceptable only as a coarse UI
description; it is not safe for target planning. The write framework needs a parameterized algebra:

- signed and unsigned integers with width/range;
- decimal with precision and scale;
- floating point with width;
- boolean;
- UTF-8 string with optional maximum length and collation metadata;
- binary with optional maximum length;
- date;
- time with precision and timezone presence;
- timestamp with precision and explicit with/without-timezone semantics;
- UUID;
- JSON;
- explicitly supported arrays/collections with element types;
- `OpaqueNative(engine, name, modifiers)` for discovered but unmapped source types.

Nullability belongs to the column contract, not the scalar type. Default values, generated/identity
columns, collations, and constraints must also be preserved in catalog metadata because they affect
whether a write is legal.

### Compilation rules

For every column the planner produces a `TypePlan`:

| Classification | Meaning | Phase-one action |
|---|---|---|
| `exact` | Same value domain and semantics | Execute |
| `lossless` | Target domain is a proven superset, e.g. narrower integer → wider integer | Execute and show mapping in preview |
| `explicit_transform_required` | A user-visible transform could make it valid but changes semantics | Refuse until a typed, plan-hashed transform is explicitly configured |
| `unsupported` | Target cannot faithfully represent it | Fail planning with source/target column and reason |

There is no default `VARCHAR`. Precision reduction, integer narrowing, float↔decimal conversion,
timezone removal, binary→text, JSON stringification, array stringification, collation change, or
truncation is never implicit.

Planning validates structural ranges from catalog metadata. Execution validates actual values before
or while staging, because sources can report imprecise metadata. One invalid value rolls back the
whole write session; partial success is not a valid database-sync result.

Any future explicit transform must be a typed, closed operation (for example an explicitly chosen
UTC normalization), appear in preview, and be bound into the plan hash. A driver may not hide a cast.

## Identifier and target identity safety

Use structured `CatalogRef`, `SchemaRef`, `TableRef`, and `ColumnRef` values. Renderer-owned quoting
must escape the dialect's quote character and qualification rules; data values use bind parameters.
Do not concatenate a user string into SQL.

The database target name is a logical name. Phase one does not resolve it to an arbitrary
user-supplied physical table. The framework derives a compact owner namespace from the length-
delimited opaque `{workspace, connector, connection}` identity and a table token from the logical
stream. It stores the full identity beside that address; a token collision therefore produces a
foreign-owner refusal rather than shared rows. Validate identifier length, encoding, forbidden
control/NUL characters, qualification count, and engine rules before rendering.

PostgreSQL uses a Polymetrics control schema with versioned owner and managed-table relations. Each
managed-table row records the full owner identity, logical and physical refs,
managed-column/schema fingerprint, and the observed `pg_namespace`/`pg_class` object identity. A
physical table that exists without this record, whose record names another identity, whose native
object ID changed, or whose record is unreadable is never adopted, repaired, dropped, or written.
Creation of the owner namespace/table/indices and owner record is part of the approved PostgreSQL
transaction. Display-name changes do not change identity, matching `warehouse.Owner.SameIdentity`.

The destination fingerprint should bind:

- connector and protocol;
- server/database/catalog identity observable without secrets;
- credential revision and non-secret configuration digest;
- the shared workspace/connector/connection owner triple;
- logical name, derived physical schema/table refs, control-registry version, and native object ID;
- managed-column version and observed target schema/constraint hash.

Execution must re-establish the same fingerprint under a target-scoped lock. Missing, disabled,
unowned, foreign, or changed identity fails closed. This directly rejects both the old tenant
design's fallback to a default database and the shared-target collision class fixed by PR #3901.

## Durability and acknowledgement

### The defect this must prevent

`/Users/karthiksivadas/karthik-agent-workspace/data/learnings.md:6-38` names the repository's dominant
defect class: asserting a fact before establishing it. Its concrete PostgreSQL incident is at lines
17-19: CDC acknowledged before the downstream write was durable, the replication slot advanced,
WAL was discarded, and data became permanently unrecoverable.

The design rule is therefore:

> An acknowledgement proves the fact the next checkpoint depends on; it is never a report that an
> attempt was made or that this process observed an intermediate state.

### Target transaction boundary

For PostgreSQL phase one:

1. Approval is consumed before any staging mutation.
2. Open a target transaction, acquire the owner-scoped target lock, and positively establish the
   required native durability settings; for PostgreSQL, require server `fsync=on` and set/verify
   `synchronous_commit=on` for the transaction. Refuse execution when the server advertises a mode
   in which a successful commit need not be flushed. This proves durability to the database's local
   committed-WAL contract, not backup or replica durability unless a stronger policy is separately
   declared and checked.
3. Re-query the control registry and native object ID. If the target is absent exactly as previewed,
   create the owner namespace, managed table, system columns, keys/indexes, and owner record in this
   transaction. If a physical table is present without the exact owner assertion, stop.
4. Create a private temporary stage and bulk load every bounded batch through the existing pgx
   dependency.
5. Validate counts, type bounds, key nullability, duplicate ordering, system-column integrity,
   ownership, and target drift.
6. Apply the selected mode inside the same transaction.
7. Receive a successful server response to `COMMIT` under the established durability setting.
8. Only now construct `DurabilityReceipt{run, owner, native target identity, mode, input digest,
   committed_at}` from the driver's confirmed commit fact.
9. Convert that receipt to the unforgeable downstream acknowledgement accepted by
   `synccontract.CommitAfterDownstreamAcknowledgement`.

The current `DurableETLDestination.AcknowledgeETLDurability` is called after all batch writes at
`internal/app/app.go:1190-1205`. Its ordering is good, but the method is logically separable from the
commit it claims. The database path should make the receipt a return value of `WriteSession.Commit`,
then adapt it immediately to the existing synccontract acknowledgement. A later method must not be
able to manufacture database durability from only a run ID.

### Failure outcomes

| Failure point | Target state | Receipt? | Checkpoint? | Required response |
|---|---|---:|---:|---|
| Before transaction | unchanged | No | No | Return typed validation/connection error |
| During stage/validation/apply, rollback confirmed | unchanged | No | No | Abort/rollback; retry from pinned input |
| COMMIT rejected and rollback/abort known | unchanged | No | No | Return failure |
| Connection lost before COMMIT was sent | presumed uncommitted, but verify driver evidence | No | No | Rollback/reconnect as safe; retry |
| COMMIT sent, outcome not knowable | unknown | **No** | **No** | `CommitOutcomeUnknown`; stop for reconciliation. Never blindly retry append or claim success |
| COMMIT confirmed under required durability | committed | Yes | Only after receipt | Return success and commit delivery state |
| Target committed, local checkpoint persistence fails | committed target, old checkpoint | Confirmed receipt existed in the completed process | No new checkpoint yet | Upsert/dedupe can converge on replay; append may duplicate. Surface the uncertainty and reconcile honestly |

A cross-database distributed transaction is neither required nor recommended. The durable warehouse
handoff and explicit state machine make the two legs independently recoverable.

### Append and exactly-once honesty

The captain already ruled that phase-one append is **at least once** after an unknown commit and that
the destination idempotency ledger is deferred. The managed-table ruling does not silently override
that decision. `_polymetrics_run_id` and the ownership registry improve diagnosis, but neither alone
is a transaction receipt for zero-row writes, full replacement, or every mode.

Design the future ledger port now so it can atomically record
`(owner, managed target, attempt, plan hash, input digest, mode)` in the row transaction without
changing sync-mode semantics. The newly mandatory Polymetrics control namespace makes that later
addition cheaper, so the captain may revisit timing, but phase-one docs and contracts must continue
to say that an append retry after unknown commit can duplicate. Upsert/dedupe convergence still does
not prove external side effects occurred exactly once.

## PostgreSQL-to-PostgreSQL end to end

### Recommended phase-one route

1. Resolve and fingerprint the PostgreSQL source.
2. Discover its table, logical types, primary/unique keys, and cursor candidates.
3. Begin a stable snapshot. For incremental reads, use lexicographic keyset paging on
   `(cursor, stable_unique_tie_breaker...)` under one defined snapshot/barrier contract.
4. Emit bounded batches into the source connection's warehouse JSONL WAL; fsync the WAL and required
   directory chain before acknowledging extraction progress.
5. Rebuild/publish the connection-owned single Parquet table and an immutable input manifest/digest.
6. Persist the extraction checkpoint only after that replay boundary is established. Keep the
   warehouse materialization available until delivery succeeds.
7. Create a reverse plan bound to the materialization, destination identity, target schema, mapping,
   mode, key, and the same workspace/source-connector/connection-ID owner as the materialization.
8. Preview, approve, and consume approval using the existing application flow.
9. Stream the pinned Parquet input in bounded batches into one PostgreSQL managed `WriteSession`;
   create the owner-scoped table if this is its first approved delivery, otherwise assert the stored
   owner/native object identity.
10. On confirmed target commit, produce the durability receipt and advance the delivery checkpoint;
    an unknown commit stops for reconciliation under the phase-one at-least-once contract.

Extraction and delivery state must be distinct. Once the local warehouse is a durable replay
boundary, an extraction checkpoint may advance before the customer destination commits; the
end-to-end sync is not complete until the delivery checkpoint advances. Retention must ensure the
advanced extraction state never points past data that has been discarded locally.

### Why not direct streaming first

| Concern | Warehouse-mediated | Direct source→target stream |
|---|---|---|
| Source/target failure window | Durable handoff separates failures | Must solve a two-system acknowledgement window |
| Retry | Reload without rereading source | Often rereads source; append duplicates likely |
| Preview/approval | Operates on immutable pinned input | Input can change while waiting for approval |
| Memory | Bounded batches | Can be bounded, but needs its own spool for recovery |
| Audit | Input manifest and warehouse table retained | Must build equivalent evidence separately |
| Cost | WAL + Parquet + target I/O, local disk, added latency | Lower steady-state I/O/latency |
| Correct phase-one choice | **Yes** | No |

The cost is real: roughly one local WAL write, one Parquet materialization, one Parquet read, and one
target write, plus temporary staging. That is acceptable for the reference implementation because it
buys correctness and operability. Measure it before adding a fused path. If a later fused path is
needed, it may skip the Parquet publication but must retain a durable spool/run ledger and the exact
same plan/approval/durability rules.

## Judgment of the Ruby prior art

The old implementation contains useful decomposition and algorithms. It should be treated as design
evidence, not ported line for line.

| Ruby component | What it did | Verdict | Go/framework change |
|---|---|---|---|
| `Core::BaseConnector` (`base_connector.rb:5-28`) | One abstract object for connect/read/write/catalog/query validation | **Keep the capability shape; reject the monolithic inheritance contract.** | Use small structural interfaces at consumers. A read-only engine need not fake write methods, and a driver is not forced to own sync policy. |
| DuckDB `Client` (`client.rb:5-31`) | Composed connection, writer, and query validator, but never wired the existing reader or catalog methods required by `BaseConnector` | **Keep composition; reject incomplete runtime assembly.** | `database.Connector` composes only supported capabilities; compile-time interface assertions, exact descriptor admission, and conformance prevent metadata from claiming an unwired method. |
| `connection.rb` | Opened local or MotherDuck connections | **Revise.** | Context-aware open/close, typed secret-safe errors, shared `sqltls`, bounded pools, destination fingerprinting, and no credential interpolation into a URI/log. |
| `reader.rb:12-45` | Table or arbitrary query with limit/offset | **Reject for sync.** | No arbitrary SQL in connector sync. Use structured table refs and stable keyset pagination; limit/offset without an order is neither stable nor resumable. |
| `reader.rb:58-60` | Counted an arbitrary subquery | **Reject as a planning primitive.** | Introspection/preview may estimate using safe native metadata; it must not execute caller SQL or promise exact counts when it cannot prove them. |
| `query_validator.rb:16-81` | `EXPLAIN`, then wrapped the query and assumed success when wrapping failed | **Reject.** | The generic validator can accept mutating or unwrappable statements and even returns `true` on analysis failure. Database writes use typed plans, not “validated” raw SQL. |
| Writer identifier checks (`writer.rb:39-83`) | Strict regex plus quoting | **Keep the intent, revise the model.** | Structured refs, catalog resolution, dialect-owned quoting, length/control/qualification validation. Note that Ruby validated `database_name` but did not use it in target construction—a sealed destination fingerprint prevents that class. |
| PK checks (`writer.rb:51-64`) | Confirmed named keys exist when provided | **Keep and strengthen.** | Dedupe/upsert require a non-empty key; enforce uniqueness metadata, non-null staged values, composite order, and no silent fallback. |
| Schema/table creation (`writer.rb:85-112`) | `CREATE ... IF NOT EXISTS` automatically | **Keep creation; reject unowned `IF NOT EXISTS`.** | A database destination must be Polymetrics-created. DDL is plan-visible, previewed, approved, identity-scoped, and paired with an owner assertion/native object ID. An existing unowned or foreign table is refused, never adopted. No implicit ALTER. |
| Appender bulk path (`writer.rb:123-157`) | Bulk appended inside an explicit transaction | **Keep.** | Make bulk load a driver primitive under framework-owned session orchestration. Context cancellation, bounded batches, joined cleanup errors, and native durability confirmation are mandatory. |
| Value formatting (`writer.rb:167-170`) | Serialized hashes/arrays to JSON | **Reject as an implicit rule.** | JSON/array mapping must be explicitly lossless for the target; no automatic stringification. |
| Runtime type mapping (`writer.rb:180-188`) | Uppercased any incoming type string | **Reject.** | Closed `LogicalType` and exact renderer mapping. Caller strings never become SQL types. |
| `mapping.json:1-18` | Declared mappings and default `VARCHAR` | **Keep a strict declarative map; reject the default.** | Ruby runtime ignored this file, and its default silently coerces unknowns. Go validates the map at load and fails unknown/unsupported types. |
| `connection_specification.json:1-38` | Closed local-path versus MotherDuck-token credential alternatives | **Keep the closed/discriminated shape; separate secrets from engine policy.** | Connector configuration continues to declare user inputs and secret fields. `database.json` separately declares dialect/type/capability policy. Drivers receive resolved secret handles, never embed tokens in logged URIs as `connection.rb:38-42` did. |
| Temporary staging (`writer.rb:194-209`) | Created temp table and bulk-loaded it | **Keep.** | Deterministic per-run private stage, bounded loads, type/key/count validation, and cleanup within a formal session. |
| Upsert (`writer.rb:211-250`) | `ON CONFLICT`; no PK silently appended | **Keep staging/upsert; reject fallback.** | Mode-specific plans. Missing PK fails. Resolve PK-only schemas without emitting an empty update clause. Dedupe winners are deterministic. |
| Transaction boundary | Appender transactions existed, but staging/upsert/cleanup were separate calls without one explicit end-to-end durable receipt | **Reject as insufficient.** | All target mutations for the session are governed by one state machine; acknowledgement is constructed only from confirmed commit. |
| `metadata.json:1-19` | Declared source full/incremental, destination append/overwrite/upsert, bulk load, and transactions | **Keep the taxonomy; reject declaration without execution evidence.** | Map profiles to the existing `synccontract.Mode` values and expose only modes proven by the exact registered executor and conformance suite. The old client did not even wire `Reader`, showing why metadata alone is insufficient. |
| Multi-tenant identity/isolation | Separated platform and tenant databases, pooling, TLS/audit/health concepts | **Keep the identity boundary.** | Destination identity is explicit, sealed, least-privilege, TLS-aware, and fail-closed. |
| Tenant fallback (`architecture.md:488-507`, `solution.md:163-174`) | Missing/disabled/not-found tenant used the default/platform DB | **Reject categorically.** | Never redirect a requested target. Return a typed identity/configuration error before mutation. |
| Tenant status/provisioning (`solution.md:1119-1149`) | Marked status in steps separate from external durable facts | **Revise.** | Derive status from receipts/reconciliation, not intent. Do not mark healthy merely because calls returned in sequence. |
| Provisioning SQL (`solution.md:1168-1183`) | Interpolated database/user/password values into SQL | **Reject.** | No generic provisioning in the sync framework; if later added, use engine-specific safe identifier APIs and secret-safe protocol operations under a separate approval contract. |

### Where the Ruby design was most right

- One public writer entry point with explicit table/schema/primary-key inputs.
- Early identifier and primary-key validation.
- Separate connection, read, write, validation, metadata, and mapping responsibilities.
- Staging before merge/upsert and use of a bulk appender.
- A recognizable taxonomy of full refresh, incremental, append, upsert, and dedupe.
- Treating tenant database identity and transport security as first-class.

### Where it was most wrong

- Dynamic conventions were not contracts: `mapping.json` did not control runtime mapping.
- The client left declared read/catalog capabilities unwired while metadata still advertised broad support.
- Arbitrary type strings became SQL types, and unknowns defaulted to text.
- Query “validation” could execute/accept unsafe shapes and failed open.
- Offset pagination lacked deterministic order and durable resume semantics.
- Missing primary keys silently changed upsert into append.
- Staging, apply, cleanup, and acknowledgement were not one typed durability state machine.
- Tenant lookup failed open into the platform database.
- Destination/database identity was validated but not consistently used.

The Go implementation should preserve the decomposition and bulk/staging ideas while replacing
dynamic trust with compile-time interfaces, strict schemas, typed plans, context-aware lifetimes,
shared conformance, and explicit durable facts.

## Migration path for existing PostgreSQL and MySQL packages

### PostgreSQL: reference implementation

Do not rewrite every read component before write ships. Use the current native package as a
compatibility shell:

1. Add the framework and PostgreSQL `database.json`.
2. Adapt existing connection parsing/pooling and shared `sqltls` into
   `database/drivers/postgres.Driver`.
3. Add native introspection sufficient for exact target columns, constraints, keys, generated
   columns, type modifiers, control-registry state, and `pg_namespace`/`pg_class` identity.
4. Implement the shared owner identity, managed-column schema, PostgreSQL transactional provisioning,
   confirmed-commit durability receipt, `ValidateWrite`, mutation-free `DryRunWrite`, and
   `WriteSession`. Keep the future destination idempotency-ledger port explicit but unimplemented.
5. Register `postgres.table_write.v1` as a `DatabaseWriteExecutor` through the shared native
   admission descriptor and conformance evidence.
6. Route the existing reverse-ETL writer into the session while retaining the native factory/name.
7. Prove PostgreSQL source → warehouse → PostgreSQL destination live before changing
   `capabilities.write` to true.
8. In the next phase, move PostgreSQL catalog/read policy into the shared reader and fix the
   single-cursor resume hole.
9. Rebase and land the approved pglogrepl CDC source from
   `origin/fm/cli-postgres-cdc-logical-replication-r1`; preserve its protocol/lifecycle core and
   apply the concrete restructuring in the correction addendum. Do not rebuild it in the database
   writer.

No legacy Go files need to disappear in the first slice. Once shared read/write conformance is green,
reduce them to adapter/factory/CDC code and remove duplicated quoting/type/catalog logic in coherent
steps.

### MySQL: second implementation and abstraction test

MySQL is valuable precisely because it will expose PostgreSQL assumptions:

1. Add a MySQL definition selecting closed renderer and mapping primitives.
2. Adapt the existing Go MySQL wire dependency, TLS modes, connection, catalog, and row decoding.
3. Reuse the current cursor+keyset lessons, generalizing beyond its single-column-PK restriction.
4. Implement MySQL bulk/staging/apply using the same `WriteSession`, owner assertion, and
   confirmed-commit semantics. Because MySQL DDL can implicitly commit, its provisioning adapter needs a
   durable, idempotent `provisioning → ready` recovery state rather than pretending PostgreSQL's
   single transactional-DDL boundary is portable.
5. Run the identical conformance corpus and the existing `native/dbtest` live harness; do not copy
   the harness ([maintainer guide](../../internal/connectors/native/dbtest/README.md)).
6. Add PostgreSQL→MySQL and MySQL→PostgreSQL matrices.
7. Only then set MySQL write capability true and delete duplicated policy from its native package.

The current MySQL binlog reader can remain database-specific. Its public CDC capability remains
truthful only when its executor and synccontract evidence are actually registered and admitted.

## Roadmap and cost after the framework

| Phase/database | What remains after shared framework | Expected cost | Main risks |
|---|---|---:|---|
| 1. PostgreSQL write + PostgreSQL→PostgreSQL | Framework itself; managed namespace/control registry/system columns; transactional provisioning; pgx bulk adapter; confirmed-commit receipt; reverse lifecycle integration; live harness | **High** (reference investment) | ownership/object-ID drift, unknown append commit, durability setting, atomic full refresh, long transactions, approval/batch integration |
| 2. PostgreSQL read/catalog migration | Move existing policy; composite keyset/snapshot semantics; exact source typing | **Medium** | equal-cursor loss, snapshot barriers, exotic PostgreSQL types |
| 3. MySQL | Definition + existing protocol adapter + managed registry/physical naming + recoverable implicit-commit provisioning + bulk/apply + cross-engine tests | **Medium–high** | DDL implicit commits, ownership recovery, locking/version behavior, type/time differences, key limits |
| 4. MariaDB | Reuse MySQL protocol/provisioner when proven; separate definition/version matrix | **Medium** | DDL/recovery divergence masquerading as MySQL compatibility |
| 5. SQL Server | New protocol driver; managed schema/extended native identity; staging/bulk; parameter limits; identity/computed handling; update-then-insert primitive | **High** | unsafe `MERGE`, collation, locks, ownership metadata, transactional DDL variation |
| 6. Oracle | Future protocol/package research; managed registry with recoverable DDL state; array binding; `NUMBER`/time/identifier semantics; sequences | **Very high** | DDL auto-commit, ownership reconciliation, client packaging, lossy types, atomic refresh feasibility |
| 7. CDC integration + optional fused transport | Existing PostgreSQL pglogrepl source adaptation, gap-free bootstrap, durable transaction spool, and optional transport optimization | **High, integration-focused** | publication scope, slot/log retention, ack order, schema/key changes, deletes, approval windows |

“Add a database” after phase three should mean: implement/adapt a protocol driver, write a strict
definition, and pass shared conformance. It should not mean copy a connector, cataloger, reader,
writer, mode logic, safety flow, and test harness.

## Phased implementation plan (TDD)

### Phase 1 — PostgreSQL write and PostgreSQL-to-PostgreSQL vertical slice

Phase one intentionally includes the shared write core; otherwise PostgreSQL would become another
bespoke implementation.

1. **Red: definition/type admission**
   - Reject unknown config fields, SQL/template fields, unknown renderer primitives, false capability
     claims, unknown native types, lossy mappings, precision/range loss, reserved system-column
     collisions, and secret-bearing errors.
2. **Green: typed foundation**
   - Extract shared native admission from source dispatch; implement strict definition loading,
     logical types, type-plan compilation, structured identifiers, target fingerprinting, typed
     statements, shared owner identity, managed columns/addresses, and PostgreSQL definition/driver.
3. **Red: write modes and transactions**
   - Live and unit tests for full overwrite, full append, incremental append, upsert, dedupe,
     composite keys, null keys, duplicate winners, mapping failure, rollback, cancellation,
     target/schema drift, table missing/foreign/unreadable owner records, two connection identities
     with the same logical table, native object replacement, and unknown commit outcome.
4. **Green: write session**
   - Implement assert-or-create managed target → stage → validate → apply → confirmed commit,
     with bounded batches and a durability receipt available only after an unambiguous commit.
     Preserve `CommitOutcomeUnknown`; do not add the deferred destination idempotency table.
5. **Red: app safety and durability**
   - Prove no database mutation before approval consumption; every database write has a persisted
     preview; preview has no mutations; plan/owner/native-target/schema/input drift fails; approval
     cannot replay; ordinary ETL cannot bypass the database plan; no checkpoint precedes receipt.
6. **Green: native/app attachment**
   - Register the concrete native write executor through the common admission registry; adapt reverse
     ETL to a pinned batch iterator and one write session; preserve existing CLI safety gates.
7. **Red/green: end-to-end live proof**
   - Add the minimum PostgreSQL source path needed by this slice: one stable snapshot streamed under a
     defined database snapshot in bounded application batches and deterministically digested. Full
     resumable/incremental read migration remains phase two.
   - Use two PostgreSQL instances/databases through `dbtest`; source fixture larger than batch size;
     exact returned and target row counts; injected mid-stage/apply failures; dedupe; full refresh;
     checkpoint/retry assertions. Add two source connection IDs targeting the same logical table and
     prove they resolve to separate managed namespaces; remove/replace/foreign-stamp an owner record
     and prove exit is non-zero with neither table changed. Inject an unknown append commit and prove
     it returns uncertainty without a checkpoint or blind retry. The regression must assert data,
     not exit status.
8. **Parity and verification**
   - Update metadata only after preflight is executable; runtime help, bare namespace behavior where
     affected, `docs/cli/**`, website docs, generated help/manual, connector inspection, surface-sync,
     tests, vet/build, non-suite verification gates, full CI, verify-work, and code-review.

Phase-one invariant: create the managed target when both its derived physical object and owner record
are absent; otherwise require exact ownership. Never adopt an arbitrary table. Creation is sealed in
plan/preview/approval and transactional on PostgreSQL. Automatic business-column ALTER remains
excluded until a separate managed-schema migration contract exists.

### Phase 2 — PostgreSQL shared read/catalog

- Red tests for exact catalog types/modifiers, quoted discovered identifiers, composite PK/unique
  constraints, stable cursor+tiebreaker paging, snapshot barrier, cancellation, and schema drift.
- Move all public catalog/read policy into the framework, generalize the phase-one snapshot path, and
  keep the native factory plus the admitted pglogrepl CDC adapter.
- Prove equal-cursor records across more than one page are complete.

### Phase 3 — MySQL migration and bidirectional database sync

- Run the same red/green conformance against MySQL.
- Implement MySQL renderer/bulk primitives without forking mode policy.
- Prove PostgreSQL↔MySQL through the warehouse and preserve shared TLS semantics.

### Phases 4–6 — MariaDB, SQL Server, Oracle

- Add each through driver + definition + conformance.
- Extend the framework only with closed, named primitives that have cross-driver tests.
- Leave a mode unavailable when the engine cannot prove its atomic/durable guarantee.

### Parallel prerequisite — land and harden PostgreSQL CDC

- The `pglogrepl` decision is approved. Rebase the active CDC lane, preserve its source identity,
  slot lifecycle, resume envelope, and post-durability standby feedback, and complete the P0 fixes in
  the correction addendum before advertising database-to-database CDC.
- Do not make PostgreSQL write wait for a rewrite of the protocol core. The two lanes meet at the
  transaction sink and managed-record enrichment boundary.

### Phase 7 — CDC integration and optional low-latency transport

- Add gap-free snapshot/changefeed bootstrap and relation/schema/publication fingerprinting around
  the existing pglogrepl implementation.
- Frame each source transaction in the connection-owned warehouse WAL; fsync its commit marker and
  extraction checkpoint before acknowledging the LSN.
- Materialize immutable delivery windows, enrich them with the managed `_polymetrics_*` schema, and
  pass them through plan/preview/approval into the managed target writer.
- Measure warehouse cost; add a fused path only if it retains a durable transaction spool, truthful
  unknown-outcome behavior, and identical owner/approval identity.

## Verification strategy

### Unit and static contract

- Definition schema denies unknowns and executable text.
- Every admitted mode has renderer/driver support and complete conformance evidence.
- Type matrices exercise boundaries, not only friendly examples.
- SQL rendering fuzz/property tests cover quotes, qualification, reserved words, control characters,
  maximum lengths, and parameter binding; no value appears in statement text.
- Error tests ensure secrets and raw connection strings never appear.
- Context and cleanup tests cover cancellation at open/read/stage/apply/commit boundaries.

### Driver conformance

Every driver runs the same fixtures:

- catalog and logical type fidelity;
- composite deterministic read order and resume;
- full/append/incremental/upsert/dedupe semantics;
- missing/null/duplicate key behavior;
- unsupported type and target drift refusal;
- all-or-nothing multi-batch writes;
- rollback and commit-outcome classification;
- exact capability/native-executor admission.

### Live database harness

Use `internal/connectors/native/dbtest`, whose [maintainer guide](../../internal/connectors/native/dbtest/README.md)
owns the `Config` setup; add an engine without copying the harness. Keep integration opt-in and
follow that guide's explicit Docker-or-Podman runtime and direct-local-Unix-endpoint contract,
uniquely owned resources, unconditional cleanup, and sequential engines by default.

For each live write test, query the destination and assert row values/counts. Inject failures before,
during, and at the commit boundary. The original direct-read defect exited 0 while losing records;
exit status alone is never sufficient evidence.

### Local gates for an implementation PR

Follow repository policy: focused changed-package tests and `internal/cli` with `-timeout 20m`, then
vet/build and the non-suite `make verify` gates individually. Let CI carry the full 550+ connector
suite when per-command timeouts make one local invocation unreliable. Finish with GSD verify-work and
code-review, plus the automated review routing contract. No write capability becomes true before all
of this is green.

## Explicit captain decision record

All six questions now have captain rulings. The options remain here so an implementation lane can
see the rejected alternatives and avoid reopening them accidentally; there are no unresolved
captain decisions in this report.

### 1. Destination table ownership — **resolved by captain**

The destination is a Polymetrics-created, Polymetrics-owned managed table, structurally scoped by
the same workspace + source connector + connection ID as `warehouse.LocationFor`. Phase one creates
it through the approved plan and refuses an arbitrary existing table, a missing/unreadable ownership
record, a foreign `SameIdentity`, or a changed native object identity. Automatic schema evolution is
still a separate future contract.

### 2. Append delivery after unknown commit — **resolved by captain**

- **Option A — at-least-once append.** Omit a destination attempt/idempotency ledger even though a
  managed control namespace exists; retry may duplicate. A confirmed commit still returns the
  ordinary `DurabilityReceipt` used for checkpoint ordering.
- **Option B — PM-owned receipt table.** Atomically record owner/target/attempt/plan/input identity for
  reconciliation and duplicate suppression inside the already-required control schema.
- **Option C — claim exactly once without a ledger.** Not technically supportable.

**Ruling: Option A for phase one; Option C rejected.** State at-least-once prominently and design the
ledger abstraction so Option B can be added later without changing modes. The managed ownership
registry lowers Option B's future provisioning cost but is not itself a commit ledger and does not
supersede this ruling.

### 3. Should database-to-database bypass the warehouse? — **resolved by captain**

- **Option A — warehouse-mediated durable handoff.** Higher disk/I/O, strongest replay and approval
  identity.
- **Option B — direct streaming.** Lower latency/I/O, but needs a new durable spool and two-system
  recovery before it is safe.
- **Option C — offer both in phase one.** Doubles the initial correctness surface.

**Ruling: Option A.** Revisit Option B only after measuring phase-one cost; never ship a direct path
without a durable spool.

### 4. Which public spelling should users see? — **resolved by captain**

- **Option A — canonical `incremental_dedupe`.** Matches persisted `synccontract.Mode`.
- **Option B — accept `incremental_dedup` as an input alias.** Normalize immediately and persist only
  `incremental_dedupe`.

**Ruling: Option A.** `incremental_dedupe` is the sole public and stored spelling. Do not add the
`incremental_dedup` alias.

### 5. PostgreSQL CDC and `pglogrepl` — **resolved by captain**

`pglogrepl` is approved and the implementation lane is active. The exact pinned branch dependency is
`github.com/jackc/pglogrepl v0.0.0-20260401131349-e37c41485510`, with
`github.com/jackc/pgx/v5 v5.10.0`. The remaining design work is how to merge its working
protocol/lifecycle core with the framework's transaction, type, bootstrap, and
managed-target contracts. The correction addendum gives that direction.

### 6. Should the framework live in a separate repository? — **resolved by captain**

- **Option A — keep it in `cli`.** One atomic change boundary with app safety, sync state, definitions,
  warehouse, tests, and release.
- **Option B — create a library repository now.** Independent versioning but immediate coordination
  and compatibility burden.
- **Option C — incubate in `cli`, extract after PostgreSQL and MySQL plus a second consumer.**

**Ruling: Option C, which means no new repository now.** Extraction is justified only after the API
is proven by at least two engines and another product actually needs it.

## Risks and deliberate non-goals

### Primary risks

- A supposedly declarative dialect grows escape-hatch SQL and becomes a remote code surface.
- PostgreSQL-specific assumptions leak into the core and make MySQL the first rewrite.
- A target commit is reported durable under a server setting that permits asynchronous loss.
- Full refresh is implemented as multiple committed batches and exposes a partial target.
- Equal cursor values are skipped because a stable tie-breaker is absent.
- Target schema changes after approval but before execution.
- A shared physical target or forged/missing owner record lets one connection replace another's rows.
- Append is retried blindly after an unknown commit or marketed as exactly once without the deferred
  destination idempotency ledger.
- Capability metadata is flipped before native preflight and live conformance pass.
- DDL, triggers, foreign keys, generated columns, or permissions make a nominal mode non-atomic.

Each is addressed by a closed config schema, shared conformance, structural owner scoping plus an
independent in-database assertion, explicit target/type/order plans, one write session, target
re-fingerprinting, truthful delivery semantics, and capability admission.

### Non-goals

- Generic SQL query or mutation tools.
- Adoption or mutation of arbitrary pre-existing customer tables.
- Automatic business-column schema alteration after managed-table creation.
- Hidden lossy casts or default-to-string behavior.
- Cross-database distributed transactions.
- Direct stream transport in phase one.
- Reimplementation of the approved PostgreSQL pglogrepl protocol core inside the writer framework.
- Windows support.
- Moving the framework to another repository now.

## Evidence and commands

### Workflow and provenance

```text
scripts/gsd doctor
  → passed; 69 commands available

scripts/gsd sources discuss-phase
scripts/gsd sources plan-phase
  → official GSD Core source resolved through the project-local Pi adapter

scripts/gsd prompt discuss-phase cli-database-connector-framework-design-r1 --auto
scripts/gsd prompt plan-phase cli-database-connector-framework-design-r1 --auto --tdd --skip-research
  → prompts generated; scout slug handled by documented manual inline fallback

go run ./cmd/agentcontractgen check
  → passed

bin/fm-decision-hold.sh complete cli-database-connector-framework-design-r1 \
  target-table-creation append-delivery-guarantee database-transfer-path \
  sync-mode-spelling postgres-cdc-pglogrepl framework-repository
bin/fm-decision-hold.sh verify cli-database-connector-framework-design-r1
  → initial report verified; follow-up rulings resolve target-table ownership and pglogrepl.
    The completion gate is rerun after those structured decisions and this report are updated.
```

Planning artifacts produced in the disposable worktree:

```text
.planning/phases/cli-database-connector-framework-design-r1/CONTEXT.md
.planning/phases/cli-database-connector-framework-design-r1/DISCUSSION-LOG.md
.planning/phases/cli-database-connector-framework-design-r1/RESEARCH.md
.planning/phases/cli-database-connector-framework-design-r1/PLAN.md
.planning/phases/cli-database-connector-framework-design-r1/VALIDATION.md
```

### Focused verification

```text
go test -count=1 -timeout 20m ./internal/synccontract
  → ok   polymetrics.ai/internal/synccontract  0.298s

go test -count=1 -timeout 20m ./internal/connectors/native/postgres \
  ./internal/connectors/native/mysql
  → ok   polymetrics.ai/internal/connectors/native/postgres  0.761s
  → ok   polymetrics.ai/internal/connectors/native/mysql     1.389s
```

These tests verify the inspected baseline only; no implementation was attempted.

The write-surface count was rechecked directly:

```text
for file in internal/connectors/defs/*/writes.json; do
  dir=${file%/writes.json}
  jq -r '.integration_type' "$dir/metadata.json"
done | sort | uniq -c

 239 api
   1 queue
```

### Current-code evidence index

| Claim | Evidence |
|---|---|
| PostgreSQL/MySQL writes unsupported | `internal/connectors/native/postgres/connector.go:97-101`; `internal/connectors/native/mysql/connector.go:43-46` |
| Write capabilities false | `internal/connectors/defs/postgres/metadata.json:7-18`; `internal/connectors/defs/mysql/metadata.json:8-22` |
| No database operation kind | `internal/connectors/engine/schema/operations.schema.json:19-36` |
| Current batch write DTO lacks sessions | `internal/connectors/connectors.go:564-602` |
| Preview extension contracts | `internal/connectors/connectors.go:615-644` |
| Seven canonical modes | `internal/synccontract/mode.go:8-31` |
| Durable ack before checkpoint | `internal/synccontract/commit.go:11-31,51-96` |
| Named native executor/conformance | `internal/synccontract/native.go:15-120` |
| Contract modes currently refused without executor | `internal/app/app.go:1038-1050` |
| Current ETL writes bounded batches independently | `internal/app/app.go:1111-1185` |
| Current durability acknowledgement follows writes | `internal/app/app.go:1190-1205` |
| Reverse plan and seal | `internal/app/app.go:1368-1483` |
| Preview re-hash/dry run | `internal/app/app.go:1734-1806` |
| Approval consumption before write | `internal/app/app.go:2106-2181` |
| PostgreSQL single-cursor order hole | `internal/connectors/native/postgres/reader.go:62-84` |
| MySQL cursor+PK paging and current restriction | `internal/connectors/native/mysql/reader.go:273-383,401-474` |
| PostgreSQL unknown type → string | `internal/connectors/native/postgres/cataloger.go:127-147` |
| PostgreSQL CDC remains a stub on inspected `main` | `internal/connectors/native/postgres/cdc.go:10-29` |
| Approved branch implements pglogrepl CDC | `aeafb4ff0:internal/connectors/native/postgres/cdc.go:18-269`; `cdc_lifecycle.go:21-295` |
| Approved branch dependency pins | `aeafb4ff0:go.mod:8-9` |
| Shared database live harness | [dbtest maintainer guide](../../internal/connectors/native/dbtest/README.md) |
| Shared fail-closed TLS vocabulary | `internal/connectors/native/sqltls/sqltls.go:1-8,20-67,69-133` |
| Acknowledged-before-durable defect class | `/Users/karthiksivadas/karthik-agent-workspace/data/learnings.md:6-38` |
| Existing ETL system-column enrichment | `internal/app/app.go:1167-1171`; `internal/app/local_warehouse.go:189-193` |
| Raw warehouse lineage envelope | `internal/app/local_warehouse.go:39-48` |
| Delete normalization, including Airbyte CDC | `internal/app/local_warehouse.go:520-538` |
| Connection-scoped warehouse owner identity | `internal/app/local_warehouse.go:549-554`; `internal/warehouse/layout.go:128-144,285-309` |
| Fail-closed owner and table assertion | `internal/warehouse/layout.go:344-424` |

### Ruby prior-art evidence index

Root inspected: `/Users/karthiksivadas/Development/polymetrics/ruby_connectors`.

| Component | Evidence |
|---|---|
| Shared abstraction | `lib/ruby_connectors/core/base_connector.rb:5-28` |
| DuckDB client wiring gap | `lib/ruby_connectors/duckdb_connector/client.rb:5-31` |
| Closed connection alternatives | `lib/ruby_connectors/duckdb_connector/connection_specification.json:1-38` |
| Capability/mode declarations | `lib/ruby_connectors/duckdb_connector/metadata.json:1-19` |
| Writer public entry point | `lib/ruby_connectors/duckdb_connector/writer.rb:13-35` |
| Identifier and PK validation | `writer.rb:39-83` |
| Schema/table/PK creation | `writer.rb:85-112` |
| Transactional appender | `writer.rb:123-157` |
| Arbitrary type uppercasing | `writer.rb:180-188` |
| Temp stage and upsert; no-key append fallback | `writer.rb:194-250` |
| Unordered limit/offset read | `lib/ruby_connectors/duckdb_connector/reader.rb:12-45,75-90` |
| Query validator fail-open assumption | `lib/ruby_connectors/duckdb_connector/query_validator.rb:16-81` |
| Default-to-VARCHAR mapping | `lib/ruby_connectors/duckdb_connector/mapping.json:1-18` |
| Tenant fail-open fallback | `/Users/karthiksivadas/Development/polymetrics/docs/multi-tenant-database/architecture.md:488-507`; `solution.md:163-174` |
| Provision status and interpolated provisioning SQL | `solution.md:1119-1183` |

## Final recommendation

Approve the architecture and start implementation as one issue-first GSD phase whose first vertical
slice is PostgreSQL write plus PostgreSQL→warehouse→a PostgreSQL managed table. Keep the
configuration closed, derive the target from the same opaque connection ownership as the warehouse,
refuse every unowned/foreign table, keep the write lifecycle inside reverse ETL, and return a
durability receipt only after confirmed commit. Keep phase-one append explicitly at-least-once after
an unknown commit, per captain ruling. Then use MySQL as the test of whether the framework is truly
shared before adding MariaDB, SQL Server, or Oracle.

Land the approved PostgreSQL CDC lane in parallel after the focused hardening below. Its pglogrepl,
slot, source-identity, and resume work is real and reusable; the database framework should supply the
transaction sink, durable warehouse spool, managed-record enrichment, and target writer around it,
not defer it and not copy it.

The old Ruby code proves the product shape—connection, catalog, reader, writer, mapping, staging,
keys, and modes—was directionally right. The Go design must add what it lacked: typed contracts,
deterministic ordering, lossless mappings, exact target identity, shared conformance, irreversible
write approval, and acknowledgements that are only issued after the fact they claim is durable.

## Correction addendum A — approved PostgreSQL CDC branch

### Corrected fact and branch state

The original recommendation to defer PostgreSQL CDC and the `pglogrepl` decision was wrong. The
captain has approved the dependency, and the existing lane is substantial implementation work, not
a stub proposal. I inspected it read-only at
`aeafb4ff0f403cff5aa4498265c177e57bf654fd`:

```text
git rev-list --left-right --count HEAD...aeafb4ff0
  → 13  2

git merge-base HEAD aeafb4ff0
  → cb1d3c45fe08afaf51cb684284669d588e7b1d30

git show aeafb4ff0:go.mod
  → github.com/jackc/pglogrepl v0.0.0-20260401131349-e37c41485510
  → github.com/jackc/pgx/v5 v5.10.0
```

Its two ahead commits add the logical-replication reader/lifecycle, live test, dependency, capability
declaration, and GSD evidence. The branch's recorded live proof ran against PostgreSQL 12.22 and
covered insert/update/delete, restart, and slot teardown
(`aeafb4ff0:.planning/phases/cli-postgres-cdc-logical-replication-r1/VERIFICATION.md:5-24`). That work
should be preserved.

### Verdict

**It fits as the PostgreSQL protocol and slot-lifecycle driver, but it does not fit unchanged as the
framework's durable database-to-database changefeed boundary.** Do not rewrite pglogrepl, source
identification, slot inspection, resume envelopes, or the receive/feedback loop. Restructure the
consumer boundary, relation/schema fidelity, and bootstrap contract around them.

There are two distinct gates:

1. **CDC source merge gate:** rebase safely; prevent cross-relation emission; make delivery identity
   truthful; prove the durability order with a real durable test sink; use the current database
   harness; and stop lossy decoder coercion.
2. **Database-to-database CDC gate:** add a transaction-scoped sink, gap-free snapshot/changefeed
   bootstrap, schema/publication generation checks, durable warehouse transaction framing, managed
   record enrichment, and the approved managed-table writer. This gate need not block landing the
   corrected source implementation, but the legacy callback API must not be used as the final
   database-sync seam.

### Component-by-component fit

| Branch component | Verdict | Direction |
|---|---|---|
| `go.mod` pglogrepl/pgx pins | **Keep.** | The dependency decision is closed. Rebase should retain main's pgx 5.10.0 and add the exact pglogrepl pseudo-version. |
| `ChangefeedExecutorDescriptor` in `cdc.go:18-35` | **Keep and strengthen.** | Exact declaration/runtime matching is the correct anti-stub admission model. Extend runtime matching to cover delivery/event schema, not only checkpoint fields. |
| `ReadCDC` validation/open/start in `cdc.go:42-102` | **Keep structure.** | Real-source requirement, structured checkpoint, replication connection, and `StartReplication` are sound. Route consumption through the transaction sink and a publication-scoped source plan. |
| Source identity in `cdc_lifecycle.go:71-100` | **Keep base identity; extend generation.** | System ID + database + fully qualified stream is the right identity. Add relation object identity, schema fingerprint, and publication membership/options fingerprint to generation/rebootstrap checks. |
| Derived slot and lifecycle in `cdc_lifecycle.go:127-230,233-295` | **Keep mechanics.** | Deterministic connector-owned name, plugin/database/active checks, retained-position inspection, persistent reuse, and explicit inactive teardown are reusable. Bind the slot plan to the inspected publication/relation generation and expose lifecycle through a framework-owned interface. |
| Replication loop in `cdc.go:145-225` | **Keep protocol loop; replace callback boundary.** | Begin/commit tracking, keepalive handling, and standby feedback only after the durable callback are directionally correct. One transaction session must own its events and candidate checkpoint. |
| Checkpoint envelope in `cdc.go:228-260` | **Keep.** | Structured source/generation/barrier/LSN/rebootstrap state is exactly the framework seam. Event identity must be distinct from this transaction checkpoint. |
| Decoder parser/cache in `cdc_decode.go:13-131,141-280` | **Keep the parser/cache structure; revise output.** | Relation caching, truncation errors, binary refusal, and omission of unchanged TOAST are useful. Preserve relation/key/type metadata and exact values in a typed event envelope, require exact tuple shape, and invalidate stale relation metadata deliberately. |
| Decoder scalar conversion in `cdc_decode.go:283-300` | **Replace.** | PostgreSQL `numeric` is converted to `float64`, invalid booleans become false, and parse failures silently become strings. That violates the framework's lossless-or-fail rule. |
| Live scenario in `cdc_integration_test.go:23-145` | **Keep scenario; replace harness and durability fake.** | Real DML/restart/cleanup is strong. Run it through `native/dbtest` and make the downstream acknowledgement depend on an actual fsynced spool or target transaction. |
| Bundle/docs/generated capability files | **Regenerate after hardening.** | `capabilities.cdc=true` may describe the corrected source capability. It must not imply gap-free initial snapshot or CDC-to-database delivery until those separate gates pass. |

### Source-merge blockers to fix in the active lane

#### 1. Scope events to the requested relation

`ReadCDC` binds `req.Stream` into `SourceIdentity` (`cdc_lifecycle.go:82-99`), but the consumer creates
an unfiltered decoder and emits every decoded relation in the publication
(`cdc.go:145-146,208-220`). The event itself has no relation field (`connectors.go:390-394`). The docs
only require that the publication *include* the selected table; they do not require it to contain
exactly one table. As written, a multi-table publication can emit another table's rows under the
requested stream identity.

Change these exact areas:

- `internal/connectors/native/postgres/cdc_lifecycle.go`: inspect the declared publication before
  slot/start, prove the selected schema/table is a member, and capture its native relation identity
  plus publication options. Reject absent or ambiguous membership.
- `internal/connectors/native/postgres/cdc_decode.go`: retain relation schema/table on each decoded
  change and return it in the typed envelope.
- `internal/connectors/native/postgres/cdc.go`: compare every DML relation to the canonical requested
  relation and emit only exact matches. Unselected relations may advance this per-stream slot but may
  never escape as records for the selected stream.
- `cdc_decode_test.go` and `cdc_integration_test.go`: create one publication containing the selected
  and a second table, mutate both, and assert exact selected-table values/counts—not only operation
  names or exit status.

Longer term, one publication/slot session should be able to route several relations instead of
opening one slot per stream and decoding the whole publication repeatedly. That is a cost
optimization behind the same event contract, not a reason to hold the safe per-stream source.

#### 2. Make event identity match the delivery claim

The bundle declares `delivery.dedupe_key: ["lsn"]`
(`defs/postgres/changefeed.json:16-20`), but the runtime calls
`decoder.decode(xlog.WALData, "")` (`cdc.go:212`), so event `State` carries no LSN. Even if populated,
one LSN is not unique for multiple changes in one source transaction.

Capture `BeginMessage.FinalLSN`/XID and a monotonically increasing transaction ordinal. The event
identity is `(transaction_final_lsn, ordinal)`; the **checkpoint** remains the commit transaction-end
LSN. Change `changefeed.json` and executor admission to declare and match those different facts.
`ChangefeedExecutorDescriptor` currently matches status/mechanism/executor/checkpoint only
(`connectors.go:631-645` on the branch); include delivery ordering, duplicates, deletes, dedupe key,
and event-schema version so metadata cannot claim semantics the executor omits.

#### 3. Stop lossy and type-unstable decoding

Change `decodeTextValue` to return `(value, error)` and consume OID **and typmod**. Preserve relation
column flags/replica identity currently discarded at `cdc_decode.go:233-241`. At minimum:

- parse booleans strictly; invalid text is an error, not false;
- preserve `numeric` exactly (canonical decimal/pgtype value or exact lexical form plus logical type),
  never `float64`;
- fail a declared numeric/integer/float parse rather than changing the column's runtime type to
  string;
- retain unknown OID as `OpaqueNative(postgres, oid, typmod)` and stop database delivery at type-plan
  compilation unless a declared lossless mapping exists;
- retain old/key tuples for update/delete, including key-changing updates;
- represent unchanged TOAST as explicit “not present,” distinct from null.
- require a tuple's column count to equal its relation descriptor instead of accepting a truncated
  prefix; and
- handle or explicitly refuse every pgoutput v1 message the configured publication can emit. The
  decoder currently accepts only `B/C/R/I/U/D`; `Type`, `Origin`, and `Truncate` cannot fall through
  as generic parser surprises. Type metadata feeds the logical type descriptor, origin filtering is
  a declared source policy, and truncate remains a separately destructive capability.

The decoder remains target-agnostic. The shared database adapter, not pgoutput code, compiles these
source logical values into the managed target schema.

#### 4. Prove, rather than name, downstream durability

The source's call order is correct on paper: candidate committer at `cdc.go:199-202`, then standby
status at `203-206`. The interface does not couple that committer to the emitted transaction, though.
`DurableChangefeedCheckpointCommitter` is a separate callback
(`connectors.go:376-388,421-428`), and the live test creates a “durable” acknowledgement and puts the
checkpoint in a buffered in-memory channel (`cdc_integration_test.go:247-278`). That proves call
order, not downstream survival.

For the active source lane:

- extract a small replication receiver/feedback port from the concrete `*pgconn.PgConn` loop so a
  unit test can record exact calls;
- add failure-injection tests proving event/spool failure, transaction-commit failure, and
  checkpoint-persistence failure issue **no** standby feedback;
- make the live sink append a transaction commit frame to a temporary file, `fsync` it, persist the
  checkpoint, and only then return, or use the real warehouse WAL adapter when available;
- after a forced sink failure, query `pg_replication_slots.confirmed_flush_lsn` and prove it did not
  advance past the last durable transaction; and
- pass the validated start LSN into `consumeLogicalReplication` and initialize `lastDurable` from
  that value, not zero (`cdc.go:145-147`). On resume, a reply-requested keepalive can arrive before
  the first new commit; feedback and the first `DedupeWindow.Start` must remain at the durable
  checkpoint, never report LSN zero or fall back to the slot's older barrier. Add a
  keepalive-before-DML regression that proves feedback never regresses.

The public constructor for `DownstreamAcknowledgement` makes an acknowledgement well-formed; it
cannot establish that an arbitrary caller actually wrote durable bytes. Production admission must
therefore wire the framework-owned sink, and conformance must test the fact behind the value.

#### 5. Use the current shared database harness

The branch's live test is environment-only (`POLYMETRICS_INTEGRATION`) and predates the current
explicit-runtime harness contract. Rebase it onto the repository contract:

- add `//go:build databaseintegration`;
- enable and configure it through the [dbtest maintainer guide](../../internal/connectors/native/dbtest/README.md),
  which owns the explicit Docker-or-Podman runtime and direct local Unix endpoint inputs;
- configure PostgreSQL through `internal/connectors/native/dbtest.Config`; its existing
  `EngineArgs` supports `-c wal_level=logical`, `-c max_wal_senders=...`, and
  `-c max_replication_slots=...`, so no copied harness is needed;
- defer `Harness.Close` immediately, keep generated resources unique, and assert data, slot
  feedback, restart, relation filtering, and teardown.

#### 6. Rebase without discarding current `main`

The branch is 13 commits behind. A read-only `git merge-tree` shows overlapping changes in `go.mod`,
`go.sum`, `internal/connectors/connectors.go`, PostgreSQL docs/spec, `native/postgres/connection.go`,
and generated website catalogs.

Rebase rules for the lane:

- in `connectors.go`, preserve current main's dynamic-catalog, opaque-cursor, direct-read page, and
  warehouse additions; add the structured checkpoint/changefeed types rather than taking the old
  file;
- in `native/postgres/connection.go`, preserve main's shared `sqltls`, `openPool`, root CA/server-name
  handling, and unconditional password validation. Drop the branch's fixture-password relaxation;
  current fixtures already provide a secret, and `ReadCDC` rejects fixture mode before connection
  resolution;
- keep pgx v5.10.0, add pglogrepl, and regenerate `go.sum` normally;
- regenerate docs/spec/website artifacts from the merged source rather than choosing either side of
  a generated-file conflict;
- replace the branch's runtime-compose mutation as live-test evidence with `dbtest`; do not make the
  shared runtime stack the CDC test harness.

### Database-to-database integration changes (after the source merge)

#### Transaction-scoped consumer interface

Do not make the final framework trust `emit func(CDCEvent) error` plus an unrelated committer. Add an
additive strong contract in `internal/connectors/changefeed_transaction.go` (package `connectors`) so
MySQL/polling can migrate without a flag day:

```go
type ChangeOperation string // closed: insert, update, delete; truncate is separately gated

type ChangeEventEnvelope struct {
    Stream            string
    Operation         ChangeOperation
    Key, Before, After Record
    PresentColumns    []string
    TransactionID     string
    TransactionOrdinal uint64
    SchemaFingerprint string
}

type ChangefeedTransactionSink interface {
    BeginChangefeedTransaction(context.Context, ChangefeedTransactionMeta) (
        ChangefeedTransactionWriter, error)
}

type ChangefeedTransactionWriter interface {
    AppendChange(context.Context, ChangeEventEnvelope) error
    // Success means the transaction data and candidate checkpoint are durable.
    CommitChangefeedTransaction(context.Context, synccontract.CheckpointEnvelope) error
    AbortChangefeedTransaction(context.Context) error
}

type TransactionalChangefeedExecutor interface {
    ChangefeedDescriptorProvider
    ReadCDCTransactions(context.Context, CDCReadRequest, ChangefeedTransactionSink) error
}
```

Names may be shortened during implementation, but the ownership is fixed: `cdc.go` begins on a
pgoutput Begin, appends bounded events, commits on pgoutput Commit, and sends standby feedback only
after the writer's commit returns. `Abort` is deferred for every open source transaction. The
existing `ReadCDC` can be a compatibility adapter over this core; the database framework must
require `TransactionalChangefeedExecutor`.

The warehouse implementation belongs in
`internal/connectors/database/changefeed_sink.go`. It writes explicit begin/event/commit frames to
the connection-owned JSONL WAL, fsyncs the commit and checkpoint boundary, and retains that window
until managed-table delivery succeeds. A future direct target sink may stage into a database
`WriteSession`, but must preserve the same transaction/receipt contract.

#### Gap-free bootstrap is still missing

`ensureReplicationSlot` creates `NOEXPORT_SNAPSHOT` and explicitly treats the initial snapshot as a
separate sync (`cdc_lifecycle.go:207-210`). That is sufficient for “changes from this slot's
consistent point,” but it does not join an initial table snapshot to the changefeed without a gap or
duplicate window.

Add `internal/connectors/database/changefeed_bootstrap.go` and a PostgreSQL adapter that creates or
opens the slot barrier, coordinates a repeatable-read snapshot at that barrier (using the supported
PostgreSQL version-specific snapshot mechanism), finishes the snapshot into the durable warehouse
WAL, then consumes from the exact LSN. Until that conformance is green, advertise the branch as a
changefeed source starting at its slot barrier, not as a gap-free full-snapshot-plus-CDC sync.

Schema and publication changes must also be recovery events. The current generation is only
`timeline + publication name` (`cdc_lifecycle.go:94`); it does not notice a table drop/recreate,
column/type/key change, publication membership change, row/column filter change, or replica-identity
change. Fingerprint these facts at bootstrap/start and require rebootstrap or a separately approved
schema migration when they change. A PostgreSQL Truncate message must be either refused by
publication preflight or represented as a separately destructive, approved operation—never silently
treated as row deletes.

### Revised CDC route and acknowledgement boundary

The recommended phase-one continuous route is now:

```text
pglogrepl transaction
  → relation-scoped, lossless ChangeEventEnvelope(s)
  → connection-owned WAL transaction frame + fsync
  → extraction checkpoint durable
  → standby LSN feedback
  → immutable delivery window
  → plan → preview → approval
  → owned PostgreSQL table transaction + confirmed-commit durability receipt
  → delivery checkpoint
```

Acknowledging PostgreSQL after the durable local spool is correct: the spool, not the customer
target, becomes the replay boundary. It avoids retaining source WAL while a human approval is
pending and separates extraction retry from delivery retry. The spool may not be deleted until the
delivery receipt/checkpoint proves the managed target window is complete.

## Correction addendum B — captain ruling on managed database targets

### Ruling and evidence

The earlier “existing table only” recommendation was wrong. Polymetrics already enriches ETL rows
with `_polymetrics_run_id`, `_polymetrics_synced_at`, `_polymetrics_deleted`, and cursor state
(`internal/app/app.go:1167-1171`; `internal/app/local_warehouse.go:189-193`). Its raw warehouse
envelope also carries raw/run/sync/generation/extracted/loaded/cursor/key/delete facts
(`local_warehouse.go:39-48`), and delete normalization recognizes Airbyte's
`_ab_cdc_deleted_at` (`local_warehouse.go:520-538`). Dedupe, tombstones, generation replacement, and
replay need a schema Polymetrics controls.

The precedent is exact: `warehouse.LocationFor` receives
`a.state.WorkspaceID`, `conn.Source.Connector`, and opaque `conn.ID`
(`local_warehouse.go:549-554`); `Owner.SameIdentity` ignores display-name changes
(`warehouse/layout.go:128-144`); `EnsureOwnership` and `AssertOwnedTable` refuse missing, unreadable,
or foreign records (`layout.go:344-424`). The database destination must implement the same two
defenses—structural per-connection address plus an independent owner assertion—inside the target
database.

### Concrete managed-target architecture

Add these named contracts:

```text
internal/ownership/identity.go
  Identity{Workspace, Connector, Connection}
  SameIdentity(Identity, Identity) bool

internal/connectors/database/ownership.go
  TargetOwner
  ManagedTargetRef
  OwnershipState: absent | owned | foreign | unreadable | drifted
  OwnershipError

internal/connectors/database/managed_columns.go
  ManagedColumnSetV1
  EnrichManagedRecord
  reserved-prefix validation

internal/connectors/database/provisioning.go
  ManagedTableSpec
  ProvisioningPlan
  InspectManagedTarget
  AssertManagedTarget

internal/connectors/database/drivers/postgres/ownership.go
  PostgreSQL control registry + native object identity observation

internal/connectors/database/drivers/postgres/provision.go
  typed transactional DDL for owner schema/table/system columns/indices
```

`TargetOwner` projects `ownership.Identity` plus a non-identity display name. The plan derives it
from the pinned source connection that owns the warehouse input, not from a credential name. The
destination server/database, credential revision, and configuration digest remain separate target
fingerprint fields. `ReversePlan` therefore needs stable
`SourceConnectionID`/`SourceConnector` fields; legacy stored plans without them must be replanned,
not guessed from a display name.

For PostgreSQL, use a closed framework renderer to create:

- one versioned Polymetrics control schema;
- an owner/managed-table registry containing the full identity triple, logical name, derived
  physical namespace/table, managed schema version/fingerprint, native namespace/table OIDs, and
  creation time;
- one compact owner-scoped data schema (or equivalently collision-resistant structural namespace)
  per identity and a managed table per logical stream;
- a closed future extension point for a destination idempotency ledger, without creating that ledger
  in phase one.

Physical names are derived from length-delimited opaque identity tokens, never display names or
credentials. The full identity record is authoritative, so even a token collision fails
`SameIdentity`; the full logical stream/target identity is checked independently, so two logical
names that collide to one compact token are also refused. Bind the record to PostgreSQL object OIDs
so a drop/recreate at the same name is detected. Database comments/extended properties may be
defense-in-depth but are not the sole owner record.

The control schema is shared infrastructure, so its own version/shape is validated exactly. If it
exists in an unknown or modified shape, refuse. The threat boundary is accidental collision and
configuration drift, not a hostile database owner who can rewrite both table and catalog; no marker
can out-authorize the database administrator.

### Admission truth table

| Observed physical target | Owner/control assertion | Result |
|---|---|---|
| Absent | No row | Plan/preview creation; after approval create target + owner row transactionally |
| Present | Exact identity, expected OID/schema/version | Admit after drift recheck and lock |
| Present | Missing row | Refuse; never adopt or backfill ownership |
| Present | Foreign identity | Refuse with both safe owner/writer IDs; touch neither table |
| Present | Unreadable/unknown-version row | Refuse; no repair or fallback |
| Present | Exact identity but OID/schema changed | Refuse as replacement/drift; require explicit recovery/migration |
| Absent | Existing owner row | Refuse/reconcile as incomplete provisioning; never recreate over possible lost data |

Planning and preview are read-only. Execution consumes approval, opens the transaction, takes a
target-scoped lock, and reruns the assertion before any stage or DDL. PostgreSQL can create the
registry row, namespace, table, columns, constraints, and indexes in the same transaction as the
first load. MySQL/MariaDB/Oracle implicit-commit DDL requires a versioned provisioning state machine
with idempotent reconciliation; a driver cannot claim PostgreSQL atomicity it does not have.

### Managed system-column contract

Define the reserved columns once in `managed_columns.go`, not independently in app, warehouse,
PostgreSQL, and each future driver. Version 1 must give exact semantics and lossless native mappings
for at least:

```go
type ManagedRecord struct {
    Business connectors.Record
    System   ManagedColumnValuesV1
}

type ManagedRowBatch []ManagedRecord
```

| Column/fact | Purpose |
|---|---|
| `_polymetrics_raw_id` | Stable raw row/change identity within the durable source log |
| `_polymetrics_run_id` | Extraction or delivery run lineage, with one documented meaning |
| `_polymetrics_sync_id` | Stable sync lineage distinct from a transient attempt |
| `_polymetrics_generation_id` | Full-refresh generation/replacement lineage |
| `_polymetrics_extracted_at` | Source extraction observation time |
| `_polymetrics_loaded_at` | Durable warehouse/load time |
| `_polymetrics_synced_at` | Managed target synchronization time |
| `_polymetrics_cursor` and opaque cursor state | Resume/order evidence; never a substitute for a required tie-breaker |
| `_polymetrics_primary_key` | Canonical composite source-key tuple where applicable |
| `_polymetrics_deleted` | Normalized tombstone/delete state, including `_ab_cdc_deleted_at` input |
| `_polymetrics_reverse_plan_id` | Approved delivery plan lineage for the managed database sink |

The warehouse/source adapter constructs this typed envelope from its already-durable metadata; the
database writer validates and carries extraction lineage forward, then adds only delivery-owned
facts such as target sync time and reverse-plan identity. It flattens system columns into the
managed target only after business-column mapping. User records cannot set, rename, or override
system values. A business column using the reserved prefix is a planning error unless an explicit
mapping moves it to a non-reserved name. This avoids treating an untrusted map key as trusted run,
cursor, key, or tombstone state.

API reverse-ETL writers continue to omit internal metadata from provider payloads; retaining the
managed columns is a database-sink contract, not permission to leak them to APIs.

Every admitted database definition must prove exact types for the system schema before write
capability becomes true. A target that cannot represent an opaque cursor, exact timestamp, boolean
tombstone, generation, or lineage identifier does not get a fallback string cast.

### Consequences for the rest of the design

1. **Table creation moves from optional to mandatory capability.** Phase one is create-or-assert-
   owned, never existing-table-only and never adopt-if-present.
2. **The user-facing table name becomes logical.** The sealed plan shows the derived physical target;
   there is no phase-one raw `schema.table` override into customer objects.
3. **Owner identity joins every safety artifact.** Plan hash, preview digest, approval target, write
   session, durability receipt, checkpoint, errors, and audit records bind the same
   workspace/connector/connection triple.
4. **Connection IDs, not names, must survive reverse planning.** Renames remain harmless; deleted or
   legacy ambiguous identities require replan.
5. **A future receipt ledger becomes cheaper, not implicitly approved.** The owner control schema
   supplies a natural home for it, but the separate captain ruling keeps phase-one append
   at-least-once and defers that ledger.
6. **Unknown commits remain explicit.** Without the deferred ledger, append stops with
   `CommitOutcomeUnknown`; a retry may duplicate and the checkpoint does not advance.
7. **Delete and CDC semantics become implementable.** Tombstones, key-changing updates, generation,
   and replay have controlled columns and indexes; the pglogrepl adapter supplies events, while the
   framework supplies managed lineage.
8. **Schema evolution is narrower, not broader.** Polymetrics owns the table but phase one still
   refuses unplanned business-column drift. A future managed migration may add/change columns only
   through plan/preview/approval and a versioned recovery strategy.
9. **Driver cost rises where DDL is non-transactional.** MySQL/MariaDB and especially Oracle need
   durable provisioning recovery and native object-identity conformance; “driver + JSON” still does
   not mean zero engine code.
10. **Cleanup becomes destructive lifecycle work.** Deleting a connection never silently drops its
    remote tables. Teardown must be a separately previewed/approved operation that reasserts owner
    identity immediately before drop.

### Revised first implementation slice

The first PostgreSQL write slice should end only when a live test proves all of these facts:

1. an approved plan creates its owner-scoped control rows/table/system schema and loads exact data;
2. a second run with the same identity reuses the table only after `SameIdentity` + OID/schema checks;
3. two connections with the same logical table are structurally isolated;
4. an arbitrary pre-created table and missing/foreign/corrupt owner records are refused before
   mutation;
5. full refresh, append, incremental append, upsert, dedupe, and tombstone behavior preserve the
   required managed columns;
6. an injected unknown append commit returns `CommitOutcomeUnknown`, does not advance a checkpoint,
   and does not claim or perform a blind safe retry;
7. no local delivery checkpoint advances before target commit is confirmed durable;
8. PostgreSQL source → connection-owned warehouse → managed PostgreSQL target completes through
   plan → preview → approval → execute with exact source and target row assertions.
