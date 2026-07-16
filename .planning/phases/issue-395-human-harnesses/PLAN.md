# PLAN: Issue 395 - Humans Need Harnesses Too

## Objective

Publish a repository-grounded blog essay that connects Polymetrics runtime mutation controls with
the issue, test, review, release, and deployment harnesses used by this repository.

## Delivery Context

- Issue: `#395`
- Parent issue: `#293`
- Parent PR: `#346`
- Base branch: `feat/293-blog-annotations`
- Work branch: `docs/395-human-harnesses`
- PR base: `feat/293-blog-annotations`

## Required Skills Used

- `ui-refactoring-docs`: preserve the existing information architecture and write for scanning.
- `vercel-react-best-practices`: keep the static catalog model and avoid new client-side work.
- `vercel-composition-patterns`: use the existing shared article renderer and data contract.
- `golang-documentation`: make only evidence-backed CLI and architecture claims.
- `golang-cli`: describe stdout, stderr, exits, and mutation boundaries as contracts.
- `golang-security`: avoid secrets and distinguish policy from enforced controls.
- `frontend-design` and `web-design-guidelines`: not installed in the current skill set; the
  repository design system, `ui-refactoring-docs`, and direct browser verification are the recorded
  fallback.

## GSD Activation

- `scripts/gsd doctor`: pass.
- `scripts/gsd prompt programming-loop init --phase issue-human-harnesses --dry-run`: failed because
  the repo-local registry does not expose `programming-loop`.
- Fallback: manual GSD loop using this plan, `TDD-LEDGER.md`, and `VERIFICATION.md`.

## Verified Source Set

- Runtime rules and connector inventory under `internal/connectors/defs`.
- Agent delivery contracts and Shepherd validator under `.agents/agentic-delivery`.
- GitHub workflows under `.github/workflows`.
- Current `main` branch-protection status queried through GitHub.
- Existing blog schema and article renderer under `website/lib/blog.ts` and `website/components/blog`.

## Tasks

1. Add a focused failing catalog test for the `human-harnesses` slug and key editorial contract.
2. Add the article to `BLOG_POSTS` without changing the renderer or route architecture.
3. Replace stale inventory figures with reproducible counts from the current parent branch.
4. Explain the repository harness as intent, evidence, deterministic checks, adversarial review,
   release, and production deployment layers.
5. State current gaps honestly: process-level human merge gate, optional review-bot availability,
   and production-environment approval not being equivalent to a required GitHub review count.
6. Run targeted tests, full website checks, build, and desktop/mobile browser verification.
7. Commit and push coherent plan, red, implementation, and review-fix checkpoints.
8. Open a stacked PR with `Refs #395` and `Refs #293`; do not merge the parent PR into `main`.

## Scope Boundaries

- Expected production edit: `website/lib/blog.ts` only.
- Expected test edit: one focused blog catalog unit test.
- Planning artifacts live only in `.planning/phases/issue-395-human-harnesses`.
- No workflow, runtime, auth, comment, database, generated website data, or dependency changes.

## Risks

- Future-dated content is immediately addressable because the catalog has no publication scheduler.
- Workflow prose can overstate checks that run but are not required by branch protection.
- Connector inventory numbers can drift as the parent branch advances.
- A long essay can become hard to scan in the existing article layout.

## Commit Checkpoints

1. Plan and manual-GSD evidence.
2. Red catalog test.
3. Green article implementation and updated verification evidence.
4. Review fixes, if any.

## Requested Narrative Revision

- Reframe the essay around the oversized, roughly million-line pull request that made review,
  conflict resolution, and integration confidence collapse into one merge event.
- Explain how that failure led to the parent issue, bounded sub-issue, isolated worktree, collision
  rule, stacked PR, review-coverage, and human-gated parent architecture.
- Keep the technical workflow evidence, but make the prose sound like a person telling the story
  rather than a repository audit report.
- Remove the closing repository, documentation, and dated inventory bullet list.
- End with a natural request to star the repository.
- Preserve the verified inventory numbers in the technical body where they provide necessary scale.

## Publication-Date And Deployment Follow-Up

- Change only the `human-harnesses` `publishedAt` and `updatedAt` values from the future placeholder
  to `2026-07-16`, the current date for this delivery.
- Extend the focused catalog contract first so any unchanged future date fails before the production
  metadata edit.
- Re-run the focused catalog test, typecheck, production build, and browser assertion that the
  article header renders `2026-07-16`.
- Document the existing deployment path rather than changing it: merge the stacked sub-PR into the
  parent branch, complete parent verification and human review, merge the parent PR into `main`,
  and let Website CI/CD publish and deploy the immutable GHCR image.
- No workflow, runner, VPS, OAuth, database, environment, or deployment mutation is part of this
  follow-up.

## Local Review Loop Follow-Up

- Replace the stale remote Claude-workflow section with the current local exact-head review gate.
- Describe the real local sequence observed in the active Pi sessions: verification, independent
  read-only review, disposition, isolated repair, focused verification, and re-review until clean.
- Explain that moved heads invalidate review evidence, accepted findings are fixed by a separate
  bounded worker, and the correction loop stops for a human when its cap is reached.
- Keep optional remote PR-bot review as supplemental shadow/canary input rather than the default
  delivery gate.
- Remove the detailed Shepherd implementation description. Mention Shepherd only as the next
  engineering story about supervising the loop itself.
- Update the focused catalog contract before editing the article, then run the focused test,
  typecheck, production build, and local browser assertions.
- Live evidence consulted read-only: the active Shepherd proof/recovery and CLI architecture Pi
  sessions, their worktree state, and the local-review contract/workflow artifacts. Warp UI access
  was blocked by the desktop safety layer; no live session was interrupted or mutated.
