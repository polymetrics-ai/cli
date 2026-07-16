# Summary: Issue 395 - Humans Need Harnesses Too

## Delivered

- Added `Humans Need Harnesses Too` as the featured, data-driven blog post.
- Expanded the supplied draft into 14 sections covering runtime controls, issue-first delivery,
  GSD/TDD evidence, verification, security, static AI review, releases, deployment, branch
  protection, and current hardening gaps.
- Corrected stale inventory figures and separated connector inventory from implementation and live
  certification.
- Corrected the Shepherd description so it matches the repository's supervisor/validator role.
- Added a focused catalog contract test without changing the article renderer or route system.
- Reframed the article around the roughly million-line pull request that inspired the harness
  architecture, including parent issues, bounded sub-issues, isolated worktrees, stacked PRs,
  layered checks, and the final human merge gate.
- Removed the repository, documentation, and inventory-snapshot footer and replaced it with a
  conversational request to star the repository.
- Changed only the `human-harnesses` publication and update dates to `2026-07-16` so the article
  header reflects the current delivery date.

## Evidence

- Inventory: 547 connectors and 29,129 operations: 14,780 GET reads, 3 HEAD checks, 14,169
  explicit HTTP mutations, 177 mixed/nonstandard rows, 2,903 deletes, and 7,088 streams.
- Red: catalog test failed because `human-harnesses` was absent.
- Green: focused catalog test passed after the post was added.
- Typecheck: pass.
- Unit tests: 64 passed.
- Blog Playwright smoke: 6 passed.
- Production build: pass; the new slug was statically generated.
- Website generation: pass with no drift.
- Desktop/mobile screenshots and overflow assertions: pass.
- Narrative contract: the giant-PR origin story, isolated-worktree architecture, 14-minute reading
  time, and star request are present; the removed footer strings are absent.
- Date contract: focused test, typecheck, production build, and a local Chromium render assertion
  pass with `2026-07-16`; the old `2026-08-04` placeholder is absent.

## Scope

- Production: `website/lib/blog.ts` only.
- Test: `website/tests/blog-catalog.test.ts` only.
- No dependencies, workflows, auth, database, comments, runtime, connectors, or generated artifacts
  changed.

## Delivery Notes

- Manual GSD fallback was used because the repo-local command registry does not expose
  `programming-loop`.
- Required skills: `ui-refactoring-docs`, `vercel-react-best-practices`,
  `vercel-composition-patterns`, `golang-documentation`, `golang-cli`, and `golang-security`.
- `frontend-design` and `web-design-guidelines` were unavailable; the existing design system and
  direct browser verification were used as the recorded fallback.
- Parent issue `#293` and parent PR `#346` remain human-gated.

## Follow-Up Risks

- Future-dated catalog entries are not hidden before publication.
- Shared mobile note controls can temporarily overlay the viewport's lowest line.
- The parent branch can log a bookmarks auth-card hydration mismatch during e2e despite passing.
