## Summary

Open the #3855 polling/apply foundations parent as a draft-only, planning-backed stacked scaffold.
This PR adds no connector, transport, PostgreSQL, or product implementation.

Refs #3855
Refs #4015
Refs #3862

## Stacked PR

- Primary issue: #3855
- Parent branch: `feat/3855-polling-apply-foundations`
- Current dependency-only base: `feat/3862-any-to-any-transport`
- Head: `feat/3855-polling-apply-foundations`
- Required state: draft
- Retarget rule: before final parent integration, retarget to
  `docs/4015-connector-release-certification` once the reviewed transport seam is present there.
  This branch must never target or merge to `main`.

The temporary base records a dependency only. It does not authorize integrating #3855 into #3862.

## Pending children

- #3856 — immutable polling-watermark conformance corpus
- #3857 — polling descriptor and transport preflight, after #3856
- #3858 — page-safe polling source executor, after #3857
- #3859 — native apply strategies, after #3857
- #3860 — polling-watermark eligibility and limitation documentation, after #3856–#3859

The required core implementation order is `#3856 -> #3857 -> (#3858 || #3859)`. #3860 is a
follow-on documentation child, not a parallel core implementation lane.

## Historical partial reuse

Merged PR #3880 / commit `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` is recorded as partial
reusable polling implementation. It does not complete or close #3856, #3857, #3858, #3859, or
#3860, and this parent does not cherry-pick or reimplement it.

## Parent orchestration

- State ledger: `.planning/phases/issue-3855-polling-apply-foundations-parent/`
- Current canonical state: `issue_map`
- Next state: `parent_draft_pr` after live draft PR creation and inspection evidence is recorded.
- Worker mode: single-worker inline/manual GSD fallback; no roles were spawned.
- Integration/merge: prohibited in this scaffold; final parent integration remains human-gated.

## Verification

- GSD adapter and canonical delivery projection passed before the seed.
- This diff is planning-only; no `cmd/` or `internal/` path is changed.
- Remaining local gates, no-mistakes, and live `gh-axi` PR validation are recorded in the phase
  ledger and will be updated before the draft PR is opened.

## Automated review

- Primary route: pending — this PR intentionally remains draft.
- Fallback route: none.
- Coverage status: pending; no review is requested while draft.
- Certification/executable behavior: not claimed.
