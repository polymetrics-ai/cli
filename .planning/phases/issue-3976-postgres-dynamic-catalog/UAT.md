---
phase: issue-3976-postgres-dynamic-catalog
issue: 3976
status: passed-with-explicit-cdc-exclusion
source: manual-inline verify-work fallback
---

# UAT — Issue #3976

The generated `/gsd-verify-work` workflow was completed through the documented
manual-inline fallback because this issue has no numbered GSD phase and the
canonical delivery contract permits one active worker.

## Automated deliverables

- **D1 — typed pg_catalog source discovery:** passed through focused native,
  database, and engine package tests.
- **D2 — structured native/logical mapping and fail-closed rejection:** passed
  through unit tests and the PostgreSQL race suite.
- **D3 — one typed runtime source of truth:** passed through the compatibility
  projection unit test and connector-boundary gate.

## Live deliverable

- **D4 — live PostgreSQL catalog and source reads:** passed through Docker's
  explicit Colima Unix socket. The harness independently verified two
  PostgreSQL schema fixtures, discovered the seeded read table, returned the
  full IDs `1,2,3,4,5`, and returned only `3,4,5` after cursor `10`.
- **D5 — current cursor-field semantics:** observed live. No configured
  `cursor_field` ignores a stored cursor and returns the full set; an unknown
  column is rejected; a null cursor row is omitted by `>` filtering; and the
  connection-level `sequence` setting cannot serve a table whose cursor is
  `alternate_cursor`.
- **CDC:** the historical logical-replication integration test was invoked and
  skipped by its unconditional source fence. This is deliberate fail-closed
  behavior, not a passed integration result and not scope for this source-read
  child.

## Result

The dynamic catalog and legacy direct-read deliverables changed in this branch
are proven locally, including the live dbtest path. CDC execution and the
declaration-selected #3858 tuple/checkpoint path remain separately owned and
explicitly unproved.
