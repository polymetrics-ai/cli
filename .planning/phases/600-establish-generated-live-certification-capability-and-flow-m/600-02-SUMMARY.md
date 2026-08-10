---
phase: 600
plan: 02
status: complete
issue: 3984
completed: 2026-08-10
---

# Plan 600-02 summary: workflow, sync, and pair-flow baseline

## Delivered

- Added generated workflow obligations for ETL, reverse ETL, flow authoring,
  and schedule.
- Added the stable seven-sync-mode by four-warehouse-facing-primitive
  scoreboard. Every unavailable primitive or mode has a named reason; change
  capture is restricted to database read into the warehouse.
- Added all four warehouse-mediated source/destination flow classes with
  source==destination support, exact-pair resolution, independent readback
  proof, and delivery-guarantee limitations.
- Made final connector status depend on every applicable capability, workflow,
  sync-mode, and flow cell. The generated status projection visibly labels an
  incomplete connector COMMUNITY BUILD, UNCERTIFIED without blocking it.

The urgent preservation checkpoint intentionally saved the capability and flow
implementation together before the follow-up documentation/transcript pass.
They remain distinct generated artifacts with independent baseline summaries
and one joint drift gate.

## TDD evidence

- **Red:** the original matrix contract failed to compile before implementation;
  its compiler transcript is retained in TDD-LEDGER.md. The captain expanded
  workflow/sync/flow scope inside that checkpoint, so the ledger records the
  lack of a separate scope-expansion RED transcript rather than fabricating
  one.
- **Green:** the focused connectorgen certification test command validates
  stable primitive discovery, database-write stubs, CDC applicability,
  workflow incompleteness, GitHub-to-GitHub pairs, and round-trip proof
  requirements without a provider call.

## Generated baseline

There are **556 connectors** and **0 certified** connectors. All workflow,
sync-mode, and flow live-tested and complete counts are zero.

| Workflow | Applicable | Declared | Implemented | Fixture tested | Complete |
|---|---:|---:|---:|---:|---:|
| etl | 556 | 555 | 556 | 431 | 0 |
| reverse_etl | 279 | 242 | 272 | 208 | 0 |
| flow_authoring | 556 | 556 | 556 | 0 | 0 |
| schedule | 556 | 556 | 556 | 0 | 0 |

| Flow | Exact pairs | Applicable | Declared | Implemented | Complete |
|---|---:|---:|---:|---:|---:|
| api_to_api | 309,136 | 301,401 | 131,520 | 0 | 0 |
| api_to_database | 309,136 | 2,196 | 548 | 0 | 0 |
| database_to_api | 309,136 | 2,196 | 960 | 0 | 0 |
| database_to_database | 309,136 | 16 | 4 | 0 | 0 |

The 28 sync-mode cells have zero fixture-tested, live-tested, and complete
counts. For each non-CDC mode, the repeated row pattern is:

| Primitive | Applicable | Declared | Implemented |
|---|---:|---:|---:|
| api_read_into_warehouse | 549 | 548 | 549 |
| api_write_from_warehouse | 549 | 240 | 0 |
| database_read_into_warehouse | 4 | 4 | 4 |
| database_write_from_warehouse | 4 | 1 | 0 |

For change_capture, only database_read_into_warehouse is applicable (4 /
declared 0 / implemented 2); the other three primitives are named
not-applicable. The red write rows reflect the missing durable API destination
contract and missing database-write executor, rather than an omitted score.
