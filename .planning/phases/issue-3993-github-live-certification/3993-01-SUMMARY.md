---
phase: 3993-github-live-certification
plan: 01
subsystem: connector-certification
tags: [github, certification, provenance, tdd, duckdb, parquet]
requires: []
provides:
  - Closed execution-model provenance for the external GitHub proof runner.
  - Current-SHA built-binary assessment with an explicit safe live blocker.
affects: [github-certification, issue-3993, pr-4061]
tech-stack:
  added: []
  patterns:
    - External per-operation process reports cannot claim in-process certification.
    - Missing approved credential or boundary is needs-decision, never a synthetic pass.
key-files:
  created:
    - .planning/phases/issue-3993-github-live-certification/3993-01-SUMMARY.md
  modified:
    - scripts/github-live-proof-sweep.mjs
    - scripts/tests/github-live-proof-sweep.test.mjs
    - .planning/phases/issue-3993-github-live-certification/TDD-LEDGER.md
    - .planning/phases/issue-3993-github-live-certification/VERIFICATION.md
key-decisions:
  - "Credentialed current-SHA evidence must identify built_pm_in_process; the external Node runner fails closed before provider dispatch."
  - "Without a full-parity GitHub App credential and immutable Polymetrics-Cert boundary, live closure is needs-decision."
requirements-completed: []
duration: 17min
completed: 2026-08-11
---

# Phase 3993 Plan 01: Current-SHA provenance and safety recovery Summary

**The external GitHub proof runner now fails closed on false one-process provenance, while the real in-process certification route is measured honestly and blocked safely rather than promoted.**

## Performance

- **Duration:** current recovery session on 2026-08-11
- **Started:** 2026-08-11T20:33:39+05:30
- **Completed:** 2026-08-11T15:20:41Z
- **Tasks:** 4/4 execution tasks completed; Task 4 ended in its planned `needs-decision` safety outcome.
- **Files modified:** 7 implementation/evidence files plus this summary.

## Accomplishments

- Added a fresh lineage 1/5 RED/GREEN regression that rejects an external
  per-operation `pm` runner from credentialed-live evidence.
- Closed the report schema around `built_pm_in_process` and
  `external_pm_per_operation`, fixed the unused import, and retained exact
  structural configuration redaction.
- Built and hashed the current `pm`, measured the in-process legacy harness
  boundaries, and distinguished it from #3993 full-surface certification.
- Recorded no credential/provider use and a `needs-decision` block when the
  required App credential and immutable run-owned boundary were unavailable.

## Task Commits

1. **Task 1: Capture false provenance and scan regressions (RED)** — `26af6995d` (`test`)
2. **Task 2: Fail closed and repair scanner findings (GREEN)** — `aeaec4dd1` (`fix`)
3. **Tasks 3–4: Built-binary assessment and live-safety result** — `42da61459` (`docs`)

## Files Created/Modified

- `scripts/github-live-proof-sweep.mjs` — validates closed execution-model
  provenance and refuses external credentialed-live dispatch.
- `scripts/tests/github-live-proof-sweep.test.mjs` — RED/GREEN proof and
  structural redaction coverage.
- `TDD-LEDGER.md`, `RUN-STATE.json`, and `VERIFICATION.md` — durable fresh
  lineage, gate output, hashes, and blocker evidence.

## Decisions Made

- A barrier label is launch topology, not rate-coordinator evidence.
- The built `pm connectors certify` route is one-process infrastructure but
  remains a legacy, partial harness; it cannot certify the full #3993 surface.
- No weaker credential, personal repository, or unverified boundary may be
  substituted for the approved App-owned live closure.

## Deviations from Plan

None — the plan specified `needs-decision` as the only correct result when the
approved credential or boundary was unavailable.

## Issues Encountered

- No approved full-parity GitHub App credential or immutable run-owned
  `Polymetrics-Cert` boundary configuration was present in this isolated
  worktree. No provider operation was attempted.
- Warehouse-to-GitHub authored action remains dependent on #3994/#4059; real
  schedule firing remains dependent on #3992. #4060 parent integration is
  still pending and was not rebased.

## User Setup Required

An authorized operator must make the existing issue-contract App credential and
immutable run-owned boundary available through an approved secret channel, then
explicitly authorize a new bounded live-validation window. Do not place a
secret in this repository or this chat.

## Next Phase Readiness

The branch is ready for GSD verification/review and a direct update of draft
PR #4061. It is not ready to claim GitHub certification or execute any provider
mutation.

## Self-Check: PASSED

The implemented source/test change, fresh RED/GREEN record, current built-binary
identity, focused gates, and explicit safe blocker all match the scoped plan;
the absent live proof is named rather than inferred.
