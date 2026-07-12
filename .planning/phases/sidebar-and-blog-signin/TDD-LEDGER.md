# TDD Ledger

Phase: sidebar-and-blog-signin

Record failing test evidence before production code for every behavior-adding task.

## Red: Navbar Hydration And Blog Sidebar Auth

- Command:
  `DATABASE_URL=postgres://website:dev@localhost:55432/website PLAYWRIGHT_BASE_URL=http://localhost:3100 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node ./node_modules/@playwright/test/cli.js test tests/e2e/blog-smoke.spec.ts tests/e2e/docs-smoke.spec.ts --project=chromium`
- Result: failed as expected.
- Evidence: 4 failures waiting for `header[data-navbar-hydrated="true"]`, which is the new hydration marker not yet implemented. Remaining unrelated smoke tests passed.
- Note: `npx -y pnpm@11.7.0 ...` and bundled `pnpm` could not be used for this red run because the local npm/pnpm launcher executed under an older Node runtime; direct Playwright CLI was run with Node 24.13.1.

## Red: Right Rail On Every Desktop Page Shell

- Command:
  `DATABASE_URL=postgres://website:dev@localhost:55432/website PLAYWRIGHT_BASE_URL=http://localhost:3100 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node ./node_modules/@playwright/test/cli.js test tests/e2e/blog-smoke.spec.ts --project=chromium --grep "mounts both sidebars"`
- Result: failed as expected.
- Evidence: `/blog` exposed `.home-sidebar-panel`, but `.page-aside-panel[data-site-toc]` was not found.

## Green: Right Rail On Every Desktop Page Shell

- Command:
  `DATABASE_URL=postgres://website:dev@localhost:55432/website PLAYWRIGHT_BASE_URL=http://localhost:3100 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node ./node_modules/@playwright/test/cli.js test tests/e2e/blog-smoke.spec.ts --project=chromium --grep "mounts both sidebars"`
- Result: passed.
- Evidence: desktop shells for `/`, `/blog`, `/blog/:slug`, `/bookmarks`, `/changelog`, `/patterns`, and `/docs` expose their expected left and right sidebar panels.

## Red: Current Discussion Card In Right Rail

- Command:
  `DATABASE_URL=postgres://website:dev@localhost:55432/website PLAYWRIGHT_BASE_URL=http://localhost:3100 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node ./node_modules/@playwright/test/cli.js test tests/e2e/blog-smoke.spec.ts --project=chromium --grep "moves current discussion"`
- Result: failed as expected.
- Evidence: `.home-sidebar-panel [data-blog-auth-card]` still resolved to 1 element on `/blog/one-cli-to-rule-them-all`; expected 0 after moving the card to the right rail.

## Green: Current Discussion Card In Right Rail

- Command:
  `DATABASE_URL=postgres://website:dev@localhost:55432/website PLAYWRIGHT_BASE_URL=http://localhost:3100 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node ./node_modules/@playwright/test/cli.js test tests/e2e/blog-smoke.spec.ts --project=chromium --grep "moves current discussion"`
- Result: passed.
- Evidence: left sidebar no longer contains `[data-blog-auth-card]`; right rail contains `[data-github-discussion-card]`; card and footer link both target the current page's GitHub Discussions search URL.

## Red: Blog Sign-In Above GitHub Discussion

- Command:
  `DATABASE_URL=postgres://website:dev@localhost:55432/website PLAYWRIGHT_BASE_URL=http://localhost:3100 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node ./node_modules/@playwright/test/cli.js test tests/e2e/blog-smoke.spec.ts --project=chromium --grep "moves blog sign-in"`
- Result: failed as expected.
- Evidence: `.page-aside-panel [data-blog-auth-card]` was visible, but did not contain `Join blog discussion`; the card still rendered the GitHub discussion CTA instead of the blog sign-in CTA.

## Green: Blog Sign-In Above GitHub Discussion

- Command:
  `DATABASE_URL=postgres://website:dev@localhost:55432/website PLAYWRIGHT_BASE_URL=http://localhost:3100 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node ./node_modules/@playwright/test/cli.js test tests/e2e/blog-smoke.spec.ts --project=chromium --grep "moves blog sign-in"`
- Result: passed.
- Evidence: `.page-aside-panel [data-blog-auth-card]` now contains `Join blog discussion` and a `Sign in` button, while `.page-aside-panel [data-github-discussion-link]` points to the current page's GitHub Discussions URL.
