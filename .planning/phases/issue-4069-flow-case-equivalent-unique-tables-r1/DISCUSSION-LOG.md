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

## Correction 1 / 5 addendum

**Command:** `scripts/gsd prompt discuss-phase issue-4069-flow-case-equivalent-unique-tables-r1`
**Execution:** inline/manual under the same no-delegation fallback; the prior
no-mistakes CI monitor reports the branch synchronized and its current help
permits post-pipeline commits on top of the preserved head.

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Same-owner policy | Permit two spellings through a new durable alias/physical-name design, reject the full project at open, or reject new inventory and fence legacy mutation | Accepted option 1: reject new local-warehouse configuration after defaults; preserve legacy state and exact proven reads; block legacy sync before mutation. |
| SQL error | Reuse one-owner `AmbiguousTableError`, expose DuckDB catalog text, or add a typed same-owner error | Add a dedicated typed error with a truthful exact-read/replacement-connection remedy; do not claim that `connection` can resolve one owner's collision. |
| Query availability | Refuse all SQL for a legacy collision or suppress only the irreducible identifier bindings | Suppress only the collision's bare/generated bindings so `SELECT 1` and unrelated tables remain executable. |
| Delivery | Create a new issue/PR or use the existing #4069 correction budget | Use existing #4069, draft #4071, and correction 1 / 5; preserve every prior and pipeline commit. |

## Authorized Website generated-data recovery addendum

**Command:** `scripts/gsd prompt plan-phase issue-4069-flow-case-equivalent-unique-tables-r1 --gaps`
**Execution:** inline/manual under the existing no-delegation fallback. The
captain-authorized strict fast-forward preserved the completed pipeline chain
at `9a5b23fe14aba16d04f55b28ff52be0a5940cb68` before this gap was planned.

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Ownership | Open a new child/correction loop, edit website source, or correct the existing aggregate | Keep correction 1 / 5 in #4069 and refresh only the aggregate derived from its already-committed source pages. |
| RED evidence | Treat stale CI as an environment warning or record the exact generator failure | Record the exact #4071 Website CI/CD and Website Data failures before local generation. |
| Delivery topology | Let no-mistakes create/adopt a PR or preserve the existing stacked draft | Run one fresh no-mistakes pass with `--skip=pr,ci`, then use natural #4071 checks as the only CI authority. |

## Correction 2 / 5 addendum — flow manual golden drift

**Command:** `scripts/gsd prompt plan-phase issue-4069-flow-case-equivalent-unique-tables-r1 --gaps`
**Execution:** inline/manual under the same no-delegation fallback. The
captain decision `[key=flow-manual-golden-drift]` fixes the exact failed CI
contract without broadening behavior or delivery topology.

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Correctness boundary | Edit generated markdown, bypass the golden, or make embedded help authoritative | Update only `flowHelp`, then regenerate derived manual/transcript/website output through checked-in generators. |
| Runtime behavior | Change flow parsing or query policy, or document the already delivered behavior | Preserve all runtime behavior; this correction is help/manual parity only. |
| CI recovery | Rerun the failed Verify workflow or push a regenerated head naturally | Do not rerun; commit a new generated head, run a fresh local pipeline with PR/CI skipped, then let #4071 trigger checks naturally. |

## Correction 3 / 5 addendum — destination-scoped legacy admission

**Command:** `scripts/gsd prompt discuss-phase
issue-4069-flow-case-equivalent-unique-tables-r1`
**Execution:** inline/manual; the #4069 phase is not a numbered ROADMAP phase
and its canonical contract forbids role delegation.

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Non-local protection | Preserve the global preflight, remove all legacy collision checks, or condition the guard on the selected destination | Restore destination-scoped admission: a non-local ETL must be unaffected by a legacy local-warehouse collision. |
| Portability | Test a connector name/warehouse literal, or ask the configured destination whether it materializes locally | Use only the existing `LocalWarehouseMaterializer`-backed abstraction. |
| Safety control | Trust the old suite, or prove the inverse behavior explicitly | Retain the existing same-connection local typed-error test beside the restored non-local regression. |
