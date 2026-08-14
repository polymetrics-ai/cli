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
- Accepted transport seam: #4059 is now merged into this base at
  `c67f40a5ff67a131950f3123e70527027dca8493`; the #3855 planning-only range was safely replayed
  onto that head so future #3856 work inherits it. Audited non-force bridge `72a4fc32...` carries
  the preserved ancestry only; its tree is identical to checkpoint `3b6e3e2e...`.
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

The logical core dependency graph is `#3856 -> #3857 -> (#3858 || #3859)`. The active programme
executes it more strictly as `#3856 -> #3857 -> #3858 -> #3859`. #3860 is a follow-on documentation
child, not a parallel core implementation lane.

## Historical partial reuse

Merged PR #3880 / commit `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` is recorded as partial
reusable polling implementation. It does not complete or close #3856, #3857, #3858, #3859, or
#3860, and this parent does not cherry-pick or reimplement it.

## Parent orchestration

- State ledger: `.planning/phases/issue-3855-polling-apply-foundations-parent/`
- Current canonical state: `parent_draft_pr`
- Verified draft: #4041 is open and draft with base `feat/3862-any-to-any-transport` and head
  `feat/3855-polling-apply-foundations`; post-refresh validation will re-inspect its final head
  SHA and ancestry from `c67f40a5...`.
- Next state: `map_wave_phase`; child implementation remains pending.
- Worker mode: single-worker inline/manual GSD fallback; no roles were spawned.
- Integration/merge: prohibited in this scaffold; final parent integration remains human-gated.

## Verification

- GSD adapter and canonical delivery projection passed before the seed.
- This diff is planning-only; no `cmd/` or `internal/` path is changed.
- The terminal document commit from no-mistakes run `01KZPY9EBYX84WZM11EN1F6C83` was retained in
  ancestry and whole-reverted; the accepted tree and protected #4015 architecture blob were
  restored exactly.
- The accepted transport refresh replays eight patch-equivalent planning-only commits without a
  product change. The causal pre-bridge no-mistakes start was rejected by its internal
  non-fast-forward validation ref before a new run or external effect; audited `-s ours` bridge
  `72a4fc32...` passed its tree, parent, ancestry, and protected-blob assertions.
- No force push, rebase, reset, cherry-pick, ordinary content merge, external push, PR mutation,
  or integration merge is authorized by this recovery checkpoint.

## Automated review

- Primary route: pending — this PR intentionally remains draft.
- Fallback route: none.
- Coverage status: pending; no review is requested while draft.
- Certification/executable behavior: not claimed.
