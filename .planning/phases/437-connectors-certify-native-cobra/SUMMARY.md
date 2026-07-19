# Phase 437 Summary

Status: planned; RED pending.

## Identity

- Session: `issue-437-pi-sol-high-20260719T095145Z`
- Profile: Sol/high
- Branch: `refactor/437-connectors-certify-native-cobra`
- Exact start/base: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b` / `feat/cli-architecture-v2`
- Parent #397, umbrella #407, draft parent PR #438

## Planned delivery

Native Cobra ownership for `connectors` and nested certify modes while preserving dynamic connector legacy dispatch, current outputs/parser compatibility, certify re-entrancy and exits, bounded concurrency/cancellation/events/telemetry, and credential-value exclusion. Directly applicable connectors manual/docs/website parity only.

GSD adapter doctor/list/plan passed; `programming-loop` is unavailable, so manual universal-loop TDD is recorded. No production or test edits yet.

## Safety / handoff

Fixture/replay/local-only; no connector defs, credentials, live checks/writes, dependencies/services, PR, or review. Final worker handoff will include commits, verification, exemptions, and private-material exclusion.
