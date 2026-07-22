# PR 346 Review Fixes

## Objective

Resolve correctness findings found during current-head review of parent PR #346,
then verify the assembled parent with stacked PR #394 before any GitHub merge.

## Scope

1. Prevent bookmarks for body and point blocks with matching text/offsets from
   colliding in storage.
2. Restore optimistic comment/bookmark state when a DELETE request rejects at
   the network layer.
3. Run targeted red/green tests, the website suite, a production build, and a
   local production-style deployment on port 3100.

## Delivery constraints

- Work only in the isolated `codex/review-pr-346` worktree.
- Do not print or copy OAuth/database secret values.
- Do not merge the parent to `main` or trigger production deployment during
  local verification.
- The repo-local `scripts/gsd prompt programming-loop` command is unavailable;
  this phase uses the documented inline/manual GSD fallback.

## Required skills used

- `gsd-code-review`
- `gsd-programming-loop` (manual fallback)
- `vercel-react-best-practices`
- `vercel-composition-patterns`

`frontend-design` and `web-design-guidelines` are required by repo policy but
are not installed in this checkout; existing website tokens and component
patterns remain the design fallback.
