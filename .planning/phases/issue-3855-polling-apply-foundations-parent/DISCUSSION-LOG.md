# Discussion log — #3855 parent scaffold

## Inline manual-GSD fallback

The canonical contract forbids spawning a planner, reviewer, verifier, or GSD role for this parent
job. The project-local adapter was healthy, all required command sources resolved, and generated
prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`
were inspected. The workflow is therefore executed inline and recorded in this phase.

## Inputs treated as decisions

1. The parent is #3855, with existing children #3856, #3857, #3858, #3859, and #3860; no issue is
   created.
2. The parent starts on `origin/feat/3862-any-to-any-transport`, opens as a draft against that same
   branch, and later retargets only to `docs/4015-connector-release-certification`.
3. The logical issue graph permits #3858 and #3859 after #3857; the active programme's stricter
   execution schedule is #3856, then #3857, then #3858, then #3859. #3860 is a follow-on
   documentation child that depends on all four core children.
4. PR #3880 / `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` is recorded as partial reuse, not child
   completion.
5. The sole deliverable is a reviewable parent seed and acceptance ledger; implementation and
   certification assertions are prohibited.

## Deferred decisions

- Public polling-versus-CDC capability taxonomy is a #3857 prerequisite, not a decision for this
  parent scaffold.
- Any #3864 review outcome, child implementation, PostgreSQL behavior, and final #4015 integration
  remain outside this phase.

## Resolved external transition and scope recovery

`gh-axi` REST verifies draft PR #4041 as open with head
`feat/3855-polling-apply-foundations` and temporary dependency-only base
`feat/3862-any-to-any-transport`.

The completed local no-mistakes run retained its terminal document commit
`81f37c2b2cf638bfd6b06b35393256232ea6e23d` in ancestry, then whole-revert
`4099335632a79e3c70ce20c004a5f63933171280` restored the accepted
`b2d0f63029a34f8d647e57f34a747ad3983e1578` tree. The out-of-scope inherited #4015 architecture
file remains protected at blob `889e4eddffabb76aa8be46a934c3e9abe0610f4c`.

This is custody/topology recovery, not a correction round or product implementation. Fresh
no-mistakes validation is required for the post-evidence SHA.

## Accepted transport-seam recovery decision

The fixed temporary base name did not change, but its accepted head did: transport PR #4059 merged
into #4019 at `c67f40a5ff67a131950f3123e70527027dca8493`. The previous #3855 history was based on
`30b2fb4aeb121641b6158903fe1d3b54668599a6`, and `c67f40a5...` was not its ancestor. The required
answer is therefore a safe rebase of the existing parent history, not a new branch, PR, issue, or
product change. All eight existing commits replayed cleanly and range-diff as patch-equivalent.

The only fresh decision is mechanical and already fixed by the recovery brief: future #3856 must
inherit the accepted transport seam. The draft remains on the named temporary branch relationship;
it does not authorize a merge into #3862 or `main` and does not advance the canonical state beyond
`parent_draft_pr`.

## Audited no-mistakes gate recovery decision

The direct fresh no-mistakes start at `e541170eef71f687990684b4f22fa7b21ab5a9fc` is recorded as
causal RED, not a pipeline outcome: its internal validation-gate branch remained at preserved
`b61d0fa7eefc719c39593e44afbcb1b7a3f76613` and rejected the refreshed branch as non-fast-forward
before a new run, external push, PR update, or CI action occurred.

The bounded recovery is already decided by the reconciliation audit. After checkpoint
`3b6e3e2e...`, every fixed SHA and the protected blob remained exact, and the specified
`--no-ff -s ours` merge of the preserved recovery ref produced `72a4fc32...`. Its content tree is
unchanged, its parents are checkpoint then `b61d0fa7...`, it carries all required ancestry, and its
remote comparison is `0 13`; its sole purpose is to carry both valid ancestries so the validation
gate can accept a normal fast-forward relation. This is neither an ordinary merge nor authorization
to merge #4041, #4019, or `main`. A fresh local pipeline remains the next gate.
