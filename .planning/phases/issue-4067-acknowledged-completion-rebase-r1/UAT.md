---
phase: issue-4067-acknowledged-completion-rebase-r1
status: r3_local_automated_acceptance_complete_pending_review
mode: coverage-aware automated acceptance under documented inline/manual GSD fallback
updated: 2026-08-12
---

# #4067 R2 automated acceptance record

The issue has deterministic persisted-state acceptance criteria, so the named-phase manual GSD fallback uses the real JSON-store tests below rather than fabricating a human product-judgment session. This is local behavioral acceptance evidence, not certification, external CI, or merge readiness.

| Acceptance check | Expected result | Automated result |
|---|---|---|
| Post-ack cancellation after an unrelated revision | Each of seven sync modes applies once, preserves the acknowledged stream and unrelated stream/checkpoint/run, returns the durable failed run, and preserves `context.Canceled`. | Pass — `TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes`. |
| Representative post-ack source error | One apply, preserved winner/unrelated state, matching failed run, and source sentinel stays detectable. | Pass — `TestRunETLTransportAcknowledgedFailurePreservesSourceError`. |
| Failure finalizer guard | Changed/removed stream and terminal/removed exact run fail closed: zero run, latest state unchanged, original source error plus typed conflict. | Pass — `TestRunETLTransportAcknowledgedFailureFailsClosedWhenTargetChanges`. |
| Missing target completion contract | Each of seven modes returns zero, applies once, preserves reopened state, and exposes `errStateRevisionConflict`. | Pass — `TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes`. |
| Persistence truth and prior boundary | Definite/non-committed, committed, and indeterminate failure outcomes are truthful; #4046/R7/R8 regressions remain green. | Pass — outcome test, #4046/R7/R8 selectors, and focused `-race` selector. |

## Summary

total: 5
passed: 5
issues: 0
pending: 0
skipped: 0
blocked: 0

No live provider, credential, warehouse, container, external service, or human merge judgment was exercised. Fresh no-mistakes delivery, exact-head CI, and the required independent Sol audit remain outside this local acceptance result.

## R3 automated acceptance record

| Acceptance check | Expected result | Automated result |
|---|---|---|
| Two-page typed conflict in all seven modes | Page one stays acknowledged; a real winner advances the target; loser page two applies once and loses CAS; loser returns/reopens failed with typed conflict. | Pass — C12 selector, normal `-count=3`, and race `-count=3`; winner/unrelated preservation and no replay are direct assertions. |
| Existing #4046 and r2 paths | Typed-conflict-only exception, ordinary acknowledged-error guard, R7/R8, cancellation, target guards, and persistence truth stay unchanged. | Pass — focused transport regression selector and full `internal/app`. |
| Existing GitHub connector proof | Real binary builds; current definitions/hooks/preflight/inspection and bounded harness remain executable. | Pass — focused connector/CLI tests and real binary inspection. |
| Production registration / credentialed smoke | Do not imply unsupported roles, certification, or a provider result. | Explicit limitation — GitHub/PostgreSQL transport roles are unsupported; no approved credential/channel was supplied, so no smoke was attempted. |

No human merge judgment, provider mutation, credential disclosure, warehouse mutation, container, external certification, push, PR action, or CI result was exercised.
