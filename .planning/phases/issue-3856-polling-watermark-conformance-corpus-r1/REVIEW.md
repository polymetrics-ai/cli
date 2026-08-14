---
status: clean
phase: issue-3856-polling-watermark-conformance-corpus-r1
depth: standard
files_reviewed: 3
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
review_mode: inline/manual GSD fallback
reviewed_source_head: aa4d8c8a9
follow_up_review_fix_head: cd92037fe
---

# Code review — #3856 polling conformance corpus

## Execution mode

`scripts/gsd sources code-review` and
`scripts/gsd prompt code-review --auto --depth=standard --files=...` resolved
the installed workflow. The repo's canonical single-worker contract and the
Codex runtime prohibit spawning the GSD reviewer role, so the required review
was performed inline against the explicit source scope:

- `internal/connectors/engine/polling_conformance.go`
- `internal/connectors/engine/polling_conformance_test.go`
- `internal/connectors/engine/testdata/polling_watermark_conformance/v1.json`

## Review coverage

- Corpus byte digest/version, fixture completeness, defensive cloning, and the
  absence of skip/filter inputs.
- Registration/evidence fail-closed behavior and typed-nil lane protection.
- Checkpoint envelope validity, durable acknowledgement, replay positions,
  source-generation/schema recovery, and no durable high-water regression.
- Raw JSON cursor precision/coercion behavior, duplicate ordering tuple
  rejection, tombstone/history behavior, and hard-delete invisibility.
- Context, concurrency, error wrapping, serialization, and scope boundaries.

## Finding resolved during review

The original bounded-overlap reference lane copied its expected request rather
than deriving and checking it. This was a real corpus-behavior gap, recorded as
subissue #4074 and correction loop 2/5. Its focused RED/GREEN evidence is in
`TDD-LEDGER.md`; the corrected lane derives the lower timestamp request and
rejects a mismatching expectation.

`make lint` then identified S1016 in the test factory. The direct equivalent
struct conversion fixed it; this was style-only and does not increment the
substantive correction count.

## Follow-up review correction

The later review-fix commit `cd92037fe` closes two additional false-
certification gaps: registration now rejects unsafe descriptor-shaped
contracts, and persisted checkpoint provenance must match the fixture. The
focused RED/GREEN record is owned by `TDD-LEDGER.md`; this report's original
reviewed source head remains `aa4d8c8a9`.

## Stacked-base scope disposition

`intent-test-corpus-only-scope` was a documented false positive: the initial
no-mistakes review compared the child to `main` at `da7747a...`, not to the
stacked parent. At review head `261e24d4e`, the merge base with
`origin/feat/3855-polling-apply-foundations` was exactly
`fa5eef681a4b06c09519574326a22683b26bd996`; its six child commits changed
only #3856 phase evidence and the polling conformance engine/corpus paths.
No inherited Transport or CLI code was modified, split, rebased, or removed.

## Verdict

The original review was clean after the two recorded items. The follow-up
review fix above is in scope; the required independent Sol audit remains
pending for the final target.
