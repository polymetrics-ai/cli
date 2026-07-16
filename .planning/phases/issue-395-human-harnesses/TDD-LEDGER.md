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

- Status: pending.
- Expected result: pass after the smallest `BLOG_POSTS` entry is added.

## Refactor

- Status: pending.
- Editorial pass will remove unsupported claims, duplicated transitions, and unnecessary renderer
  changes while preserving the supplied first-person voice.
