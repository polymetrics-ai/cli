# Phase 601: issue-3754-shared-rate-coordinator - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in `601-CONTEXT.md` — this log preserves the alternatives considered.

**Date:** 2026-08-11
**Phase:** 601-issue-3754-shared-rate-coordinator
**Areas discussed:** backend mode, UDS ownership, atomic admission, crash/deadline semantics, TDD evidence

---

## Backend mode

| Option | Description | Selected |
|---|---|---|
| Process-local only | Retain the existing one-process registry with no cross-process protection. | |
| Closed local/shared selection | Preserve local behavior and make `require_shared` fail closed without a ready injected coordinator. | ✓ |
| External service | Require a durable Redis/SQLite/other coordinator. | |

**User's choice:** Closed local/shared selection.
**Notes:** The user explicitly excludes external dependencies and requires no fallback under `require-shared`.

---

## Shared coordinator lifecycle

| Option | Description | Selected |
|---|---|---|
| Durable daemon/state | Reconstruct shared budget truth after a restart. | |
| Run-owned UDS owner | Use a short-lived same-host owner, bounded closed protocol, permissions, cleanup, and a fresh epoch per run. | ✓ |
| Cross-host service | Coordinate through Redis/Dragonfly or another network service. | |

**User's choice:** Run-owned UDS owner.
**Notes:** The owner is ephemeral; crash/restart must fail closed rather than claim recovered provider truth.

---

## Reservation semantics

| Option | Description | Selected |
|---|---|---|
| Sequential policy admission | Reserve each matching policy in turn. | |
| Atomic batch plus lease | Decide all matching consumptive policies and one in-flight lease together; a block consumes nothing. | ✓ |
| Best-effort refund | Reserve sequentially and attempt rollbacks on a later block. | |

**User's choice:** Atomic batch plus lease.
**Notes:** Finish is idempotent by opaque lease ID; expiry releases only concurrency and never refunds an uncertain consumption.

---

## Failure and deadline behavior

| Option | Description | Selected |
|---|---|---|
| Fallback to process-local | Continue with per-process protection if the owner is unavailable. | |
| Fail closed | Return a typed unattempted refusal; reject old-epoch clients; refuse waits that cannot meet the caller deadline. | ✓ |
| Send optimistically | Allow the transport call and reconcile only from a response. | |

**User's choice:** Fail closed.
**Notes:** No transport request may start after missing/lost shared coordination or a too-short deadline.

---

## TDD evidence

| Option | Description | Selected |
|---|---|---|
| Mocked metadata tests | Assert configuration only. | |
| Actual local multiprocess tests | Use eight real helper processes against a UDS owner plus targeted/race/cleanup checks. | ✓ |
| Live provider tests | Use credentials and provider traffic. | |

**User's choice:** Actual local multiprocess tests.
**Notes:** The user supplied the six mandatory RED test names and exact shared-three/local-eight control outcome. No live provider credentials are in scope.

---

## the agent's Discretion

- Internal protocol encoding and deterministic test implementation, constrained by the closed, bounded, opaque contract.

## Deferred Ideas

- Operator-visible wait/refusal events and public CLI/docs parity — #3755.
- GitHub GraphQL/REST policies and response-cost integration — #3990.
- Bounded 665-case certification executor and live evidence — #3993/#3758.
