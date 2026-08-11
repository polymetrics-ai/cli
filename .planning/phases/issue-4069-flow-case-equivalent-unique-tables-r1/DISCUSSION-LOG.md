# #4069 — Discussion log

> Audit trail only. `CONTEXT.md` is the execution input.

**Command:** `scripts/gsd prompt discuss-phase issue-4069-flow-case-equivalent-unique-tables-r1 --auto`
**Execution:** inline/manual, because the issue phase is not registered as an
active numbered ROADMAP phase and GSD-role delegation is forbidden.

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Ownership | Reopen #4066, create a duplicate, or make one fresh child | Use the captain-authorized fresh #4069 child; #4066 remains the exhausted 5 / 5 contract owner. |
| Collision handling | Filter SQL text, globally remove aliases, or extend the snapshot policy | Extend the existing immutable resolver-snapshot view policy only. |
| Regression evidence | Mock DuckDB metadata or use local Parquet/DuckDB state | Use real structural local ownership and Parquet files through the production query/flow boundary. |
| Generic behavior | Accept a duplicate-view failure or remove generic SQL | Preserve unrelated generic `SELECT 1` and existing generic alias behavior. |
| Flow error | Surface DuckDB catalog text or preserve domain error | Return `*warehouse.AmbiguousTableError` so the flow engine adds the existing manifest remedy. |
| Delivery order | Run all suites now or reserve heavy work for the CPU lane | Run RED/GREEN and short focused selectors now; pause after the targeted correction commit. |

No product choice remains open: the audit disposition and parent contract fix
the required behavior.
