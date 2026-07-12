# Agent Trace: reliability-observability

## Rendered Prompt Or Prompt Reference

Reliability role contract, `OBSERVABILITY.md`, `TEST-PLAN.md`, and final implementation diff.

## Files Inspected

- Replay ordering/correlation, output structs, safety exit classes, driver guard timing, harness
  snapshots, Make integration, and repository verification logs.

## Actions Taken

- Required deterministic filename/result ordering, bounded tuple search over at most 64 events,
  exact precedence for blocked human waits, and valid final-human-ready exclusions.
- Verified status/replay output is bounded and drivers emit one typed denial without log creation.

## Commands Run

- Repeated targeted/race/CLI/shell gates, aggregate target, uninterrupted full `make verify`.

## Findings

- First-occurrence matchers could be decoy-suppressed; exhaustive bounded candidates fixed the
  class rather than individual fixtures.
- Interim wait and terminal mismatch needed explicit context to avoid ambiguous policy output.

## Handoff Summary

Reliability review approves the deterministic Phase 0 oracle and fuse.

## Verification Evidence

Race, repeated package, shell side-effect, and full repository gates pass.

## Unresolved Risks

- Replay is a closed characterization oracle, not the later durable controller/state runtime.
