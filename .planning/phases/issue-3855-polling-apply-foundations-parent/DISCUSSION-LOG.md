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
3. Core delivery order is #3856, then #3857, then #3858 and #3859 in parallel. #3860 is a
   follow-on documentation child that depends on all four core children.
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
