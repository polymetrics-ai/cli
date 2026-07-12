# Agent Trace: reviewer

## Rendered Prompt Or Prompt Reference

Read-only adversarial review against issue #325, `TEST-PLAN.md`, and the uncommitted red/green diff.

## Files Inspected

- Fixture tests/data, shell harness, partial uncommitted implementation, SPEC/test/threat contracts.

## Actions Taken

- Rejected conclusion-shaped events, forced observed!=required assumptions, circular wrapper
  inventory, ambient tool/auth exposure, weak resume I/O proof, and missing semantic mutations.
- Required deletion of partial implementation before a second red capture.
- Required exact three-HALT identity accounting, missing-field/bounds/symlink/duplicate negatives,
  hostile opt-in canaries, and non-echoing errors.

## Commands Run

- Read-only inspection only; no mutation or verification command was claimed by the reviewer.

## Findings

- All high-priority findings were incorporated into tests, fixtures, and contracts before green.

## Handoff Summary

Proceed to implementation only after the strengthened red evidence is recorded.

## Verification Evidence

Second red cycle recorded; green and broad verification remain pending.

## Unresolved Risks

- A final adversarial pass is still required on the implemented diff.
