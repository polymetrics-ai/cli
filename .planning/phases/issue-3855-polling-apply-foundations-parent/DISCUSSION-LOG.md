# Discussion log — #3855 parent scaffold

## Inline manual-GSD fallback

The canonical contract forbids spawning a planner, reviewer, verifier, or GSD role for this parent
job. The project-local adapter was healthy, all required command sources resolved, and generated
prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`
were inspected. The workflow is therefore executed inline and recorded in this phase.

## Inputs treated as decisions

1. The parent is #3855, with existing children #3856, #3857, #3858, and #3859; no issue is created.
2. The parent starts on `origin/feat/3862-any-to-any-transport`, opens as a draft against that same
   branch, and later retargets only to `docs/4015-connector-release-certification`.
3. Delivery order is #3856, then #3857, then #3858 and #3859 in parallel.
4. PR #3880 / `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` is recorded as partial reuse, not child
   completion.
5. The sole deliverable is a reviewable parent seed and acceptance ledger; implementation and
   certification assertions are prohibited.

## Deferred decisions

- Public polling-versus-CDC capability taxonomy is a #3857 prerequisite, not a decision for this
  parent scaffold.
- Any #3864 review outcome, child implementation, PostgreSQL behavior, and final #4015 integration
  remain outside this phase.

## External-state exception

GitHub API reads through `gh-axi` were rate-limited. That is recorded as pending verification rather
than converted into a topology assumption or a correction round.
