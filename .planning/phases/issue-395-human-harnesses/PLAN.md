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

