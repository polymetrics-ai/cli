# TDD Ledger

| Slice | RED evidence | GREEN evidence | Status |
| --- | --- | --- | --- |
| #373 qualification | `new-milestone` returned with M001 still in `pre-planning`; unknown `plan` became a quick task; first actual Sol session exposed thinking `off` | Governed query works; settings mismatch now fails admission; actual Sol/high validator dispatch remains required | PARTIAL |
| #374 workflow contract | Missing four-field dispatch, forbidden typed tool, unsafe scope, and overlong handoff tests failed first | Project preference/hooks, agents, skill, contracts, schemas, and negative tests pass | GREEN |
| #375 domain/store | Transition, fencing, outbox, and module-isolation tests failed before packages existed | Typed domain, SQLite/WAL grant/lease/outbox store, and root isolation pass | GREEN |
| #376 runtime | Missing runner plus malformed/oversized event, silent-process, blocked-exit, and premature-success tests | Supervised process tree, bounded events, heartbeat, query, and early-exit reconciliation pass; fire-and-forget UI misclassification found and fixed by canary | GREEN |
| #377 authority | Stale-head/model/expiry, duplicate effect, missing grant, and merge-denial tests | Exact-head attestation/recheck and grant-gated idempotent outbox pass; credentialed publisher remains disabled | PARTIAL |
| #378 telemetry | Raw payload, duplicate, and torn-tail tests | Fsynced normalized JSONL spool with recovery and 0600 files passes; analytics exporters remain pending | PARTIAL |
| #379 replay/canary | Named failure-class guards failed against absent controller | Core incident guard matrix and intake canary pass; issue-to-draft-PR canary and legacy cutover remain pending | PARTIAL |

## Rules

- A production behavior is not added before its failing test or capability probe exists.
- Focused package tests are progress evidence; full nested-module and root gates remain required.
- Failed or incomplete verification is recorded as failed, never inferred from partial success.
