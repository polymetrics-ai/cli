# Context — Issue #4090: PostgreSQL native database source transport

## Goal

Expose the bounded exact PostgreSQL full snapshot from the #3976 typed-catalog
foundation as one registered `native_database` `synctransport.SourceExecutor`.
The PostgreSQL connector definition, not `App` composition, declares the exact
executor, allowed full modes, stream contract, delivery guarantees, and
conformance reference.

## Fixed decisions

- Reuse `Connector.TypedCatalog`, `database.Catalog`, the native PostgreSQL
  connection/TLS configuration, and #3974's immutable bounded read-plan
  primitives. Do not duplicate discovery or type mapping.
- The adapter is PostgreSQL-only under `internal/connectors/native/postgres/`.
  It accepts only the declared connector, concrete native executor reference,
  full modes, bounded positive batch size, declared relation, and complete
  resume identity.
- Descriptor, connector-family, executor-registration, mode, and resume
  validation must complete before opening a pool or issuing a database query.
- The full read emits typed projected records ordered by the catalog-selected
  stable key, in bounded pages, and attaches a deterministic source/schema
  checkpoint candidate. It does not implement #3858 polling/keyset resume,
  targets, warehouse materialization, CDC, or a generic SQL interface.
- The live proof uses the existing explicit Docker-or-Podman `dbtest` harness
  against PostgreSQL 16.10 and prints rows, schema/identity, and checkpoint.
- `internal/connectors/engine/bundle.go` and certification report schema
  version remain untouched.

## GSD command record

`scripts/gsd doctor`, every required `scripts/gsd sources` lookup, and
`go run ./cmd/agentcontractgen check` passed. Generated prompts for
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` were resolved. This is an inline/manual fallback because #4090
is an issue phase and the canonical single-worker contract forbids role
spawning; the fallback preserves the full TDD, verification, and review path.

## Skills loaded

`github-issue-first-delivery`, `no-mistakes`, `golang-how-to`,
`golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-testing`, `golang-context`, `golang-concurrency`, and
`golang-database`.

## CLI/docs parity

No command, flag, help topic, or user-facing command surface is introduced.
The PostgreSQL connector definition and connector-specific docs are updated
only if the new declared transport needs a truthful operator-facing statement;
runtime help/manual/website parity is otherwise not applicable.
