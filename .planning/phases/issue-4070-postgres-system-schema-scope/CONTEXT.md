# Issue #4070 — PostgreSQL system-schema scope

**Date:** 2026-08-12
**Issue:** #4070 `fix(postgres): reject system-owned schemas before catalog discovery`
**Parents:** #3976 → #3972 → #4015
**Branch:** `fix/4070-postgres-system-schema-scope`
**Required PR base:** `feat/3976-postgres-dynamic-catalog`

## Manual GSD fallback

`scripts/gsd prompt discuss-phase issue-4070-postgres-system-schema-scope --auto`
and `scripts/gsd prompt plan-phase issue-4070-postgres-system-schema-scope
--skip-research --tdd --auto` were resolved. `gsd-sdk query init.phase-op
issue-4070-postgres-system-schema-scope` returned `phase_found: false`: this is
an issue-specific correction rather than a numbered product-roadmap phase.

The repository contract permits an inline/manual fallback when the runtime
cannot provide a compatible phase. This directory is that explicit fallback.
The issue body, the promoted Sol disposition report, and the ship handoff lock
the product decisions, so `--auto` selects the sole in-scope behavior without
asking a human to re-decide it. No subagents are used in this isolated task.

## Domain

The native PostgreSQL source discovers a typed catalog dynamically from exactly
one configured schema. That discovery must remain a read-only, live-production
path for allowed user schemas, but must fail before creating a pool when the
configured schema names a PostgreSQL-owned namespace.

## Locked decisions

1. Reject exactly `pg_catalog`, `information_schema`, `pg_toast`, every
   `pg_toast_*`, and every `pg_temp_*` schema at the typed-catalog boundary,
   after basic identifier validation and before resource/pool creation.
2. Return a named, identifier-free typed scope error. Do not reveal the
   configured schema, connection details, or credentials in that error.
3. Leave valid, exact user schemas dynamic. The implementation must not replace
   them with connector JSON, fixture rows, or a static table/column list.
4. Keep `mode=fixture` isolated as test-only behavior; the live guard applies
   to the production `TypedCatalog`/legacy `Catalog` path.
5. Preserve the #3976 pipeline fixes already carried by `49a9386`; do not edit,
   synchronize, respond to, or advance its parked 5/5 no-mistakes run.
6. Prove the change both with a committed RED/GREEN unit path and a new,
   loopback-only PostgreSQL fixture exercised through a freshly built `pm
   catalog refresh` production registry path. Use PostgreSQL queries as the
   independent oracle and clean every run-owned resource idempotently.
7. Do not add destination writes, Parquet work, CDC behavior, outbound delivery,
   sync-mode apply, transport registration, or generic certification wiring.

## Canonical references

- `.agents/agentic-delivery/canonical/delivery-contract.json` — issue-first
  lifecycle and no-mistakes child command.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — PR linkage,
  GSD/TDD, and review evidence requirements.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — manual fallback
  and official command resolution.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` —
  documentation parity checklist.
- `internal/connectors/native/postgres/typed_catalog.go` — production typed
  discovery boundary.
- `internal/connectors/native/postgres/cataloger.go` — legacy catalog
  projection that routes live callers through `TypedCatalog`.
- `internal/connectors/native/postgres/typed_catalog_test.go` — focused unit
  boundary coverage.
- `internal/connectors/native/postgres/dynamic_catalog_integration_test.go` —
  isolated PostgreSQL integration harness and independent metadata oracle.
- `internal/connectors/defs/postgres/docs.md` — authoritative connector
  documentation.
- `docs/migration/HANDOFF-CODEX.md` and `docs/migration/conventions.md` —
  connector-authoring canon.

The retained causal evidence was read from the task workspace at
`data/cli-pg-4065-post5-real-db-disposition-r1/report.md`. It is intentionally
not copied into this repository: the report contains its own safe, redacted
real-database evidence and remains the audit record for #4070.

## Code context

- `Connector.TypedCatalog` resolves configuration, rejects fixture mode, checks
  the schema identifier, validates resource policy, derives a bounded operation
  context, then calls `openTypedCatalogPool`.
- The three system-catalog SQL queries are parameterized by the configured
  schema and share executable-read authorization. They are not static metadata.
- `Catalog` is a one-way live projection of `TypedCatalog`; fixture streams are
  confined to `mode=fixture`.
- Existing integration coverage already creates two materially different user
  schemas, tests unsupported and empty shapes, and compares live results to
  `information_schema`.

## Scope fences

The correction owns only pre-connection system-schema rejection, focused
coverage, accurate PostgreSQL connector documentation, and evidence required
to ship that correction on the required stacked child PR. Any inability to
reach the production registry would be a #4015 wiring matter, not a reason to
invent a catalog behavior change.
