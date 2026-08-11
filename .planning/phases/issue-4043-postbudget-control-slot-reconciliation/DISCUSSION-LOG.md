# Issue #4043: Post-budget control-slot reconciliation - Discussion Log

> Audit trail only. Decisions are captured in CONTEXT.md.

**Date:** 2026-08-11
**Mode:** GSD discuss-phase --auto, executed through the documented inline
single-worker fallback for a named issue phase.

## Temp cleanup durability

| Option | Description | Selected |
|---|---|---|
| Best-effort cleanup | Release after a pre-rename failure regardless of temp outcome. | |
| Durable cleanup transition | Retain Temporary and the slot until close/remove/discards sync reconcile. | ✓ |

**Auto-selected decision:** durable cleanup transition, required by the #4043
acceptance addendum and audit reproduction.

## Recovered reservation coverage

| Option | Description | Selected |
|---|---|---|
| Clear poison from visible controls | Existing behavior; permits an unreserved retained entry. | |
| Reconcile every retained entry | Reserve or retain poison before any delivery-capable transition. | ✓ |

**Auto-selected decision:** full entry-to-control coverage, required by the
audit and accepted issue addendum.

## Delivery boundary

| Option | Description | Selected |
|---|---|---|
| Preserve current receiver boundary | Check reservation before receiver and retain receipt semantics. | ✓ |
| Alter receipt or acknowledgement formats | Broaden scope beyond the defect. | |

**Auto-selected decision:** preserve immutable receipt-before-acknowledgement;
block only ineligible work before receiver invocation.

## Deferred Ideas

None. The issue contract excludes protocol, PostgreSQL service, CLI, provider,
credential, target, and warehouse work.
