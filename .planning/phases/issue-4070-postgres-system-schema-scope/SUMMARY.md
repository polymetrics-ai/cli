---
phase: issue-4070-postgres-system-schema-scope
issue: 4070
status: locally_complete_paused_before_no_mistakes
base_candidate: 49a9386d2c629e53594c6bba1dd9a74a05b3bff5
branch: fix/4070-postgres-system-schema-scope
required_pr_base: feat/3976-postgres-dynamic-catalog
key_files:
  created:
    - .planning/phases/issue-4070-postgres-system-schema-scope/SUMMARY.md
    - .planning/phases/issue-4070-postgres-system-schema-scope/VERIFICATION.md
    - .planning/phases/issue-4070-postgres-system-schema-scope/REVIEW.md
    - .planning/phases/issue-4070-postgres-system-schema-scope/UAT.md
  modified:
    - internal/connectors/native/postgres/typed_catalog.go
    - internal/connectors/native/postgres/typed_catalog_test.go
    - internal/connectors/native/postgres/dynamic_catalog_integration_test.go
    - internal/connectors/defs/postgres/docs.md
    - website/data/connectors.generated.json
    - website/lib/connectors.catalog.data.generated.json
---

# #4070 summary — PostgreSQL system-owned catalog scope

## Outcome

The fresh #4070 correction rejects the required PostgreSQL-owned configured
schema names before resource creation or a pool connection. It preserves
dynamic discovery for allowed user/application schemas and leaves the
unsupported-shape behavior fail-closed.

The causal defect was reproduced first from the preserved #3976 pipeline
candidate `49a9386d2c629e53594c6bba1dd9a74a05b3bff5`: a live temporary table in
PostgreSQL's physical `pg_temp_N` schema appeared in an actual freshly built
`pm catalog refresh` result. The Green implementation is the narrow guard in
`TypedCatalog`, which both the production registry path and direct typed/legacy
connector paths traverse.

## Custody and branch identity

- Old #3976 remote PR #4065 head remains
  `24d0055f5c9421f0bd18d0d33313a3917210ba84`.
- The preserved old #3976 local worktree remains clean at
  `46ee78620dfecb091090e40fc8986025f073d6a9`.
- The immutable, parked #3976 no-mistakes pipeline head is
  `49a9386d2c629e53594c6bba1dd9a74a05b3bff5`; its correction budget remains
  exhausted at 5/5. This work did not respond to, sync, alter, or recover that
  run.
- This branch was created by fetching that exact preserved commit graph into
  `fix/4070-postgres-system-schema-scope`; it did not move any existing ref.

## Delivered change

- `ErrSystemCatalogSchema` is a stable, identifier-free sentinel.
- After configured-schema identifier validation and before definition/resource
  validation, operation-context creation, or `openTypedCatalogPool`, the guard
  rejects exact `pg_catalog`, `information_schema`, and `pg_toast`, plus
  `pg_toast_*` and `pg_temp_*`.
- The existing compatibility `Catalog` path inherits the result because it
  calls `TypedCatalog`; no static connector metadata or fixture stream is used
  for a live catalog.
- A closed-loopback unit test proves all exact/prefix cases fail before any
  transport attempt. The opt-in integration test retains a held physical
  `pg_temp_N` table and checks both catalog entry points.
- PostgreSQL connector documentation and its repository-owned generated website
  data state the same scope boundary.

## Evidence and state

- Committed RED: `e23a945` and `d890902`; the candidate attempted a pool/
  transport connection for every reserved namespace, and the real PM path
  discovered a held temporary table.
- Committed Green implementation: `a571861`.
- Fresh loopback PostgreSQL Green proof, independent server-catalog oracle,
  two dynamic allowed schemas, unsupported enum behavior, and idempotent
  cleanup are in `traces/green-live-boundary.md` and `traces/cleanup.md`.
- Manual GSD verification and deep source review are complete in
  `VERIFICATION.md` and `REVIEW.md`; official formal GSD phase initialization
  is unavailable because #4070 is not a roadmap phase, so this directory is
  the AGENTS.md-required manual fallback.

## Deliberate pause

At Firstmate's request, this local evidence commit is the stopping point. No
new #4070 no-mistakes run has started and no pipeline action was taken while
GitHub run `01KZSJG7P7QREZSZ7N52TPBA5F` finishes. When explicitly resumed, run
the fresh #4070 no-mistakes pipeline without `--yes`, own every response to a
terminal result, then push and create only the requested draft child PR against
`feat/3976-postgres-dynamic-catalog`.
