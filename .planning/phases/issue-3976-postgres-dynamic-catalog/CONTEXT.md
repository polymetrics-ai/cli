# Context — Issue #3976: PostgreSQL dynamic typed catalog discovery

## Fixed delivery decisions

- The one target connector is `postgres`; this is a PostgreSQL-specific native
  source/catalog adapter slice.
- #3976 already owns runtime discovery of the configured database's allowed
  schema scope, base relations, ordered columns, nullability, ordered primary
  keys, supported native/logical types, structured identities, and catalog
  fingerprints. No missing-boundary sub-issue is required.
- Discovery remains source-side and read-only. It must not introduce target
  DDL/write/evolution (#3982), Parquet/workset typing (#3980), outbound target
  delivery (#3983), CDC execution (#3977), or canonical flow/mode dispatch
  (#3987).
- The accepted scope is the configured PostgreSQL database and configured
  schema (default `public`), with ordinary base tables only. System schemas and
  views are intentionally excluded; the discovered catalog must make the
  configured database and schema identity explicit rather than flattening them
  into a table name.
- PostgreSQL-native type identity and modifiers must be retained alongside the
  typed logical type from #4034. A shape without an explicit lossless mapping
  is rejected with a typed, secret-safe error; it is never silently converted to
  `string`/`object`.
- #3977 consumes the read-plan/cursor/CDC details after this catalog contract.
  This issue can preserve metadata needed by those later boundaries but does not
  make polling or CDC executable.
- #4058 remains a merge-order blocker for parent integration: this child may be
  developed and reviewed, but it must not be integrated into #4017 before the
  corrected #4058 path is green and merged.

## Required reading completed

- `AGENTS.md`; issue-first, stacked-child, GSD adapter, required-skills, runtime
  integration, PR-template, GSD evidence, Claude review, and review-routing
  contracts.
- Parent #3972; #3976; adjacent #3975, #3980, #3982, #3977; ownership neighbors
  #3983 and #3987; shared polling dependency #3858; parent PR #4017.
- Merged typed-foundation PR #4034, merged managed-target-state PR #4057,
  merged parent advance #4064, and open correction PR #4058.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-postgres-dynamic-schema-parity-audit-r1/report.md`.

## GSD / worker fallback

`scripts/gsd doctor`, all lifecycle `sources` resolutions, and
`go run ./cmd/agentcontractgen check` passed. The adapter generated the
`discuss-phase` and `plan-phase --tdd` prompts for this issue. #3976 is not a
numbered roadmap phase and the canonical delivery contract requires one active
worker with no GSD-role delegation, so this directory records the permitted
manual inline GSD fallback.

## Skills loaded

`golang-how-to`, `golang-database`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-context`, `golang-lint`,
`github-issue-first-delivery`, and `no-mistakes`.

## Parent-head synchronization

Before the first implementation slice, the child preserved the then-current
parent-head history from #4064 in merge commit `25bda3e73`. That historical
topology no longer defines this PR's comparison target: the current target is
`integration/4015-mvp-flat-r1`, recorded in the live-proof resumption below.

## Static-surface ownership audit

- In scope (#3976): native PostgreSQL catalog queries, old coarse
  `connectors.Catalog` projection only insofar as it must be sourced from the
  typed live catalog, native/logical mapping, fingerprints, and source identity.
- #3980: catalog-to-Parquet/DuckDB physical typing and schema evolution.
- #3982: PostgreSQL target catalog, DDL, write encoding, target evolution.
- #3983: immutable workset delivery, approval, target receipts/baselines.
- #3987: warehouse-only routing and seven-mode execution/conformance.

No credentialed configuration, raw DSN, raw SQL surface, or generic database
write capability is introduced by this plan.

## 2026-08-14 live-proof resumption

- The child is now merged onto `integration/4015-mvp-flat-r1` at
  `fbd06e7d7c5c0632182e98cbb3a223ba25b19883`; its draft PR #4065 already
  targets that integration branch. The base's Docker/Podman dbtest contract
  and PostgreSQL CDC containment fence are authoritative.
- The only live test target is PostgreSQL through the explicit local Docker
  Unix socket. The test must assert live catalog metadata, a complete bounded
  full read, a cursor-advanced read, and the exact current behavior of absent,
  missing, nullable, and connection-level `cursor_field` configuration.
- `cursor_field` remains an optional connection-level field in this issue.
  Making it user-supplied and mandatory for a mode that needs it is explicitly
  deferred to the captain's separate cursor-contract issue; this slice records
  observations instead of changing that product contract.
- PostgreSQL logical-replication CDC remains deliberately fail-closed per the
  merged base's capability fence. The historical CDC integration test's
  unconditional skip is not live-read evidence and must not be re-enabled by
  this source-read/catalog child.
- The legacy scalar reader is not claimed to be #3858's tuple/checkpoint
  executor. Its live behavior is made observable here; a declaration-selected
  PostgreSQL #3858 native hook remains a separately reviewable incremental
  execution slice.

The implementation scan found only these static PostgreSQL shapes: the
existing `mode=fixture` catalog and test-only schema-creation/oracle SQL. Live
`Catalog()` now projects one `TypedCatalog()` result and has no hard-coded
table or field list. Existing CDC test setup and the unchanged unsupported
`Write` method remain outside this child; no #3980/#3982/#3983/#3987 path was
changed.
