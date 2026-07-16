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

## Review Fix: Method Classification

- Parent-orchestrator audit identified that treating every non-GET/HEAD operation as a mutation
  mislabeled hook, wildcard, GraphQL, WebSocket, and composite-method rows.
- Updated the article and catalog contract to report 14,780 GET reads, 3 HEAD checks, 14,169
  explicit HTTP mutations, and 177 mixed/nonstandard rows.
- Tightened the article's PR Issue Guard and GSD workflow descriptions to their actual enforcement
  boundaries.
- Replaced an unsupported reference to signed release artifacts with the versioned GoReleaser
  artifacts the repository actually produces.
- Preserved the supplied draft's connection between the article annotation UI and a durable human
  feedback loop.
- Replaced the draft-style standalone approval command with the documented reverse ETL contract:
  human-readable plan output supplies the token consumed by `pm reverse run --approve`.
- Verification: catalog test, typecheck, six blog Playwright tests, production build, and
  `git diff --check` passed after the review fix.
