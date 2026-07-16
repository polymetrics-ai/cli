# TDD Ledger

Phase: `issue-395-human-harnesses`

## Baseline

- The parent catalog contains three posts and no `human-harnesses` slug.
- A live route baseline could not be captured because no process was listening on port 3100.
- The red test will therefore exercise the catalog directly and fail before the production edit.

## Red: Human Harnesses Catalog Contract

- Status: complete.
- Test: resolve `human-harnesses` through `getBlogPost` and require the verified inventory figures,
  delivery-harness section, current-limitations section, and August 4 publication metadata.
- Command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`
- Result: failed as expected because `getBlogPost('human-harnesses')` returned `undefined`.

## Green: Human Harnesses Catalog Contract

- Status: complete.
- Command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`
- Result: passed after adding the post to the existing `BLOG_POSTS` catalog.
- Scope: no renderer, route, component, dependency, auth, or database changes were needed.

## Refactor

- Status: complete.
- Replaced stale operation counts with a fresh API-surface inventory.
- Corrected the draft's Shepherd description: connector workers perform surface mapping; the
  Shepherd-style layer independently validates orchestration decisions.
- Distinguished workflows that run from status checks currently required by `main` branch
  protection.
- Reframed August 4 behavior as a release target and separated inventory, implementation,
  conformance, and live certification.
- Kept the existing renderer and static catalog architecture unchanged.
