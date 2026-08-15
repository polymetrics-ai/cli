# #4070 plan — reject PostgreSQL system-owned catalog schemas

**Type:** issue-first behavior correction, manual GSD/TDD fallback
**Status:** local implementation, verification, and review complete; paused
before the new #4070 no-mistakes run at Firstmate's request
**Issue hierarchy:** #4070 → #3976 → #3972 → #4015
**Candidate:** `49a9386d2c629e53594c6bba1dd9a74a05b3bff5` on
`fix/4070-postgres-system-schema-scope`
**Required draft PR base:** `feat/3976-postgres-dynamic-catalog`

## Why manual GSD

The official prompts and adapter health checks were run. The phase initializer
cannot resolve issue-specific phase `issue-4070-postgres-system-schema-scope`
from ROADMAP.md (`phase_found: false`), so no compatible formal GSD phase can be
created. This directory records the inline fallback required by AGENTS.md:
context → TDD plan → committed RED → minimum GREEN → refactor → verify-work →
code review. The plan is intentionally single-worker; the task did not request
subagents and the current environment disallows spawning them.

## Required skills and safety posture

Loaded: `golang-how-to`, `golang-database`, `golang-error-handling`,
`golang-security`, `golang-testing`, `golang-documentation`,
`github-issue-first-delivery`, `no-mistakes`, and the five GSD workflow skills.

The new guard is an input-validation boundary. It must not interpolate schema
input, expose identifiers or credentials in the error, or use a shared
database. All real verification uses a unique loopback PostgreSQL data directory
and a redacted environment alias only.

## TDD execution plan

### 1. Commit the RED before production code

**Files:** `internal/connectors/native/postgres/typed_catalog_test.go`,
`internal/connectors/native/postgres/dynamic_catalog_integration_test.go`,
this phase's ledger/traces.

1. Add table-driven coverage requesting `pg_catalog`, `information_schema`,
   `pg_toast`, a representative `pg_toast_*`, and a representative `pg_temp_*`
   via an otherwise valid live configuration aimed at an unreachable loopback
   endpoint. Assert the desired identifier-free typed scope outcome, not a
   transport error. The current candidate must fail this test because it opens
   a pool.
2. Add a database-integration assertion that holds a real temporary table open,
   obtains its physical `pg_temp_N` namespace from PostgreSQL, and expects the
   same scope outcome from both `TypedCatalog` and legacy `Catalog`. On the
   candidate this is a behavioral RED: the temporary table is discovered.
3. Before any production edit, build a fresh `pm`, start a unique loopback
   PostgreSQL fixture, create two user catalogs plus an open temporary table,
   exercise the production registry with `pm catalog refresh`, and independently
   query PostgreSQL. Record only safe command shapes, aliases, object names,
   hashes/counts, and outcomes in `traces/red-live-boundary.md`.
4. Commit the failing test-only state with its focused failure evidence.

### 2. Minimum GREEN

**Files:** `internal/connectors/native/postgres/typed_catalog.go`, focused tests.

1. Add one named typed-catalog scope sentinel and one narrow helper.
2. Invoke it immediately after `validateIdentifier(conn.schema)` and before
   definition/resource validation, operation context creation, or
   `openTypedCatalogPool`.
3. Match the exact required names and prefixes only. Do not add a broad `pg_`
   rule, change queries, or change allowed-schema behavior.
4. Make RED coverage green, including both typed and legacy catalog routes.

### 3. Documentation and refactor

**Files:** `internal/connectors/defs/postgres/docs.md`, derived website data if
the website generator changes it, phase evidence.

1. State that live discovery is restricted to an allowed user/application
   schema and list the rejected PostgreSQL-owned namespace forms.
2. Regenerate only the repository-owned derived documentation that changes.
3. Keep comments concise and explain the boundary rather than repeating the
   helper implementation.

### 4. Verification and review

1. Run focused PostgreSQL unit/integration/race tests, `go vet`, lint,
   `go build ./cmd/pm`, documentation/generator checks, `git diff --check`,
   GSD verification, and deep code review.
2. Re-run the loopback fixture through a freshly built binary and actual
   `pm catalog refresh`; independently query PostgreSQL as the oracle. Prove
   dynamic differences between two user schemas, unsupported-shape fail-closed
   behavior, system-scope rejection before connection, and idempotent cleanup.
3. Commit all implementation/evidence before starting a fresh #4070
   no-mistakes run. While that run is active, use only `axi run`/`axi respond`
   and never edit outside its pipeline.

## Acceptance matrix

| Case | Expected shipping behavior | Proof |
| --- | --- | --- |
| Valid user schema A | Live typed catalog with server-derived tables, columns, identity, ordering, nullable flags, keys, and supported type mapping | `pm catalog refresh` + independent PostgreSQL oracle |
| Materially different user schema B | Different correct catalog/fingerprint without connector JSON or code schema edits | Same fresh binary; separate schema/oracle comparison |
| Unsupported allowed shape | Typed and legacy catalog fail closed; no partial/static substitute | Existing unsupported enum/zero-column live fixture |
| `pg_catalog`, `information_schema`, `pg_toast` | Named typed scope error before any pool | Table-driven unreachable-loopback RED/GREEN test |
| `pg_toast_*`, `pg_temp_*` | Same named typed scope error before any pool | Table-driven test plus real held temporary-table session |
| Fixture metadata | Never substitutes for a live allowed schema | Production registry run and source-path inspection |
| Cleanup | All run-owned schemas, database/data directory, process, and workspace state removed; repeated cleanup succeeds | Post-run PostgreSQL/or filesystem checks |

## Completion boundary

The result is a draft #4070 child PR, not a merge or certification claim. It
must reference #4070, #3976, #3972, and #4015; retain base
`feat/3976-postgres-dynamic-catalog`; and be ready for a new independent Sol
audit.
