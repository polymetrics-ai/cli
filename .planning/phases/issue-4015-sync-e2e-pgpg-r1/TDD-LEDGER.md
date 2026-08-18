# TDD Ledger

This is a certification-only phase. It changes no production behavior, so the TDD boundary is the missing live proof rather than a new unit test.

## Control pipeline

- Red: At task start there is no run-specific proof that a real PostgreSQL source traverses the warehouse and changes a separately queried PostgreSQL destination. Prior per-command certification cannot satisfy this assertion.
- Green: The live harness independently opened the managed target and reported 1,001 rows plus named sample `id=1001, sequence=10010, label="event-1001"`. The proof then found a separate CDC restart recovery defect, recorded in `VERIFICATION.md`.
- Refactor: Not applicable; product-code fixes are explicitly out of scope. Evidence will be reduced to exact reproducible commands, assertions, and cleanup state.

## Incremental replay

- Red: Before the second run, the task has no run-specific evidence whether `incremental_upsert` duplicates, skips, or updates acknowledged rows.
- Green: The second approved `incremental_upsert` run completed with `records_read=0` and `records_loaded=0`; the independently queried target remained at 1,001 rows and the named sample was byte-for-byte unchanged. The observed checkpoint skip matches the declared incremental polling contract.
- Refactor: Not applicable; preserve the declared mode and report observed semantics exactly.
