# TDD LEDGER: Reader Notes And Header Clarity

## Red: Collision-Free Header

- Command: `PLAYWRIGHT_BASE_URL=http://127.0.0.1:3100 node .../playwright test tests/e2e/blog-smoke.spec.ts --grep "keeps desktop navigation"`
- Result: failed before production edits because `[data-navbar-links]` did not exist.
- Geometry captured before the edit at 1152px: navigation links ended at x=720.61 while search began at x=632.56, an 88px overlap.

## Red: Recursive Thread Builder

- Command: `node ./node_modules/vitest/vitest.mjs run tests/comment-tree.test.ts`
- Result: failed before production edits because `@/lib/comments/comment-tree` did not exist.
- Expected behavior: a reply to `reply-one` must appear in `reply-one.children`, not beside it under the root.

## Red: Pinned Note And Nested Browser Flow

- Existing database-backed annotation tests were updated before production edits to require `[data-note-preview="pinned"]` after pointer movement and `data-thread-depth="2"` for a reply-to-reply.
- The normal OAuth server correctly rejected the test-only credentials route, so this run could not reach the UI assertions without restarting in test-auth mode.
- A database-free mocked browser spec was added after implementation to exercise the same assertions without deleting or mutating local comments.

## Green

- `node ./node_modules/vitest/vitest.mjs run tests/comment-tree.test.ts`: 2 passed.
- `PLAYWRIGHT_BASE_URL=http://localhost:3100 node .../playwright test tests/e2e/blog-comments-ui.spec.ts`: 2 passed.
- `PLAYWRIGHT_BASE_URL=http://localhost:3100 node .../playwright test tests/e2e/blog-smoke.spec.ts`: 6 passed.
- `PLAYWRIGHT_BASE_URL=http://localhost:3100 node .../playwright test tests/e2e/docs-smoke.spec.ts`: 6 passed.
