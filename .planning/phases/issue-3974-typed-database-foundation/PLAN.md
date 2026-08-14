# PLAN — Issue #3974: typed database connector foundation

## GSD setup and fallback

- `scripts/gsd doctor` passed in this isolated worktree.
- Resolved command sources and generated prompts for `discuss-phase`,
  `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.
- `go run ./cmd/agentcontractgen check` passed before planning.
- #3974 is an issue foundation rather than a numbered `.planning/ROADMAP.md`
  phase. The official runtime expects a numbered phase and isolated role agents,
  which this task and the canonical contract do not provide. Manual inline GSD
  execution is therefore recorded here, as allowed by the repository contract.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  `golang-database`, `golang-naming`, `golang-documentation`, `golang-lint`,
  `no-mistakes`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
  `gsd-verify-work`, and `gsd-code-review`.

## Goal

Land a safe importable `internal/connectors/database` contract that later
PostgreSQL and database-framework issues can build against without introducing
an executable database write, SQL surface, changefeed, or capability claim.

## Allowed implementation surface

- `internal/connectors/database/**` — definition loader/schema, types, catalog,
  refs, read plan, bounded resources, and driver admission registry.
- `internal/synccontract/native.go` and its focused contract tests — split shared
  executor admission from source-sync dispatch, preserving the #3810 admission
  semantics.
- `internal/connectors/engine/bundle.go` and `internal/connectors/defs/defs.go`
  — load an optional database definition only when present so the existing
  generator/load path validates it.
- `internal/connectors/defs/postgres/database.json` and the native PostgreSQL
  package's compile-time interface assertion — the reference seam only.
- `internal/warehouse/artifact.go` and the existing ownership comparison — the
  neutral layer-one artifact/identity value required by the captain's mediator
  ruling, without a new warehouse executor.
- `.planning/phases/issue-3974-typed-database-foundation/**` — lifecycle and
  red/green evidence.

## Explicit exclusions

- No generic SQL executor, arbitrary SQL value, database-shaped REST operation,
  target DDL, provisioning, write session, transactional apply, receipt,
  checkpoint acknowledgement, CDC, polling, or transport dispatcher.
- No `synccontract.Mode` replacement or alias; only its closed vocabulary is
  reused.
- No metadata capability promotion. PostgreSQL `capabilities.write` and
  `capabilities.cdc` stay false.
- No raw credential, DSN, secret, or display name in database references,
  definitions, logs, errors, or tests.
- No replacement for the existing durable warehouse inbound/outbound paths.
  F1 may type their shared artifact boundary, but it must not add a second
  writer, a receipt, or a zero-copy shortcut.

## Design

### Strict definition boundary

`database.json` is optional for the existing broad connector fleet, but when it
exists the engine bundle loader must load it through the database package. The
definition is closed at every object level, uses a versioned schema file, and is
also decoded with `DisallowUnknownFields`. The resulting `Definition` keeps its
slice and nested-type state private and returns defensive copies.

The initial declaration is deliberately narrow: driver identity/protocol/API
version, catalog qualification and identifier policy, bounded resource policy,
native-to-logical mappings, and an empty list of admitted canonical modes. A
future behavior may add a closed primitive only with its own red/green slice;
there is no transaction, DDL, or write declaration here.

### Logical/catalog/read model

`LogicalType` is constructor-only and closed: signed/unsigned integer, decimal,
float, boolean, string, binary, date, time, timestamp, UUID, JSON, array, and
opaque native. Nullability belongs to `Column`. Compatibility returns only
`exact`, `lossless`, `explicit_transform_required`, or `unsupported`; a type
plan is executable only for exact/lossless mappings. There is no string/VARCHAR
fallback.

Structured catalog/schema/relation/column/key values reject unsafe identifiers,
retain native relation identity and stable schema fingerprints, and return
defensive projections. `ReadPlan` is a typed value, not SQL: it pins the source
identity, relation fingerprint, selected columns, page limit, and deterministic
order ending in a declared unique key.

### Admission boundary

#3810's native registry stores the smaller `NativeExecutorAdmission` contract;
`Execute` additionally requires a `NativeSyncExecutor` source runner. The
database registry first resolves an exact registered `Driver` matching the
database definition, then requires the same object to implement shared native
admission for a requested native command. A declaration alone, an unknown
driver, protocol/API-version mismatch, missing evidence, or an object that is
not a native admission fails closed. This issue does not register an executable
PostgreSQL operation.

### Warehouse mediation boundary

The foundation makes the captain's N + M seam structural without expanding
into F2/F3/F4/F5 implementation. Layer one remains connector-agnostic:
`warehouse.ArtifactIdentity` names the existing workspace/connector/connection
owner triple and `ArtifactRef` adds an opaque table component. `Owner.Identity`
uses that same value, so the warehouse layout and connector contract do not
carry competing equality rules.

Layer two lives in `database`: `WarehouseInboundRef` is source → artifact and
requires the artifact to be owned by the source identity; `WarehouseOutboundRef`
is artifact → target. `DatabaseInboundCommand` and `DatabaseOutboundCommand`
are sealed values that carry one leg plus the existing native-admission
contract. A `NativeAdmittedDriver` supplies a separate #3810
`NativeExecutorAdmission` for each concrete native leg; an inbound descriptor
cannot be used to admit an outbound command. Neither value exposes a
source/target pair or executes any I/O. The existing `runWarehouseETL` remains
the single inbound implementation and the existing reverse Parquet read remains
the shared outbound primitive.

A MySQL implementation is deliberately modeled in a conformance test using
only `database`, `warehouse`, and `synccontract` contracts. In follow-on work,
the MySQL author would add its own definition, two per-leg
descriptor/evidence admissions, and native extract/apply mechanisms; no shared
layer-one or PostgreSQL file is changed.

### Resource/cancellation boundary

Pages, batches, pools, deadlines, and bind parameters all have positive safe
defaults and finite maxima bounded by framework hard limits. An explicit caller
override is rejected when outside the declaration; it is never treated as zero,
"all", or unlimited. Context cancellation is checked at definition load,
driver admission, and typed read-plan construction.

## TDD execution sequence

1. **Red:** add focused database and native-admission tests before production
   implementation. The initial test imports the absent package / contracts and
   must fail; preserve the exact output in `traces/red-run.txt`.
2. **Green:** implement constructor-validated logical types, strict definition
   loader, resource policy, refs, catalog, fingerprints, and read plan.
3. **Green:** split `synccontract.NativeExecutorAdmission`, then implement the
   driver registry and PostgreSQL compile-time reference seam.
4. **Green:** embed/load PostgreSQL `database.json`, proving engine and
   generator paths exercise the strict loader while public capabilities remain
   unchanged.
5. **Refactor/verify:** simplify package boundaries, run focused tests and the
   required non-suite gates, then use the generated verify/review prompts and
   record their manual fallback evidence.
6. **Captain amendment Red/Green:** add the neutral artifact and isolated
   inbound/outbound leg tests before the mediator production values; retain the
   failed compile output, then prove database and warehouse packages together.

## Required proof

The test matrix must prove malformed/unknown manifests and secret-like unknown
values fail without echoing values; projections cannot mutate definitions or
catalogs; type compatibility cannot coerce unknown/unsafe types to text;
identity preserves workspace/connector/connection; a cancelled context stops
work; declaration-only and incompatible driver cases fail; PostgreSQL satisfies
the intentionally non-executing driver interface; and metadata continues to
advertise neither write nor CDC. The amendment additionally proves that a
read-plan and database admission have a warehouse leg, an inbound artifact
cannot cross source identities, layer-one owner equality is shared, and a
MySQL-shaped layer-two implementation needs no PostgreSQL or layer-one change.
