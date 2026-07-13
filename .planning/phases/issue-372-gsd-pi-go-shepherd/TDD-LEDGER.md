# TDD Ledger

| Slice | RED evidence | GREEN evidence | Status |
| --- | --- | --- | --- |
| #373 qualification | `new-milestone` returned with M001 still in `pre-planning`; unknown `plan` became a quick task; discussion required an actual human depth answer | Supported command/query/event matrix recorded; unmanaged direct adoption rejected | GREEN |
| #374 workflow contract | Contract fixtures/tests must initially reject missing objective, output, tools, or boundaries | Pending | RED pending |
| #375 domain/store | Transition, fencing, outbox, and module-isolation tests first | Pending | RED pending |
| #376 runtime | Malformed/oversized event, timeout, cancellation, silence, and blocked-exit tests first | Pending | RED pending |
| #377 authority | Stale-head, duplicate-effect, expired-grant, and merge-denial tests first | Pending | RED pending |
| #378 telemetry | Secret/raw-payload rejection, dedupe, restart ordering, and slow-sink tests first | Pending | RED pending |
| #379 replay/canary | Incident fixtures fail against incomplete controller before fixes | Pending | RED pending |

## Rules

- A production behavior is not added before its failing test or capability probe exists.
- Focused package tests are progress evidence; full nested-module and root gates remain required.
- Failed or incomplete verification is recorded as failed, never inferred from partial success.

