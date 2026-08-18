# TDD Ledger

This is a certification-only phase. It changes no production behavior, so the TDD boundary is the missing live proof rather than a new unit test.

## Control pipeline

- Red: At task start there is no run-specific proof that a real PostgreSQL source traverses the warehouse and changes a separately queried PostgreSQL destination. Prior per-command certification cannot satisfy this assertion.
- Green: Pending live `TestPMBinaryExecutesPostgresWarehousePostgres` execution plus an independently queried exact row count and named content sample.
- Refactor: Not applicable; product-code fixes are explicitly out of scope. Evidence will be reduced to exact reproducible commands, assertions, and cleanup state.

## Incremental replay

- Red: Before the second run, the task has no run-specific evidence whether `incremental_upsert` duplicates, skips, or updates acknowledged rows.
- Green: Pending a second approved run whose independently queried target count and sample content remain stable.
- Refactor: Not applicable; preserve the declared mode and report observed semantics exactly.
