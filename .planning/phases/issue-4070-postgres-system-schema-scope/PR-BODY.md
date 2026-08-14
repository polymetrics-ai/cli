## Summary

- Reject PostgreSQL-owned catalog namespaces before pool creation while keeping
  user/application schemas dynamically discovered from the configured server.
- Cover exact `pg_catalog`, `information_schema`, `pg_toast` and prefix
  `pg_toast_*` / `pg_temp_*`, including an actual held PostgreSQL temporary
  table namespace.
- Preserve allowed dynamic catalog differences and explicit unsupported-shape
  failure; update the connector's authoritative and generated documentation.

## Issue stack

Refs #4070

Refs #3976

Refs #3972

Refs #4015

## Branch and custody

- This child branch starts from preserved pipeline head
  `49a9386d2c629e53594c6bba1dd9a74a05b3bff5`.
- Required base: `feat/3976-postgres-dynamic-catalog`.
- It does not move, respond to, or consume a sixth loop from the immutable #3976
  5/5 no-mistakes lineage.

## GSD / TDD evidence

- Fresh 0/5 #4070 ledger: `.planning/phases/issue-4070-postgres-system-schema-scope/TDD-LEDGER.md`
- Context, custody, Red/Green live-boundary traces, cleanup, verification,
  review, and acceptance records: `.planning/phases/issue-4070-postgres-system-schema-scope/`
- Formal phase initialization is unavailable because #4070 is not in ROADMAP;
  the directory records the required inline manual GSD fallback.

## Verification

- Focused PostgreSQL unit, integration-tag compilation, and race tests
- Fresh `pm catalog refresh` against a unique loopback PostgreSQL fixture plus
  independent PostgreSQL system-catalog oracle
- `go vet ./...`, build, lint, docs/generator, connector boundary/canon/runtime
  gates, release workflow check, agent contract check, and `git diff --check`

## Pending delivery gate

At Firstmate's direction, the new #4070 no-mistakes pipeline has not started
yet: GitHub run `01KZSJG7P7QREZSZ7N52TPBA5F` is finishing and the shared daemon
must remain idle for Transport's stale-gate resolution. On explicit resume, run
the new #4070 pipeline without `--yes`, own all responses to completion, then
update this record with the run outcome before pushing and opening this draft
PR. No merge, retarget, readiness transition, or full-certification claim is
intended.
